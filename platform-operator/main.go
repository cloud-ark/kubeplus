package main

import (
	"flag"
	"time"
	"context"
	"sync"

	"github.com/golang/glog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/clientset/versioned"
	informers "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/informers/externalversions"
	"github.com/cloud-ark/kubeplus/platform-operator/pkg/signals"

	"k8s.io/client-go/rest"

	//apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset" //typed/apiextensions/v1beta1"
	//apiextensionsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"

)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	ctx, cancel := context.WithCancel(context.Background())

	// creates the in-cluster config
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	platformOperatorClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	platformInformerFactory := informers.NewSharedInformerFactory(platformOperatorClient, time.Second*30)
	platformController := NewPlatformController(kubeClient, platformOperatorClient, kubeInformerFactory, platformInformerFactory)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		platformController.Run(1, ctx.Done())
	}()

/*
	crdClient := apiextensionsclientset.NewForConfigOrDie(cfg)
	crdInformerFactory := apiextensionsinformers.NewSharedInformerFactory(crdClient, time.Second*30)

	crdController := NewCRDController(cfg,
		kubeClient,
		crdInformerFactory.Apiextensions().V1beta1().CustomResourceDefinitions().Lister(),
		crdInformerFactory)

	wg.Add(1)
	go func() {
		defer wg.Done()
		crdController.Run(1, ctx.Done())
	}()
*/

	go kubeInformerFactory.Start(ctx.Done())
	go platformInformerFactory.Start(ctx.Done())

	<-stopCh
	cancel()
	/*
	if err = platformController.Run(1, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
	*/
}

