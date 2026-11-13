package main

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	workflowinformers "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/informers/externalversions/workflowcontroller/v1alpha1"
	workflowlisters "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/listers/workflowcontroller/v1alpha1"
)

const (
	crossNamespacePolicyName  = "restrict-cross-ns-traffic"
	networkAccessManagedLabel = "kubeplus.io/network-access-managed"
	namespaceNameLabel        = "kubernetes.io/metadata.name"
)

type ResourcePolicyController struct {
	kubeclientset kubernetes.Interface
	policies      workflowlisters.ResourcePolicyLister
	synced        cache.InformerSynced
	workqueue     workqueue.RateLimitingInterface
}

func NewResourcePolicyController(client kubernetes.Interface, informer workflowinformers.ResourcePolicyInformer) *ResourcePolicyController {
	c := &ResourcePolicyController{
		kubeclientset: client,
		policies:      informer.Lister(),
		synced:        informer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ResourcePolicyNetworkAccess"),
	}
	// A single queue key is enough because each pass reconciles the whole whitelist.
	enqueue := func(interface{}) { c.workqueue.Add("network-access") }
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: enqueue,
		// Resync events also repair manual changes to the NetworkPolicies.
		UpdateFunc: func(_, _ interface{}) { enqueue(nil) },
		DeleteFunc: enqueue,
	})
	return c
}

func (c *ResourcePolicyController) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()
	// Wait for the full initial list before deciding which allowances to remove.
	if !cache.WaitForCacheSync(stopCh, c.synced) {
		return fmt.Errorf("failed to sync ResourcePolicy informer")
	}
	// Cleanup must also run when the initial ResourcePolicy list is empty.
	c.workqueue.Add("network-access")
	go c.runWorker()
	<-stopCh
	return nil
}

func (c *ResourcePolicyController) runWorker() {
	defer runtime.HandleCrash()
	for {
		item, shutdown := c.workqueue.Get()
		if shutdown {
			return
		}
		if err := c.syncNetworkAccess(); err != nil {
			runtime.HandleError(err)
			c.workqueue.AddRateLimited(item)
		} else {
			c.workqueue.Forget(item)
		}
		c.workqueue.Done(item)
	}
}

func (c *ResourcePolicyController) syncNetworkAccess() error {
	policies, err := c.policies.List(labels.Everything())
	if err != nil {
		return err
	}
	// allowed maps each destination namespace to its permitted source namespaces.
	allowed := make(map[string]sets.String)
	for _, policy := range policies {
		if policy.Spec.Policy.Network == nil {
			continue
		}
		for _, pair := range policy.Spec.Policy.Network.AllowedPairs {
			if allowed[pair.Namespace1] == nil {
				allowed[pair.Namespace1] = sets.NewString()
			}
			if allowed[pair.Namespace2] == nil {
				allowed[pair.Namespace2] = sets.NewString()
			}
			allowed[pair.Namespace1].Insert(pair.Namespace2)
			allowed[pair.Namespace2].Insert(pair.Namespace1)
		}
	}

	// One list also finds marked policies whose ResourcePolicy disappeared while offline.
	networkPolicies, err := c.kubeclientset.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	existing := make(map[string]*networkingv1.NetworkPolicy)
	for i := range networkPolicies.Items {
		policy := &networkPolicies.Items[i]
		if policy.Name == crossNamespacePolicyName {
			existing[policy.Namespace] = policy
		}
	}
	var errs []error
	for namespace := range allowed {
		if existing[namespace] == nil {
			errs = append(errs, fmt.Errorf("NetworkPolicy %s/%s does not exist", namespace, crossNamespacePolicyName))
		}
	}
	for namespace, policy := range existing {
		peers := allowed[namespace]
		// Missing endpoints must not block unrelated revocations or grant half a pair.
		for peer := range peers {
			if existing[peer] == nil {
				peers.Delete(peer)
			}
		}
		// Previously managed namespaces need cleanup even after their last pair is removed.
		if len(peers) == 0 && policy.Labels[networkAccessManagedLabel] != "true" {
			continue
		}
		desired := networkPolicyWithAllowedNamespaces(policy, peers)
		if equality.Semantic.DeepEqual(policy, desired) {
			continue
		}
		if _, err := c.kubeclientset.NetworkingV1().NetworkPolicies(namespace).Update(context.Background(), desired, metav1.UpdateOptions{}); err != nil {
			errs = append(errs, fmt.Errorf("update NetworkPolicy %s/%s: %w", namespace, policy.Name, err))
		}
	}
	return utilerrors.NewAggregate(errs)
}

func networkPolicyWithAllowedNamespaces(policy *networkingv1.NetworkPolicy, allowed sets.String) *networkingv1.NetworkPolicy {
	desired := policy.DeepCopy()
	var ingress []networkingv1.NetworkPolicyIngressRule
	// Preserve local and custom rules; replace only the namespace-wide allowances.
	for _, rule := range desired.Spec.Ingress {
		if !isPythonNamespaceAllowance(rule) {
			ingress = append(ingress, rule)
		}
	}
	for _, namespace := range allowed.List() {
		ingress = append(ingress, networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{{
				NamespaceSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      namespaceNameLabel,
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{namespace},
				}}},
			}},
		})
	}
	desired.Spec.Ingress = ingress
	if len(allowed) != 0 {
		if desired.Labels == nil {
			desired.Labels = make(map[string]string)
		}
		desired.Labels[networkAccessManagedLabel] = "true"
	} else {
		delete(desired.Labels, networkAccessManagedLabel)
	}
	return desired
}

// Only replace the namespace-wide, all-port rules produced by network_traffic.py.
func isPythonNamespaceAllowance(rule networkingv1.NetworkPolicyIngressRule) bool {
	if len(rule.Ports) != 0 || len(rule.From) != 1 {
		return false
	}
	peer := rule.From[0]
	if peer.PodSelector != nil || peer.IPBlock != nil || peer.NamespaceSelector == nil {
		return false
	}
	selector := peer.NamespaceSelector
	if len(selector.MatchLabels) != 0 || len(selector.MatchExpressions) != 1 {
		return false
	}
	expr := selector.MatchExpressions[0]
	return expr.Key == namespaceNameLabel && expr.Operator == metav1.LabelSelectorOpIn && len(expr.Values) == 1
}
