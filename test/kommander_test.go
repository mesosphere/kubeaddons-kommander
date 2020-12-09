package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/blang/semver"
	"github.com/mesosphere/ksphere-testing-framework/pkg/cluster/kind"
	"github.com/mesosphere/ksphere-testing-framework/pkg/experimental"
	testharness "github.com/mesosphere/ksphere-testing-framework/pkg/harness"
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta2"
	"github.com/mesosphere/kubeaddons/pkg/constants"
	addontesters "github.com/mesosphere/kubeaddons/test/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

func TestKommanderGroup(t *testing.T) {
	t.Logf("testing group kommander")

	var rawKindConfig []byte
	// If we are in CI, we set the ImageRegistries to use the Docker Hub credentials.
	if os.Getenv("CI") != "" && os.Getenv(dockerUsernameEnv) != "" && os.Getenv(dockerPasswordEnv) != "" {
		initialKindConfig, err := ioutil.ReadFile(initialE2EKindConfigPath)
		if err != nil {
			t.Fatal(err)
		}
		kindConfig := strings.Replace(string(initialKindConfig), "DOCKER_USERNAME", os.Getenv(dockerUsernameEnv), 1)
		kindConfig = strings.Replace(kindConfig, "DOCKER_PASSWORD", os.Getenv(dockerPasswordEnv), 1)
		rawKindConfig = []byte(kindConfig)
	}

	cluster, err := kind.NewClusterWithVersion(semver.MustParse(strings.TrimPrefix(defaultKubernetesVersion, "v")), cluster.CreateWithRawConfig(rawKindConfig))
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
	go experimental.LoggingHook(t, cluster, wg, stop)

	var certManagerAddon v1beta2.AddonInterface
	addonsWithoutCertManager := []v1beta2.AddonInterface{}
	for _, addon := range addons {
		if addon.GetName() == "cert-manager" {
			certManagerAddon = addon
		} else {
			addonsWithoutCertManager = append(addonsWithoutCertManager, addon)
		}
	}

	if certManagerAddon == nil {
		t.Fatal("failed to find cert manager addon")
	}
	installCertManagerAddon, err := addontesters.DeployAddons(t, cluster, certManagerAddon)
	if err != nil {
		t.Fatal(err)
	}
	waitForCertManager, err := addontesters.WaitForAddons(t, cluster, certManagerAddon)
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

	addonDeployment, err := addontesters.DeployAddons(t, cluster, addonsWithoutCertManager...)
	if err != nil {
		t.Fatal(err)
	}

	// there is well known bug of kommander not being able to be uninstalled
	// https://jira.d2iq.com/browse/D2IQ-63395
	// remove it from the addon cleanup list
	addonsToCleanup := make([]v1beta2.AddonInterface, 0)
	for _, addon := range addons {
		if addon.GetName() != "kommander" {
			addonsToCleanup = append(addonsToCleanup, addon)
		}
	}
	addonCleanup, err := addontesters.CleanupAddons(t, cluster, addonsToCleanup...)
	if err != nil {
		t.Fatal(err)
	}

	addonDefaults, err := addontesters.WaitForAddons(t, cluster, addons...)
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
		oldAddon, err := addontesters.GetLatestAddonRevisionFromLocalRepoBranch("../", comRepoRef, "kommander")
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
			break
		}

		t.Logf("an upgrade test for Kommander from previous version %s to new version %s is needed", oldVersion, newVersion)
		addonUpgrade, err := addontesters.UpgradeAddon(t, cluster, oldAddon, addon)
		if err != nil {
			t.Fatal(err)
		}

		addonUpgrades = append(addonUpgrades, addonUpgrade)
	}

	if !found {
		t.Fatal(fmt.Errorf("could not find kommander addon in test group, this shouldn't happen"))
	}

	th := testharness.NewSimpleTestHarness(t)
	th.Load(
		addontesters.ValidateAddons(addons...),
		installAutoProvisioning,
		addonDeployment,
		addonDefaults,
		addonCleanup)
	th.Load(addonUpgrades...)
	th.Load(testharness.Loadable{
		Plan: testharness.DefaultPlan,
		Jobs: testharness.Jobs{
			thanosChecker(t, cluster),
			karmaChecker(t, cluster),
			kubecostChecker(t, cluster),
		},
	})

	defer th.Cleanup()
	th.Validate()
	th.Deploy()
	th.Default()

	close(stop)
	wg.Wait()

	t.Log("kommander test group complete")
}
