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

	clientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/clientset/versioned"
	informers "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/informers/externalversions"
	"github.com/cloud-ark/kubeplus/platform-operator/pkg/signals"

	"k8s.io/client-go/rest"
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
	//resourcePolicyController := NewResourcePolicyController(kubeClient, platformOperatorClient, kubeInformerFactory, platformInformerFactory)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		platformController.Run(1, ctx.Done())
	}()

	/*
	go func() {
		defer wg.Done()
		resourcePolicyController.Run(1, ctx.Done())
	}()
	*/

	go kubeInformerFactory.Start(ctx.Done())
	go platformInformerFactory.Start(ctx.Done())

	<-stopCh
	cancel()
}
