/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This package generates openAPI definition file to be used in open API spec generation on API servers. To generate
// definition for a specific type or package add "+k8s:openapi-gen=true" tag to the type/package comment lines. To
// exclude a type from a tagged package, add "+k8s:openapi-gen=false" tag to the type comment lines.
package main

import (
	"flag"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"k8s.io/gengo/args"
	"k8s.io/kube-openapi/pkg/generators"

	generatorargs "k8s.io/kube-openapi/cmd/openapi-gen/args"
	"k8s.io/kube-openapi/pkg/util"
)

func main() {
	genericArgs, customArgs := generatorargs.NewDefaults()

	// Override defaults.
	// TODO: move this out of openapi-gen
	genericArgs.GoHeaderFilePath = filepath.Join(args.DefaultSourceTree(), util.BoilerplatePath())

	genericArgs.AddFlags(pflag.CommandLine)
	customArgs.AddFlags(pflag.CommandLine)
	flag.Set("logtostderr", "true")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if err := generatorargs.Validate(genericArgs); err != nil {
		glog.Fatalf("Error: %v", err)
	}

	// Run it.
	if err := genericArgs.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	); err != nil {
		glog.Fatalf("Error: %v", err)
	}
	glog.V(2).Info("Completed successfully.")
}
