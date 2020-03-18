package test

import (
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
	"github.com/mesosphere/kubeaddons/pkg/catalog"
	"github.com/mesosphere/kubeaddons/pkg/repositories"
)

// -----------------------------------------------------------------------------
// Private - Const & Vars
// -----------------------------------------------------------------------------

const (
	kbaURL    = "https://github.com/mesosphere/kubernetes-base-addons"
	kbaRef    = "master"
	kbaRemote = "origin"

	controllerBundle         = "https://mesosphere.github.io/kubeaddons/bundle.yaml"
	defaultKubernetesVersion = "1.16.4"
	patchStorageClass        = `{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}`
)

var (
	cat       catalog.Catalog
	localRepo repositories.Repository
	groups    map[string][]v1beta1.AddonInterface
)
