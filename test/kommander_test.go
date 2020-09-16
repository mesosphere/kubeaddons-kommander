package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/blang/semver"
	"github.com/mesosphere/ksphere-testing-framework/pkg/cluster/kind"
	"github.com/mesosphere/ksphere-testing-framework/pkg/experimental"
	testharness "github.com/mesosphere/ksphere-testing-framework/pkg/harness"
	"github.com/mesosphere/kubeaddons/pkg/constants"
	addontesters "github.com/mesosphere/kubeaddons/test/utils"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

const comRepoRef = "master"

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

	addonDeployment, err := addontesters.DeployAddons(t, cluster, addons...)
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
