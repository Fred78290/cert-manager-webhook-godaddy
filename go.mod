module github.com/Fred78290/cert-manager-webhook-godaddy

go 1.15

replace (
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.6
	k8s.io/client-go => k8s.io/client-go v0.19.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.6
)

require (
	github.com/jetstack/cert-manager v1.1.0
	k8s.io/apiextensions-apiserver v0.19.6
	k8s.io/apimachinery v0.19.6
	k8s.io/client-go v0.19.6
	k8s.io/klog/v2 v2.4.0
)
