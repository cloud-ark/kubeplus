==========================
OpenAPI Spec Generator
==========================

This is a utility that you can use to generate OpenAPI Spec for Custom Resources
that are managed by your Operator.

It wraps code available in [kube-openapi repository](https://github.com/kubernetes/kube-openapi) in a easy to use script.


How to use?
============

1) Clone this repository and put it inside 'src' directory of your GOPATH
   at following location:

   $GOPATH/src/github.com/cloud-ark/kubeplus

2) Install dependencies:

   - cd $GOPATH/src/github.com/cloud-ark/kubeplus/openapi-spec-generator

   - dep ensure

3) Invoke the script:

   ./build-openapi-spec.sh <relative-path-to-directory-where-your-types.go-is-located>

   E.g.:

   ./build-openapi-spec.sh ../operator-manager/pkg/apis/operatorcontroller/v1

   If there are no validation errors in types.go then the OpenAPI Spec will be generated
   in file named openapispec.json and this file will be copied in the input directory.