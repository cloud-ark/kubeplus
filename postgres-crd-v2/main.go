/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"flag"
	"time"
	"fmt"

	"github.com/golang/glog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/client/clientset/versioned"
	informers "github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/client/informers/externalversions"
	"github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/signals"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	crdtypedef "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	masterURL  string
	kubeconfig string
	firstTime bool
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	/*
	if firstTime {
		firstTime = false
		crdclient, err := apiextensionsclientset.NewForConfig(cfg)
		if err != nil {
			fmt.Println("Could not Register CRD. Register it using kubectl apply once the controller starts up.")
		} else {
			registerCRD(crdclient)
		}
	}
	*/

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	controller := NewController(kubeClient, exampleClient, kubeInformerFactory, exampleInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go exampleInformerFactory.Start(stopCh)

	if err = controller.Run(1, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	//firstTime = true
}

func registerCRD(crdclient *apiextensionsclientset.Clientset) {
	fmt.Println("Inside registerCRD")
	crdName := "postgreses.postgrescontroller.kubeplus"
	crdGroup := "postgrescontroller.kubeplus"
	crdVersion := "v1"
	crdKind := "Postgres"
	crdPlural := "postgreses"
	postgrescrd := &crdtypedef.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: crdtypedef.CustomResourceDefinitionSpec{
			Group: crdGroup,
			Version: crdVersion,
			Names: crdtypedef.CustomResourceDefinitionNames{
				Plural: crdPlural,
				Kind: crdKind,
			},
		},
	}

	postgrescrdObj, err := crdclient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
	if (postgrescrdObj == nil || err != nil) {
		postgrescrdObj, err = crdclient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(postgrescrd)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Postgres object created:%v", postgrescrdObj)
		}
	}
}