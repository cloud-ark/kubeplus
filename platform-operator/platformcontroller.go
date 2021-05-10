package main

import (
	"fmt"
	"time"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"net/http"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
    apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	platformworkflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
	clientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/clientset/versioned"
	platformstackscheme "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/informers/externalversions"
	listers "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/listers/workflowcontroller/v1alpha1"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	"k8s.io/client-go/rest"
)

const controllerAgentName = "platformstack-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when PlatformStack is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by PlatformStack"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "PlatformStack synced successfully"

	// Annotations to put on Consumer CRDs.
	CREATED_BY_KEY = "created-by"
	CREATED_BY_VALUE = "kubeplus"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	platformStackclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	platformStacksLister        listers.ResourceCompositionLister
	platformStacksSynced        cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewPlatformController(
	kubeclientset kubernetes.Interface,
	platformStackclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	platformstackInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Deployment and PlatformStack
	// types.
	deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()
	platformStackInformer := platformstackInformerFactory.Workflows().V1alpha1().ResourceCompositions()

	// Create event broadcaster
	// Add platformstack-controller types to the default Kubernetes Scheme so Events can be
	// logged for platformstack-controller types.
	platformstackscheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		platformStackclientset:   platformStackclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		platformStacksLister:        platformStackInformer.Lister(),
		platformStacksSynced:        platformStackInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "PlatformStacks"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	platformStackInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFoo,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*platformworkflowv1alpha1.ResourceComposition)
			oldDepl := old.(*platformworkflowv1alpha1.ResourceComposition)
			//fmt.Println("New Version:%s", newDepl.ResourceVersion)
			//fmt.Println("Old Version:%s", oldDepl.ResourceVersion)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			} else {
				controller.enqueueFoo(new)
			}
		},
		DeleteFunc: func(obj interface{}) {
		        _, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
			   controller.deleteFoo(obj)
			}
		},
	})
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting PlatformStack controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.platformStacksSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueFoo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Foo" {
			return
		}

		foo, err := c.platformStacksLister.ResourceCompositions(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueFoo(foo)
		return
	}
}


func (c *Controller) deleteFoo(obj interface{}) {

	fmt.Println("Inside delete Foo")

	var err error
	if _, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
	   panic(err)
	}

	foo := obj.(*platformworkflowv1alpha1.ResourceComposition)

	fmt.Printf("JKL\n")
	fmt.Printf("%v\n", foo.Spec)
	newRes := foo.Spec.NewResource
	fmt.Printf("newRes:%v\n", newRes)

	namespace := foo.ObjectMeta.Namespace
	if namespace == "" {
		namespace = "default"
	}
	res := newRes.Resource
	fmt.Printf("GHI - delete\n")
	fmt.Printf("%v\n NS:%s", res, namespace)
	kind := foo.Spec.NewResource.Resource.Kind
	group := foo.Spec.NewResource.Resource.Group
	version := foo.Spec.NewResource.Resource.Version
	plural := foo.Spec.NewResource.Resource.Plural
	chartURL := foo.Spec.NewResource.ChartURL
	chartName := foo.Spec.NewResource.ChartName
	fmt.Printf("Kind:%s, Version:%s Group:%s, Plural:%s\n", kind, version, group, plural)
	fmt.Printf("ChartURL:%s, ChartName:%s\n", chartURL, chartName)

	action := "delete"
	handleCRD(kind, version, group, plural, action, namespace)

 	resPolicySpec := foo.Spec.ResPolicy
 	//fmt.Printf("ResPolicySpec:%v\n",resPolicySpec)

	deleteResourcePolicy(resPolicySpec, namespace)

 	resMonitorSpec := foo.Spec.ResMonitor
 	//fmt.Printf("ResMonitorSpec:%v\n",resMonitorSpec)

	deleteResourceMonitor(resMonitorSpec, namespace)

	c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	fmt.Printf("Inside syncHandler...key:%s\n", key)
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	fmt.Printf("Inside syncHandler...Namespace:%s Name:%s\n", namespace, name)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	if namespace == "" {
		namespace = "default"
	}

	// Get the Foo resource with this namespace/name
	foo, err := c.platformStacksLister.ResourceCompositions(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("platformStack '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	fmt.Printf("ABC\n")
	fmt.Printf("%v\n", foo.Spec)
	newRes := foo.Spec.NewResource
	fmt.Printf("DEF\n")
	fmt.Printf("%v\n", newRes)
	res := newRes.Resource
	fmt.Printf("GHI\n")
	fmt.Printf("%v\n", res)
	kind := foo.Spec.NewResource.Resource.Kind
	group := foo.Spec.NewResource.Resource.Group
	version := foo.Spec.NewResource.Resource.Version
	plural := foo.Spec.NewResource.Resource.Plural
	chartURL := foo.Spec.NewResource.ChartURL
	chartName := foo.Spec.NewResource.ChartName
	fmt.Printf("Kind:%s, Version:%s Group:%s, Plural:%s\n", kind, version, group, plural)
	fmt.Printf("ChartURL:%s, ChartName:%s\n", chartURL, chartName)
	// Check if CRD is present or not. Create it only if it is not present.
	action := "create"
	handleCRD(kind, version, group, plural, action, namespace)

 	resPolicySpec := foo.Spec.ResPolicy
 	fmt.Printf("ResPolicySpec:%v\n",resPolicySpec)

	// Instantiate ResourcePolicy object
	createResourcePolicy(resPolicySpec, namespace)

 	resMonitorSpec := foo.Spec.ResMonitor
 	fmt.Printf("ResMonitorSpec:%v\n",resMonitorSpec)

	// Instantiate ResourceMonitor object
	createResourceMonitor(resMonitorSpec, namespace)

	c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func createResourceMonitor(resMonitorSpec interface{}, namespace string) {
	fmt.Println("Inside createResourceMonitor")
	resMonitorObject := resMonitorSpec.(platformworkflowv1alpha1.ResourceMonitor)
	//resPolicySpecMap := resPolicySpec.(map[string]interface{})

	// Using Typed client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	var sampleclientset clientset.Interface
	sampleclientset = clientset.NewForConfigOrDie(config)

	resMonitor, err := sampleclientset.WorkflowsV1alpha1().ResourceMonitors(namespace).Create(&resMonitorObject)
	fmt.Printf("ResourceMonitor:%v\n", resMonitor)
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
}

func deleteResourceMonitor(resMonitorSpec interface{}, namespace string) {	
	fmt.Println("Inside deleteResourceMonitor")
	resMonitorObject := resMonitorSpec.(platformworkflowv1alpha1.ResourceMonitor)
	inputResMonitorName := resMonitorObject.ObjectMeta.Name
	/*namespace := resMonitorObject.ObjectMeta.Namespace
	if namespace == "" {
		namespace = "default"
	}*/
	fmt.Printf("ResMonitor:%s, Namespace:%s\n",inputResMonitorName,namespace)

	// Using Typed client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	var sampleclientset clientset.Interface
	sampleclientset = clientset.NewForConfigOrDie(config)

	resMonList, err := sampleclientset.WorkflowsV1alpha1().ResourceMonitors(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
	for _, resMon := range resMonList.Items {
		resMonName := resMon.ObjectMeta.Name
		if resMonName == inputResMonitorName {
			fmt.Printf("Deleting ResMonitor %s\n", resMonName)
			err := sampleclientset.WorkflowsV1alpha1().ResourceMonitors(namespace).Delete(resMonName, &metav1.DeleteOptions{})
			if err != nil {
				fmt.Errorf("Error:%s\n", err)
			}
		}
	}
}

func createResourcePolicy(resPolicySpec interface{}, namespace string) {
	fmt.Println("Inside createResourcePolicy")
	resPolicyObject := resPolicySpec.(platformworkflowv1alpha1.ResourcePolicy)
	//resPolicySpecMap := resPolicySpec.(map[string]interface{})

	// Using Typed client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	var sampleclientset clientset.Interface
	sampleclientset = clientset.NewForConfigOrDie(config)

	resPolicy, err := sampleclientset.WorkflowsV1alpha1().ResourcePolicies(namespace).Create(&resPolicyObject)
	fmt.Printf("ResourcePolicy:%v\n", resPolicy)
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
}

func deleteResourcePolicy(resPolicySpec interface{}, namespace string) {	
	fmt.Println("Inside deleteResourcePolicy.")
	resPolicyObject := resPolicySpec.(platformworkflowv1alpha1.ResourcePolicy)
	inputResPolicyName := resPolicyObject.ObjectMeta.Name
	/*namespace := resPolicyObject.ObjectMeta.Namespace
	if namespace == "" {
		namespace = "default"
	}*/
	fmt.Printf("ResPolicy:%s, Namespace:%s\n",inputResPolicyName,namespace)

	// Using Typed client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	var sampleclientset clientset.Interface
	sampleclientset = clientset.NewForConfigOrDie(config)

	resPolicyList, err := sampleclientset.WorkflowsV1alpha1().ResourcePolicies(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
	for _, resPolicy := range resPolicyList.Items {
		resPolicyName := resPolicy.ObjectMeta.Name
		if inputResPolicyName == resPolicyName {
			fmt.Printf("Deleting ResPolicy object %s\n", resPolicyName)
			err := sampleclientset.WorkflowsV1alpha1().ResourcePolicies(namespace).Delete(resPolicyName, &metav1.DeleteOptions{})
			if err != nil {
				fmt.Errorf("Error:%s\n", err)
			}
		}
	}
}

func handleCRD(kind, version, group, plural, action, namespace string) error {
	fmt.Printf("Inside handleCRD %s\n", action)
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	crdClient, _ := apiextensionsclientset.NewForConfig(cfg)

	kubePlusAnnotation := make(map[string]string)
	kubePlusAnnotation[CREATED_BY_KEY] = CREATED_BY_VALUE

	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: plural + "." + group,
			Annotations: kubePlusAnnotation,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group: group,
			Version: version,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: plural,
				Kind: kind,
			},
		},
	}

	crdPresent := false
	crdList, err := crdClient.CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
		return err
	}
	crdToHandle := ""
	for _, crd := range crdList.Items {
		crdToHandle = crd.ObjectMeta.Name
		crdObj, err := crdClient.CustomResourceDefinitions().Get(crdToHandle, metav1.GetOptions{})
		if err != nil {
			fmt.Errorf("Error:%s\n", err)
			return err
		}
		group1 := crdObj.Spec.Group
		version1 := crdObj.Spec.Version
		//endpoint := "apis/" + group + "/" + version
		kind1 := crdObj.Spec.Names.Kind
		plural1 := crdObj.Spec.Names.Plural
		//fmt.Printf("Kind:%s, Group:%s, Version:%s, Endpoint:%s, Plural:%s\n",kind1, group1, version1, endpoint, plural1)

		if group == group1 && kind == kind1 && version == version1 && plural == plural1 {
			crdPresent = true
			break
		}
	}

	if !crdPresent {
		if action == "create" {
			_, err1 := crdClient.CustomResourceDefinitions().Create(crd)
			if err1 != nil {
				panic(err1.Error())
			}
		}
	} else {
		fmt.Printf("CRD Group:%s Version:%s Kind:%s Plural:%s found.\n", group, version, kind, plural)
		if action == "delete" {
			deleteCRDInstances(kind, group, version, plural, namespace)
			err := crdClient.CustomResourceDefinitions().Delete(crdToHandle, &metav1.DeleteOptions{})
			if err != nil {
				fmt.Errorf("Error:%s\n", err)
				return err
			} else {
				fmt.Printf("CRD deleted successfully.\n")
			}
		}
	}
	return nil
}

func deleteCRDInstances(kind, group, version, plural, namespace string) []byte {
	fmt.Printf("Inside deleteCRDInstances...\n")
	args := fmt.Sprintf("kind=%s&group=%s&version=%s&plural=%s&namespace=%s", kind, group, version, plural, namespace)
	serviceHost, servicePort := getServiceEndpoint("kubeplus")
	fmt.Printf("After getServiceEndpoint...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/kubeplus/deletecrdinstances?%s", serviceHost, servicePort, args)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func getServiceEndpoint(servicename string) (string, string) {
	fmt.Printf("..Inside getServiceEndpoint...\n")
	namespace := getKubePlusNamespace() // Use the namespace in which kubeplus is deployed.
	//discoveryService := "discovery-service"
	cfg, _ := rest.InClusterConfig()
	kubeClient, _ := kubernetes.NewForConfig(cfg)
	serviceClient := kubeClient.CoreV1().Services(namespace)
	discoveryServiceObj, _ := serviceClient.Get(servicename, metav1.GetOptions{})
	host := discoveryServiceObj.Spec.ClusterIP
	port := discoveryServiceObj.Spec.Ports[0].Port
	stringPort := strconv.Itoa(int(port))
    fmt.Printf("Host:%s, Port:%s\n", host, stringPort)
	return host, stringPort
}

func getKubePlusNamespace() string {
	filePath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	content, err := ioutil.ReadFile(filePath)
    if err != nil {
     	fmt.Printf("Namespace file reading error:%v\n", err)
    }
    ns := string(content)
    ns = strings.TrimSpace(ns)
    //fmt.Printf("CRD Hook NS:%s\n", ns)
    return ns
}

func queryKubeDiscoveryService(url1 string) []byte {
	fmt.Printf("..inside queryKubeDiscoveryService")
	u, err := url.Parse(url1)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("sending request failed: %s", err.Error())
		fmt.Println(err)
	}
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)

	//fmt.Println(resp.Status)
	//fmt.Println(string(resp_body))
	//fmt.Println("Exiting QueryCompositionEndpoint")
	return resp_body
}

func int32Ptr(i int32) *int32 { return &i }
