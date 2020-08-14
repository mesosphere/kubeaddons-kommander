package test

import (
	"github.com/mesosphere/ksphere-testing-framework/pkg/experimental"
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta2"
	"github.com/mesosphere/kubeaddons/pkg/catalog"
	"github.com/mesosphere/kubeaddons/pkg/repositories/git"
	"github.com/mesosphere/kubeaddons/pkg/repositories/local"
)

// -----------------------------------------------------------------------------
// Private - Test Setup
// -----------------------------------------------------------------------------

func init() {
	var err error
	localRepo, err = local.NewRepository("local", "../addons/")
	if err != nil {
		panic(err)
	}

	kbaRepo, err := git.NewRemoteRepository(kbaURL, kbaRef, kbaRemote)
	if err != nil {
		panic(err)
	}

	cat, err = catalog.NewCatalog(localRepo, kbaRepo)
	if err != nil {
		panic(err)
	}

	groups, err = experimental.AddonsForGroupsFile("groups.yaml", cat)
	if err != nil {
		panic(err)
	}
}

// -----------------------------------------------------------------------------
// Private - Test Setup - CI Values Overrides
// -----------------------------------------------------------------------------

// TODO: a temporary place to put configuration overrides for addons
// See: https://jira.mesosphere.com/browse/DCOS-62137
func overrides(addon v1beta2.AddonInterface) {
	if v, ok := addonOverrides[addon.GetName()]; ok {
		addon.GetAddonSpec().ChartReference.Values = &v
	}
}

var addonOverrides = map[string]string{
	"metallb": `
---
configInline:
  address-pools:
  - name: default
    protocol: layer2
    addresses:
    - "172.17.1.200-172.17.1.250"
`,
}
