module github.com/Fred78290/cert-manager-webhook-godaddy

go 1.15

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	k8s.io/api => k8s.io/api v0.18.14 // indirect
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.14
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.14
	k8s.io/apiserver => k8s.io/apiserver v0.18.14 // indirect
	k8s.io/client-go => k8s.io/client-go v0.18.14
	k8s.io/component-base => k8s.io/component-base v0.18.14 // indirect
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.14 // indirect
)

require (
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jetstack/cert-manager v1.0.0-beta.0
	github.com/kr/pretty v0.2.1 // indirect
	github.com/miekg/dns v1.1.31 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200622214017-ed371f2e16b4 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a // indirect
	google.golang.org/grpc v1.27.0 // indirect
	k8s.io/apiextensions-apiserver v0.18.14
	k8s.io/apimachinery v0.18.14
	k8s.io/client-go v0.18.14
	k8s.io/klog/v2 v2.4.0
	k8s.io/kube-aggregator v0.18.14 // indirect
)
