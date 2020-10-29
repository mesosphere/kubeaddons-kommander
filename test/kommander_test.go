package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/blang/semver"
	frwcluster "github.com/mesosphere/ksphere-testing-framework/pkg/cluster"
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

var autoProvisioningChartValues = []byte(`
certManagerCertificates:
  issuer:
    name: test
    selfSigned: true

konvoy:
  allowUnofficialReleases: true

kubeaddonsRepository:
  validTagPrefixes:
  - testing
`)

func TestKommanderGroup(t *testing.T) {
	t.Logf("testing group kommander")

	cluster, err := kind.NewClusterWithVersion(semver.MustParse(strings.TrimPrefix(defaultKubernetesVersion, "v")), cluster.CreateWithV1Alpha4Config(&v1alpha4.Cluster{}))
	if err != nil {
		// try to clean up in case cluster was created and reference available
		if cluster != nil {
			_ = cluster.Cleanup()
		}
		t.Fatal(err)
	}
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

	// The `groups.yaml` file defines `cert-manager` as an addon that needs to be
	// installed. We need to install it separately with `auto-provisioning`
	// and the install the rest of the addons.
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

	// `auto-provisioning` is now a dependency of the `kommander` and it is not installed
	// as an addon but `konvoy` installs it using helm. Since the base cluster is
	// a `kind` the `auto-provisioning` has to be installed manually. The
	// `auto-provisioning` depends on `cert-manager` that must be installed before
	// installing `auto-provisioning`. These jobs install `cert-manager` and wait
	// for it to complete installation and then use `helm` to install `auto-provisioning`.
	// Once these are in place other addons can be installed.
	installAutoProvisioning := testharness.Loadable{
		Plan: testharness.DeployPlan,
		Jobs: testharness.Jobs{
			installCertManagerAddon.Jobs[0],
			waitForCertManager.Jobs[0],
			installAutoProvisioningJob(t, cluster),
		},
	}

	addonDeployment, err := addontesters.DeployAddons(t, cluster, addonsWithoutCertManager...)
	if err != nil {
		t.Fatal(err)
	}

	addonCleanup, err := addontesters.CleanupAddons(t, cluster, addons...)
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

// installAutoProvisioningJob creates `harness.Job` that installs `auto-provisioner`
// on the given cluster using `helmv3`.
func installAutoProvisioningJob(t *testing.T, cluster frwcluster.Cluster) func(t *testing.T) error {
	return func(t *testing.T) error {
		t.Log("installing auto-provisioning")

		chart, err := loader.LoadDir(autoProvisioningChartPath)
		if err != nil {
			return err
		}
		values, err := chartutil.ReadValues(autoProvisioningChartValues)
		if err != nil {
			return err
		}

		kubeConfig := genericclioptions.NewConfigFlags(false)
		kubeConfig.APIServer = &cluster.Config().Host
		kubeConfig.BearerToken = &cluster.Config().BearerToken
		kubeConfig.CAFile = &cluster.Config().CAFile

		cfg := &action.Configuration{}
		if err := cfg.Init(kubeConfig, autoProvisioningNamespace, "memory", t.Logf); err != nil {
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
	}
}
