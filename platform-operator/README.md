# PlatformWorkflow Operator

## Description

Platform Workflow Operator enables publishing new Services in a cluster. Cluster Admins use this Operator to govern their cluster usage by defining and registering opinionated Services with appropriate guard rails. The new Services are registered
as new Custom Resources. Application development teams consume the Services by creating instances of these Custom Resources.

## Development steps

1. Setup:

   Follow [these steps](https://cloud-ark.github.io/kubeplus/docs/html/html/compile-components.html)

2. Modify:
   - Making changes to Operator API (types.go)
   - mkdir vendor/k8s.io/code-generator/hack
   - cp hack/boilerplate.go.txt vendor/k8s.io/code-generator/hack/.
   - ./hack/update-codegen.sh

3. Build:
   Update versions.txt before creating new versioned artifact.

  ./build-artifact.sh <latest | versioned>

   Follow semver for tagging Docker images

   Update artifacts/deployment/deployment.yaml 

   **DO NOT** create vendor directory running: go mod vendor

   Vendoring seems to over-ride module versions defined in go.mod and
   brings in newer versions. This breaks the build.
   For building the code we need precisely those versions of the modules
   defined in go.mod.

   The modules will be downloaded in following location:
   $GOPATH/mod/


