package main

import (
	"fmt"
	"time"

	_ "github.com/lib/pq"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	workflowcontrollerv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
	clientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/clientset/versioned"
	platformstackscheme "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/clientset/versioned/scheme"
	informers "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/informers/externalversions"
	listers "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/listers/workflowcontroller/v1alpha1"
)

const controllerAgent = "resourcepolicy-controller"

const (
	// SuccessSyncedResourcePolicy is used as part of the Event 'reason' when ResourcePolicy is synced
	SuccessSyncedResourcePolicy = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExistsResourcePolicy = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExistsResourcePolicy = "Resource %q already exists and is not managed by ResourcePolicy"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSyncedResourcePolicy = "ResourcePolicy synced successfully"
)

// Controller is the controller implementation for Foo resources
type ResourcePolicyController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	resourcePolicyclientset clientset.Interface

	resourcePolicyLister        listers.ResourcePolicyLister
	resourcePoliciesSynced        cache.InformerSynced

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
func NewResourcePolicyController(
	kubeclientset kubernetes.Interface,
	resourcePolicyclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	resourcePolicyInformerFactory informers.SharedInformerFactory) *ResourcePolicyController {

	resourcePolicyInformer := resourcePolicyInformerFactory.Workflows().V1alpha1().ResourcePolicies()

	// Create event broadcaster
	// Add platformstack-controller types to the default Kubernetes Scheme so Events can be
	// logged for platformstack-controller types.
	platformstackscheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgent})

	controller := &ResourcePolicyController{
		kubeclientset:     kubeclientset,
		resourcePolicyclientset:   resourcePolicyclientset,
		resourcePolicyLister:        resourcePolicyInformer.Lister(),
		resourcePoliciesSynced:        resourcePolicyInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ResourcePolicies"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	resourcePolicyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFoo,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*workflowcontrollerv1alpha1.ResourcePolicy)
			oldDepl := old.(*workflowcontrollerv1alpha1.ResourcePolicy)
			//fmt.Println("New Version:%s", newDepl.ResourceVersion)
			//fmt.Println("Old Version:%s", oldDepl.ResourceVersion)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known ResourcePolicies.
				// Two different versions of the same ResourcePolicy will always have different RVs.
				return
			} else {
				controller.enqueueFoo(new)
			}
		},
		/*
		DeleteFunc: func(obj interface{}) {
		        _, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
			   controller.deleteFoo(obj)
			}
		},*/
	})
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *ResourcePolicyController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting ResourcePolicy controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.resourcePoliciesSynced); !ok {
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
func (c *ResourcePolicyController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *ResourcePolicyController) processNextWorkItem() bool {
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
func (c *ResourcePolicyController) enqueueFoo(obj interface{}) {
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
func (c *ResourcePolicyController) handleObject(obj interface{}) {
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

		foo, err := c.resourcePolicyLister.ResourcePolicies(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueFoo(foo)
		return
	}
}


func (c *ResourcePolicyController) deleteFoo(obj interface{}) {

	fmt.Println("Inside delete Foo")

	var err error
	if _, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
	   panic(err)
	}
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *ResourcePolicyController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Foo resource with this namespace/name
	foo, err := c.resourcePolicyLister.ResourcePolicies(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("resourcePolicy '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	kind := foo.Spec.Resource.Kind
	version := foo.Spec.Resource.Version
	group := foo.Spec.Resource.Group
	plural := foo.Spec.Resource.Plural

	fmt.Printf("Custom Resource Kind:%v\n", kind)
	fmt.Printf("Custom Resource Version:%v\n", version)
	fmt.Printf("Custom Resource Group:%v\n", group)
	fmt.Printf("Custom Resource Plural:%v\n", plural)

	c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSyncedResourcePolicy, MessageResourceSyncedResourcePolicy)
	return nil
}
