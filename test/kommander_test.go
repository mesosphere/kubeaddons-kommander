package test

import (
	"sync"
	"testing"

	"github.com/blang/semver"
	"github.com/mesosphere/ksphere-testing-framework/pkg/cluster/kind"
	"github.com/mesosphere/ksphere-testing-framework/pkg/experimental"
	testharness "github.com/mesosphere/ksphere-testing-framework/pkg/harness"
	addontesters "github.com/mesosphere/kubeaddons/test/utils"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha3"
	"sigs.k8s.io/kind/pkg/cluster"
)

func TestKommanderGroup(t *testing.T) {
	t.Logf("testing group kommander")

	version, err := semver.Parse(defaultKubernetesVersion)
	if err != nil {
		t.Fatal(err)
	}

	cluster, err := kind.NewCluster(version, cluster.CreateWithV1Alpha3Config(&v1alpha3.Cluster{}))
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

	th := testharness.NewSimpleTestHarness(t)
	th.Load(
		addontesters.ValidateAddons(addons...),
		addonDeployment,
		addonDefaults,
		addonCleanup,
		testharness.Loadable{
			Plan: testharness.DefaultPlan,
			Jobs: testharness.Jobs{
				thanosChecker(t, cluster),
				karmaChecker(t, cluster),
			},
		},
	)

	defer th.Cleanup()
	th.Validate()
	th.Deploy()
	th.Default()

	close(stop)
	wg.Wait()

	t.Log("kommander test group complete")
}
