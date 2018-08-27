#!/bin/bash

if [ "$#" -ne 1 ]; then
   echo "./build-openapi-doc.sh <Path to the directory containing types.go>"
   exit
fi

cp $1/types.go typedir/.
cd typedir
sed -E '/PersistentVolumeClaim|Affinity|ObjectMeta|ListMeta|LocalObjectReference|Time/s/^/\/\//' types.go > types1.go

mv types.go types.go.orig
cp types1.go types.go
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
