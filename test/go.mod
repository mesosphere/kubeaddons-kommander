module github.com/mesosphere/kubeaddons-kommander-addons/test

go 1.15

replace (
	k8s.io/client-go => k8s.io/client-go v0.19.1
	sigs.k8s.io/kind => sigs.k8s.io/kind v0.9.0 // locked the kind version as we've run into troubles in the past when this gets unintentionally bumped, we want to intentionally manage this version
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/go-logr/logr v0.2.1-0.20200730175230-ee2de8da5be6 // indirect
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/mesosphere/ksphere-testing-framework v0.2.0
	github.com/mesosphere/kubeaddons v0.22.2
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	k8s.io/api v0.19.1
	k8s.io/apiextensions-apiserver v0.19.0-rc.3 // indirect
	k8s.io/apimachinery v0.19.1
	k8s.io/klog/v2 v2.3.0 // indirect
	k8s.io/utils v0.0.0-20200731180307-f00132d28269 // indirect
	sigs.k8s.io/kind v0.9.0
)
