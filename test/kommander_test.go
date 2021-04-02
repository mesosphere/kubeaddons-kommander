package test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/blang/semver"
	"github.com/mesosphere/ksphere-testing-framework/pkg/cluster/kind"
	testharness "github.com/mesosphere/ksphere-testing-framework/pkg/harness"
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta2"
	"github.com/mesosphere/kubeaddons/pkg/constants"
	testutils "github.com/mesosphere/kubeaddons/test/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

const comRepoRef = "master"

// autoProvisioningChartPath is a path to the `auto-provisioning` chart
// that can be used to install `auto-provisioning` from `konvoy`.
// Makefile in this project comes with a target that can prepare the chart.
// @see: auto-provisioning.prepare-chart
const autoProvisioningChartPath = "../build/chart/auto-provisioning"
const autoProvisioningNamespace = "konvoy"
const autoProvisioningName = "auto-provisioning"
const initialE2EKindConfigPath = "artifacts/kind-config.yaml"

const dockerUsernameEnv = "DOCKERHUB_ROBOT_USERNAME"
const dockerPasswordEnv = "DOCKERHUB_ROBOT_TOKEN"


const tempDir = "/tmp/kubeaddons-kommander"
const kubeaddonsControllerNamespace = "kubeaddons"
const kubeaddonsControllerPodPrefix = "kubeaddons-controller-manager-"

func TestKommanderGroup(t *testing.T) {
	t.Logf("testing group kommander")

	createOption := cluster.CreateWithV1Alpha4Config(&v1alpha4.Cluster{})
	// If we are in CI, we set the ImageRegistries to use the Docker Hub credentials.
	if os.Getenv("CI") != "" && os.Getenv(dockerUsernameEnv) != "" && os.Getenv(dockerPasswordEnv) != "" {
		initialKindConfig, err := ioutil.ReadFile(initialE2EKindConfigPath)
		if err != nil {
			t.Fatal(err)
		}
		kindConfig := strings.Replace(string(initialKindConfig), "DOCKER_USERNAME", os.Getenv(dockerUsernameEnv), 1)
		kindConfig = strings.Replace(kindConfig, "DOCKER_PASSWORD", os.Getenv(dockerPasswordEnv), 1)
		createOption = cluster.CreateWithRawConfig([]byte(kindConfig))
	}

	cluster, err := kind.NewClusterWithVersion(semver.MustParse(strings.TrimPrefix(defaultKubernetesVersion, "v")), createOption)
	if err != nil {
		// try to clean up in case cluster was created and reference available
		if cluster != nil {
			_ = cluster.Cleanup()
		}
		t.Fatal(err)
	}
	// This informs ksphere-testing-framework to gather the logs on cleanup.
	os.Setenv("KIND_CLUSTER_LOGS_PATH", "/tmp/kind/")
	defer cluster.Cleanup()

	if err := kubectl("apply", "-f", controllerBundle); err != nil {
		t.Fatal(err)
	}

	addons := groups["kommander"]
	for _, addon := range addons {
		overrides(addon)
	}

	wg := &sync.WaitGroup{}
	stop := make(chan struct{})
	go testutils.LoggingHook(t, cluster, wg, stop)

	var certManagerAddon v1beta2.AddonInterface
	addonsWithoutCertManager := []v1beta2.AddonInterface{}
	for _, addon := range addons {
		if addon.GetName() == "cert-manager" {
			certManagerAddon = addon
		} else if addon.GetName() != "kommander" {
			addonsWithoutCertManager = append(addonsWithoutCertManager, addon)
		}
	}

	if certManagerAddon == nil {
		t.Fatal("failed to find cert manager addon")
	}
	installCertManagerAddon, err := testutils.DeployAddons(t, cluster, certManagerAddon)
	if err != nil {
		t.Fatal(err)
	}
	waitForCertManager, err := testutils.WaitForAddons(t, cluster, certManagerAddon)
	if err != nil {
		t.Fatal(err)
	}

	installAutoProvisioningJobs := testharness.Jobs{
		installCertManagerAddon.Jobs[0],
		waitForCertManager.Jobs[0],
		func(t *testing.T) error {
			t.Log("installing auto-provisioning")
			kubeConfig := genericclioptions.NewConfigFlags(false)
			kubeConfig.APIServer = &cluster.Config().Host
			kubeConfig.BearerToken = &cluster.Config().BearerToken
			kubeConfig.CAFile = &cluster.Config().CAFile

			cfg := &action.Configuration{}
			err := cfg.Init(
				kubeConfig,
				autoProvisioningNamespace,
				"memory",
				t.Logf,
			)
			if err != nil {
				return err
			}

			chart, err := loader.LoadDir(autoProvisioningChartPath)
			if err != nil {
				return err
			}

			values, err := chartutil.ReadValues([]byte(`
certManagerCertificates:
  issuer:
    name: test
    selfSigned: true

konvoy:
  allowUnofficialReleases: true

kubeaddonsRepository:
  validTagPrefixes:
  - testing
`))
			if err != nil {
				return err
			}

			installAction := action.NewInstall(cfg)
			installAction.ReleaseName = autoProvisioningName
			installAction.Namespace = autoProvisioningNamespace
			installAction.CreateNamespace = true
			installAction.Wait = true
			_, err = installAction.Run(chart, values.AsMap())
			if err != nil {
				return fmt.Errorf("failed to install auto-provisioning chart: %w", err)
			}
			return nil
		},
	}
	installAutoProvisioning := testharness.Loadable{
		Plan: testharness.DeployPlan,
		Jobs: installAutoProvisioningJobs,
	}

	// there is a bug of kommander not being able to be uninstalled
	// https://jira.d2iq.com/browse/D2IQ-73310
	// remove it from the addon cleanup list
	addonsToCleanup := make([]v1beta2.AddonInterface, 0)
	for _, addon := range addons {
		if addon.GetName() != "kommander" {
			addonsToCleanup = append(addonsToCleanup, addon)
		}
	}
	addonCleanup, err := testutils.CleanupAddons(t, cluster, addonsToCleanup...)
	if err != nil {
		t.Fatal(err)
	}

	addonDefaults, err := testutils.WaitForAddons(t, cluster, addons...)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	addonUpgrades := testharness.Loadables{}
	for _, addon := range addons {
		if addon.GetName() != "kommander" {
			continue
		}
		found = true

		t.Logf("determining old and new versions of Kommander for upgrade testing")
		oldAddon, err := testutils.GetLatestAddonRevisionFromLocalRepoBranch("../", "origin", comRepoRef, "kommander")
		if err != nil {
			t.Fatal(err)
		}
		oldVersion, err := semver.Parse(strings.TrimPrefix(oldAddon.GetAnnotations()[constants.AddonRevisionAnnotation], "v"))
		if err != nil {
			t.Fatal(err)
		}
		newVersion, err := semver.Parse(strings.TrimPrefix(addon.GetAnnotations()[constants.AddonRevisionAnnotation], "v"))
		if err != nil {
			t.Fatal(err)
		}
		if oldVersion.GT(newVersion) {
			t.Fatal(fmt.Errorf("new kommander version is %s which is below the previous version %s", newVersion, oldVersion))
		}
		if oldVersion.EQ(newVersion) {
			t.Logf("Kommander itself was not updated, ignoring upgrade tests")
			addonsWithoutCertManager = append(addonsWithoutCertManager, addon)
			break
		}

		t.Logf("an upgrade test for Kommander from previous version %s to new version %s is needed", oldVersion, newVersion)
		addonUpgrade, err := testutils.UpgradeAddon(t, cluster, oldAddon, addon)
		if err != nil {
			t.Fatal(err)
		}

		addonUpgrades = append(addonUpgrades, addonUpgrade)
	}

	addonDeployment, err := testutils.DeployAddons(t, cluster, addonsWithoutCertManager...)
	if err != nil {
		t.Fatal(err)
	}

	if !found {
		t.Fatal(fmt.Errorf("could not find kommander addon in test group, this shouldn't happen"))
	}

	th := testharness.NewSimpleTestHarness(t)
	th.Load(
		testutils.ValidateAddons(addons...),
		installAutoProvisioning,
	)
	// If upgrade test necessary, install kommander just once in the upgrade test to avoid uninstalling it
	// If upgrade test not necessary, kommander deployed with addonDeployment
	th.Load(addonUpgrades...)
	th.Load(
		addonDeployment,
		addonDefaults,
		addonCleanup,
		testharness.Loadable{
			Plan: testharness.DefaultPlan,
			Jobs: testharness.Jobs{
				thanosChecker(t, cluster),
				karmaChecker(t, cluster),
				kubecostChecker(t, cluster),
			},
		})

	// Collect kubeaddons controller logs during cleanup.
	th.Load(testharness.Loadable{
		Plan: testharness.CleanupPlan,
		Jobs: testharness.Jobs{func(t *testing.T) error {
			if err := os.MkdirAll(tempDir, 0755); err != nil {
				t.Fatal(err)
			}
			dir, err := ioutil.TempDir(tempDir, "kommander-")
			if err != nil {
				t.Fatal(err)
			}

			logFilePath := filepath.Join(dir, "kubeaddons-controller-log.txt")
			t.Logf("INFO: writing kubeaddons controller logs to %s", logFilePath)

			logFile, err := os.Create(logFilePath)
			if err != nil {
				return err
			}
			defer logFile.Close()

			logs, err := logsFromPodWithPrefix(cluster, kubeaddonsControllerNamespace, kubeaddonsControllerPodPrefix)
			if err != nil {
				return err
			}
			defer logs.Close()

			_, err = io.Copy(logFile, logs)
			return err
		}},
	})

	defer th.Cleanup()
	th.Validate()
	th.Deploy()
	th.Default()

	close(stop)
	wg.Wait()

	t.Log("kommander test group complete")
}
