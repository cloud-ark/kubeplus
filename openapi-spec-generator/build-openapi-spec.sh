#!/bin/bash

if [ "$#" -ne 1 ]; then
   echo "./build-openapi-doc.sh <relative path to the directory containing types.go>"
   exit
fi

cp $1/types.go typedir/.
cd typedir
sed s'#metav1.ObjectMeta#//metav1.ObjectMeta#'g types.go > types1.go
sed s'#metav1.ListMeta#//metav1.ListMeta#'g types1.go > types2.go
mv types.go types.go.orig
cp types2.go types.go
cd ..
go run openapi-gen.go
go run builder.go
cp openapispec.json $1
echo "OpenAPI Spec file generated and copied to $1"