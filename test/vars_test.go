package test

import (
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta2"
	"github.com/mesosphere/kubeaddons/pkg/catalog"
	"github.com/mesosphere/kubeaddons/pkg/repositories"
)

// -----------------------------------------------------------------------------
// Private - Const & Vars
// -----------------------------------------------------------------------------

const (
	kbaURL    = "https://github.com/mesosphere/kubernetes-base-addons"
	kbaRef    = "release/3.0"
	kbaRemote = "origin"

	defaultKubernetesVersion = "1.18.8"
	controllerBundle         = "https://mesosphere.github.io/kubeaddons/bundle.yaml"
	patchStorageClass        = `{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}`
)

var (
	cat       catalog.Catalog
	localRepo repositories.Repository
	groups    map[string][]v1beta2.AddonInterface
)
