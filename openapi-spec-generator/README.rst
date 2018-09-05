==========================
OpenAPI Spec Generator
==========================

This is a utility that you can use to generate OpenAPI Spec for Custom Resources
that are managed by your Operator.

It wraps code available in `kube-openapi repository`__ in a easy to use script.

.. _kubeopenapi: https://github.com/kubernetes/kube-openapi

__ kubeopenapi_ 


How to use?
============

1) Clone this repository and put it inside 'src' directory of your GOPATH
   at following location:

   $GOPATH/src/github.com/cloud-ark/kubeplus

2) Install dependencies:

   - cd $GOPATH/src/github.com/cloud-ark/kubeplus/openapi-spec-generator

   - dep ensure

3) Modify your types.go to include "//+k8s:openapi-gen=true" above Type declaration.
   As an example check this_.

.. _this: https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/pkg/apis/postgrescontroller/v1/types.go#L28


4) Invoke the script:

   ./build-openapi-spec.sh <Path-to-directory-where-your-types.go-is-located>

   E.g.:

   ./build-openapi-spec.sh ../operator-manager/pkg/apis/operatorcontroller/v1

   If there are no validation errors in types.go then the OpenAPI Spec will be generated
   in file named openapispec.json and this file will be copied in the input directory.

   
   If there are validation errors then fix those and try again. 