PlatformStack Operator
-----------------------

Use Golang version 1.13 (1.14 is also okay)

Turn off GOMODULES:
export GO111MODULE=off

See setgopath.sh

source setgopath.sh

Making changes to Operator API (types.go)
-------------------------------------------
mkdir vendor/k8s.io/code-generator/hack
cp hack/boilerplate.go.txt vendor/k8s.io/code-generator/hack/.
./hack/update-codegen.sh


./build-artifact.sh <latest | versioned>

Update versions.txt before creating new versioned artifact.

Follow semver for tagging Docker images

Update artifacts/deployment/deployment.yaml 

**DO NOT** create vendor directory running: go mod vendor

Vendoring seems to over-ride module versions defined in go.mod and
brings in newer versions. This breaks the build.
For building the code we need precisely those versions of the modules
defined in go.mod.

The modules will be downloaded in following location:
$GOPATH/mod/

