require (
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23
	github.com/cloud-ark/kubeplus/platform-operator v0.0.0-20201027010238-8e4740bdc7ad
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/moby/term v0.0.0-20200312100748-672ec06f55cd // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/russross/blackfriday v1.5.2 // indirect
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sys v0.0.0-20200622214017-ed371f2e16b4 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	k8s.io/api v0.19.4
	k8s.io/apiextensions-apiserver v0.18.1
	k8s.io/apimachinery v0.19.4
	k8s.io/cli-runtime v0.19.4 // indirect
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/component-base v0.19.4 // indirect
	k8s.io/gengo v0.0.0-20200413195148-3a45101e95ac // indirect
	k8s.io/klog/v2 v2.2.0 // indirect
	k8s.io/kubectl v0.15.13-beta.0
	k8s.io/kubernetes v1.15.0
	k8s.io/metrics v0.19.4 // indirect
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/cloud-ark/kubeplus/platform-operator v0.0.0-20201027010238-8e4740bdc7ad => ../../platform-operator
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085212-47dc9a115b18
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085706-2090e6d8f84c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090043-8301c0bda1f0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090013-c9a0fc045dc1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085130-185d68e6e6ea
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090116-299a7b270edc
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085325-f29e2b4a4f84
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085942-b7f18460b210
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085809-589f994ddf7f
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085912-4acac5405ec6
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085838-f1cb295a73c9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090156-2138f2c9de18
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085625-3b22d835f165
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085408-1aef9010884e
)

module github.com/cloud-ark/kubeplus/platform-operator/helm-pod

go 1.13
