module github.com/mesosphere/kubeaddons-kommander-addons/test

go 1.13

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/mesosphere/kubeaddons v0.10.2
	github.com/mesosphere/kubeaddons-extrasteps v0.2.6
	go.uber.org/atomic v1.5.1 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
	sigs.k8s.io/kind v0.7.0
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
