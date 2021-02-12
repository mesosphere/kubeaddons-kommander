module github.com/mesosphere/kubeaddons-kommander-addons/test

go 1.15

// Locked the kind version as we've run into troubles in the past when this gets unintentionally bumped, we want to intentionally manage this version.
replace sigs.k8s.io/kind => sigs.k8s.io/kind v0.9.0

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/mesosphere/ksphere-testing-framework v0.2.6
	github.com/mesosphere/kubeaddons v0.24.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	helm.sh/helm/v3 v3.4.2
	k8s.io/api v0.19.5
	k8s.io/apimachinery v0.19.5
	k8s.io/cli-runtime v0.19.5
	k8s.io/klog/v2 v2.3.0 // indirect
	sigs.k8s.io/kind v0.10.0
)
