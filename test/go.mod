module github.com/mesosphere/kubeaddons-kommander-addons/test

go 1.15

replace k8s.io/client-go => k8s.io/client-go v0.19.0-rc.4

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-logr/logr v0.2.1-0.20200730175230-ee2de8da5be6 // indirect
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/mesosphere/ksphere-testing-framework v0.0.0-20200814171113-1a98809a8734
	github.com/mesosphere/kubeaddons v0.18.4-0.20200812182156-0eefd5ea7241
	go.uber.org/zap v1.15.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	k8s.io/api v0.19.0-rc.4
	k8s.io/apiextensions-apiserver v0.19.0-rc.3 // indirect
	k8s.io/apimachinery v0.19.0-rc.4
	sigs.k8s.io/kind v0.8.1
)
