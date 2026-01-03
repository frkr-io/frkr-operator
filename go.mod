module github.com/frkr-io/frkr-operator

go 1.21

require (
	github.com/frkr-io/frkr-common v0.0.0
	k8s.io/apimachinery v0.29.0
	k8s.io/client-go v0.29.0
	sigs.k8s.io/controller-runtime v0.17.0
)

replace github.com/frkr-io/frkr-common => ../frkr-common

