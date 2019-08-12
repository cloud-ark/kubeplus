package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/tools/cache"
	//"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/client-go/kubernetes"

	restclient "k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/rest"

	admissionregistrationclientset "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	listers "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1beta1"
	apiextensionsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
)

// CRDController handles CRD objects
type CRDController struct {
	cfg *restclient.Config
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	crdsSynced    cache.InformerSynced
	crdqueue      workqueue.RateLimitingInterface
	crdLister     listers.CustomResourceDefinitionLister
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	//recorder record.EventRecorder
	//util     utils.Utils
}

// NewController returns a new sample controller
func NewCRDController(
	cfg *restclient.Config,
	kubeclientset kubernetes.Interface,
	crdListerObj listers.CustomResourceDefinitionLister,
	crdInformerFactory apiextensionsinformers.SharedInformerFactory,
	) *CRDController {

	crdInformer := crdInformerFactory.Apiextensions().V1beta1().CustomResourceDefinitions().Informer()

	// Create event broadcaster
	// Add moodle-controller types to the default Kubernetes Scheme so Events can be
	// logged for moodle-controller types.
	//operatorscheme.AddToScheme(scheme.Scheme)
	//glog.V(4).Info("Creating event broadcaster")
	//eventBroadcaster := record.NewBroadcaster()
	//eventBroadcaster.StartLogging(glog.Infof)
	//eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	//recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	//utils := utils.NewUtils(cfg, kubeclientset)

	controller := &CRDController{
		cfg:           cfg,
		kubeclientset: kubeclientset,
		crdLister: crdListerObj,
		crdsSynced:    crdInformer.HasSynced,
		crdqueue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CustomResourceDefinitions"),
		//recorder:      recorder,
		//util:          utils,
	}

	glog.Info("Setting up event handlers")

	crdInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueCRD})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *CRDController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.crdqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting CRD controller")
	//
	// // Wait for the caches to be synced before starting workers
	// glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.crdsSynced); !ok {
	 	return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")

	// Launch two workers to process CRD resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second*20, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processCRDQueueItem function in order to read and process a message on the
// workqueue.
func (c *CRDController) runWorker() {
	for c.processCRDQueueItem() {
	}
}

func (c *CRDController) processCRDQueueItem() bool {
	crdObj, shutdown := c.crdqueue.Get()
	if shutdown {
		return false
	}
	err := func(crdObj interface{}) error {
		defer c.crdqueue.Done(crdObj)
		var key string
		var ok bool
		if key, ok = crdObj.(string); !ok {
			c.crdqueue.Forget(crdObj)
			runtime.HandleError(fmt.Errorf("CRDController.go: expected string in crdqueue but got %#v", crdObj))
			return nil
		}
		if err := c.handleCRD(key); err != nil {
			return fmt.Errorf("CRDController.go: error in handleCRD key: %s, err: %s", key, err.Error())
		}
		c.crdqueue.Forget(crdObj)
		return nil
	}(crdObj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

func (c *CRDController) enqueueCRD(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	fmt.Printf("CRDController.go     : enqueueCRD adding a CRD key %s\n", key)
	c.crdqueue.AddRateLimited(key)
}

func (c *CRDController) handleCRD(key string) error {

	fmt.Printf("== Inside handleCRD function == CRD key:%s\n", key)
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("CRDController.go: invalid resource key, can't split : %s", key))
		return fmt.Errorf("CRDController.go: invalid resource key, can't split : %s", key)
	}

	cfg, _ := rest.InClusterConfig()
	crdClient, _ := apiextensionsclientset.NewForConfig(cfg)

	//crdName := "moodles.moodlecontroller.kubeplus"
	crdObj, err := crdClient.CustomResourceDefinitions().Get(name, metav1.GetOptions{})
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
	fmt.Printf("CRD Object:%v\n", crdObj)

	fmt.Println("====================================")

	admissionRegClient, _ := admissionregistrationclientset.NewForConfig(cfg)
	mutatingWebhookConfigName := "platform-as-code.crd-binding"
	mutatingWebhookObj, err := admissionRegClient.MutatingWebhookConfigurations().Get(mutatingWebhookConfigName, 
		metav1.GetOptions{})

	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
	fmt.Printf("MutatingWebhook Object:%v\n", mutatingWebhookObj)

	fmt.Println("============= Updating the Rules ==============")

	webhooks := mutatingWebhookObj.Webhooks

	webhook := webhooks[0]
	rules := webhook.Rules
	currentRule := rules[0]
	currentRule.APIGroups = append(currentRule.APIGroups, "ppp.abc")
	currentRule.Resources = append(currentRule.Resources, "ppp.abc.123")
	rules[0] = currentRule
	webhook.Rules = rules

	webhooks[0] = webhook
	mutatingWebhookObj.Webhooks = webhooks

	_, _= admissionRegClient.MutatingWebhookConfigurations().Update(mutatingWebhookObj)

	return nil
}
