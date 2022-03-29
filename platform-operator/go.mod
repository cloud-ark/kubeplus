module github.com/cloud-ark/kubeplus/platform-operator

go 1.14

require (
	github.com/gogo/protobuf v1.3.2
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/lib/pq v1.0.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.0.0-20190810101755-ebc439d6a67b
	k8s.io/apimachinery v0.18.2 //#v0.0.0-20190809020650-423f5d784010
	k8s.io/client-go v0.18.2 //v8.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190808180452-d0071a119380
	k8s.io/gengo v0.0.0-20190128074634-0689ccc1d7d6
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
)
