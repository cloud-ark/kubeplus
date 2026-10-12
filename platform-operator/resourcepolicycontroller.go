package main

import (
	"context"
	"fmt"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	platformworkflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
	workflowinformers "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/informers/externalversions/workflowcontroller/v1alpha1"
	platformlisters "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/listers/workflowcontroller/v1alpha1"
)

// ResourcePolicyController reconciles network access independently of ResourceComposition.
type ResourcePolicyController struct {
	kubeclientset          kubernetes.Interface
	resourcePoliciesLister platformlisters.ResourcePolicyLister
	resourcePoliciesSynced cache.InformerSynced
	workqueue              workqueue.RateLimitingInterface
}

// NewResourcePolicyController creates a workqueue and registers policy event handlers.
func NewResourcePolicyController(kubeclientset kubernetes.Interface, informer workflowinformers.ResourcePolicyInformer) *ResourcePolicyController {
	c := &ResourcePolicyController{
		kubeclientset:          kubeclientset,
		resourcePoliciesLister: informer.Lister(),
		resourcePoliciesSynced: informer.Informer().HasSynced,
		workqueue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ResourcePolicies"),
	}
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueResourcePolicy,
		UpdateFunc: func(old, new interface{}) {
			oldPolicy := old.(*platformworkflowv1alpha1.ResourcePolicy)
			newPolicy := new.(*platformworkflowv1alpha1.ResourcePolicy)
			// Skip periodic resync events that carry no new resource version.
			if newPolicy.ResourceVersion != oldPolicy.ResourceVersion {
				c.enqueueResourcePolicy(newPolicy)
			}
		},
		DeleteFunc: c.enqueueResourcePolicy,
	})
	return c
}

// Run waits for the cache to sync, then starts workers until stopCh closes.
func (c *ResourcePolicyController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	if ok := cache.WaitForCacheSync(stopCh, c.resourcePoliciesSynced); !ok {
		return fmt.Errorf("failed to wait for resource policy informer cache to sync")
	}
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	<-stopCh
	return nil
}

// runWorker processes queued policies until the queue shuts down.
func (c *ResourcePolicyController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// enqueueResourcePolicy queues a namespace/name key, including deletion tombstones.
func (c *ResourcePolicyController) enqueueResourcePolicy(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// processNextWorkItem reconciles one key and retries failures with backoff.
func (c *ResourcePolicyController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		// Done releases the processing marker; Forget clears retry history.
		defer c.workqueue.Done(obj)
		key, ok := obj.(string)
		if !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in resource policy workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncResourcePolicy(key); err != nil {
			return fmt.Errorf("error syncing resource policy %q: %w", key, err)
		}
		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		c.workqueue.AddRateLimited(obj)
	}
	return true
}

// syncResourcePolicy applies the latest cached policy or cleans up a deleted policy.
func (c *ResourcePolicyController) syncResourcePolicy(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid resource policy key %q: %w", key, err)
	}
	if namespace == "" {
		namespace = "default"
	}

	policy, err := c.resourcePoliciesLister.ResourcePolicies(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return c.deleteResourcePolicyNetworkPolicies(namespace, name)
		}
		return err
	}

	// Keep these namespace/name identities when pruning after all rules succeed.
	desiredPolicies := make(map[string]struct{}, len(policy.Spec.Policy.Network.Access))
	for _, rule := range policy.Spec.Policy.Network.Access {
		serviceNamespace := rule.ServiceRef.Namespace
		if serviceNamespace == "" {
			serviceNamespace = namespace
		}
		service, err := c.kubeclientset.CoreV1().Services(serviceNamespace).Get(context.Background(), rule.ServiceRef.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get service %s/%s for resource policy %s/%s: %w", serviceNamespace, rule.ServiceRef.Name, namespace, name, err)
		}

		desired, err := buildNetworkPolicy(rule, *service)
		if err != nil {
			return fmt.Errorf("build network policy for resource policy %s/%s: %w", namespace, name, err)
		}
		if desired.Annotations == nil {
			desired.Annotations = map[string]string{}
		}
		desired.Annotations[networkPolicyResourcePolicyAnnotation] = namespace + "/" + name
		if err := c.upsertNetworkPolicy(desired); err != nil {
			return fmt.Errorf("apply network policy %s/%s: %w", desired.Namespace, desired.Name, err)
		}
		desiredPolicies[desired.Namespace+"/"+desired.Name] = struct{}{}
	}

	return c.pruneResourcePolicyNetworkPolicies(namespace+"/"+name, desiredPolicies)
}

// deleteResourcePolicyNetworkPolicies reuses pruning with no policies to retain.
func (c *ResourcePolicyController) deleteResourcePolicyNetworkPolicies(namespace, name string) error {
	return c.pruneResourcePolicyNetworkPolicies(namespace+"/"+name, nil)
}

// pruneResourcePolicyNetworkPolicies deletes owned policies no longer desired across namespaces.
func (c *ResourcePolicyController) pruneResourcePolicyNetworkPolicies(resourcePolicyRef string, desired map[string]struct{}) error {
	policies, err := c.kubeclientset.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for i := range policies.Items {
		policy := &policies.Items[i]
		if policy.Labels[CREATED_BY_KEY] != CREATED_BY_VALUE || policy.Annotations[networkPolicyResourcePolicyAnnotation] != resourcePolicyRef {
			continue
		}
		if _, ok := desired[policy.Namespace+"/"+policy.Name]; ok {
			continue
		}
		if err := c.kubeclientset.NetworkingV1().NetworkPolicies(policy.Namespace).Delete(context.Background(), policy.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

// upsertNetworkPolicy creates or updates a policy, rejecting unmanaged or conflicting owners.
func (c *ResourcePolicyController) upsertNetworkPolicy(desired *networkingv1.NetworkPolicy) error {
	policies := c.kubeclientset.NetworkingV1().NetworkPolicies(desired.Namespace)
	existing, err := policies.Get(context.Background(), desired.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = policies.Create(context.Background(), desired, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	if existing.Labels[CREATED_BY_KEY] != CREATED_BY_VALUE {
		return fmt.Errorf("existing network policy is not managed by KubePlus")
	}
	// Allow adoption of KubePlus policies created before the ownership annotation.
	if owner := existing.Annotations[networkPolicyResourcePolicyAnnotation]; owner != "" && owner != desired.Annotations[networkPolicyResourcePolicyAnnotation] {
		return fmt.Errorf("existing network policy is owned by resource policy %q", owner)
	}

	// Use the server version so an intervening update is reported as a conflict.
	desired.ResourceVersion = existing.ResourceVersion
	_, err = policies.Update(context.Background(), desired, metav1.UpdateOptions{})
	return err
}
