module github.com/kyma-incubator/api-gateway

go 1.12

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0
	gotest.tools v2.2.0+incompatible
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	knative.dev/pkg v0.0.0-20190807140856-4707aad818fe
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
)
