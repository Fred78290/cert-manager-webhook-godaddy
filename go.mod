module github.com/Fred78290/cert-manager-webhook-godaddy

go 1.16

replace (

	// To be replaced once there is a release of kubernetes/apiserver that uses gnostic v0.5. See https://github.com/jetstack/cert-manager/pull/3926#issuecomment-828923436
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1

	// See https://github.com/jetstack/cert-manager/issues/3999
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.12.1

	github.com/onsi/gomega => github.com/onsi/gomega v1.10.1
	golang.org/x/net => golang.org/x/net v0.0.0-20210224082022-3d97a244fca7

	// See https://github.com/kubernetes/kubernetes/issues/101567
	k8s.io/code-generator => github.com/kmodules/code-generator v0.21.1-rc.0.0.20210428003838-7eafae069eb0

	k8s.io/gengo => github.com/kmodules/gengo v0.0.0-20210428002657-a8850da697c2

	// See https://github.com/kubernetes/kubernetes/pull/99817
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
)

require (
	github.com/jetstack/cert-manager v1.4.0
	k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/klog/v2 v2.9.0
)
