#!/bin/bash

if [ "$#" -ne 1 ]; then
   echo "./build-openapi-doc.sh <Path to the directory containing types.go>"
   exit
fi

mkdir -p vendor/k8s.io/gengo/boilerplate
cp boilerplate.go.txt vendor/k8s.io/gengo/boilerplate/.

cp $1/types.go typedir/.
cd typedir
sed -E '/PersistentVolumeClaim|Affinity|ObjectMeta|ListMeta|LocalObjectReference|Time/s/^/\/\//' types.go > types1.go
sed -E '/package */s/^/\/\//' types1.go > types2.go
sed -e '1i\
package typedir
' < types2.go > types3.go

mv types1.go types1.go.bak
mv types2.go types2.go.bak
mv types.go types.go.orig
mv types3.go types.go

cd ..
op1=`go run openapi-gen.go`
op2=`go run builder.go`

echo "$op1"

if [[ $op1 = *"API rule violation"* ]]; then
   echo FAIL
   echo "API rule violation"
else
   echo OK
   echo "OpenAPI Spec file generated and copied to $1"
   cp openapispec.json $1
fi
