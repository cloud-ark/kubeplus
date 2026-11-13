package main

import (
	"context"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	workflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
	workflowfake "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/clientset/versioned/fake"
	workflowinformers "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/informers/externalversions"
)

// The empty peer PodSelector allows traffic from Pods in the same namespace.
func baselinePolicy(namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: crossNamespacePolicyName, Namespace: namespace},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress:     []networkingv1.NetworkPolicyIngressRule{{From: []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}}}},
		},
	}
}

// This is the all-Pods, all-ports rule written by network_traffic.py.
func namespaceAllowance(namespace string) networkingv1.NetworkPolicyIngressRule {
	return networkingv1.NetworkPolicyIngressRule{From: []networkingv1.NetworkPolicyPeer{{
		NamespaceSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{namespace},
		}}},
	}}}
}

func whitelist(name string, pairs ...[2]string) *workflowv1alpha1.ResourcePolicy {
	policy := &workflowv1alpha1.ResourcePolicy{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"}}
	policy.Spec.Policy.Network = &workflowv1alpha1.NetworkAccessPolicy{}
	for _, pair := range pairs {
		policy.Spec.Policy.Network.AllowedPairs = append(policy.Spec.Policy.Network.AllowedPairs,
			workflowv1alpha1.NamespacePair{Namespace1: pair[0], Namespace2: pair[1]})
	}
	return policy
}

func testNetworkAccessController(t *testing.T, objects []runtime.Object, policies ...*workflowv1alpha1.ResourcePolicy) (*ResourcePolicyController, *fake.Clientset, cache.Indexer) {
	t.Helper()
	client := fake.NewSimpleClientset(objects...)
	informer := workflowinformers.NewSharedInformerFactory(workflowfake.NewSimpleClientset(), 0).Workflows().V1alpha1().ResourcePolicies()
	// Populate the cache directly so reconciliation tests need no running informer.
	for _, policy := range policies {
		if err := informer.Informer().GetIndexer().Add(policy); err != nil {
			t.Fatal(err)
		}
	}
	c := NewResourcePolicyController(client, informer)
	t.Cleanup(c.workqueue.ShutDown)
	return c, client, informer.Informer().GetIndexer()
}

func getNetworkPolicy(t *testing.T, client *fake.Clientset, namespace string) *networkingv1.NetworkPolicy {
	t.Helper()
	policy, err := client.NetworkingV1().NetworkPolicies(namespace).Get(context.Background(), crossNamespacePolicyName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func assertAllowed(t *testing.T, client *fake.Clientset, namespace string, peers ...string) {
	t.Helper()
	policy := getNetworkPolicy(t, client, namespace)
	// Compare complete rules so extra peers or port restrictions cannot go unnoticed.
	want := baselinePolicy(namespace).Spec.Ingress
	for _, peer := range sets.NewString(peers...).List() {
		want = append(want, namespaceAllowance(peer))
	}
	if !equality.Semantic.DeepEqual(policy.Spec.Ingress, want) {
		t.Fatalf("namespace %s: ingress = %#v, want local traffic and peers %v", namespace, policy.Spec.Ingress, peers)
	}
	if marked := policy.Labels[networkAccessManagedLabel] == "true"; marked != (len(peers) != 0) {
		t.Fatalf("namespace %s: unexpected management label %v", namespace, policy.Labels)
	}
}

func TestNetworkAccessWhitelistUpdates(t *testing.T) {
	first := whitelist("first", [2]string{"chatbot-a", "rag-a"}, [2]string{"rag-a", "chatbot-a"}, [2]string{"chatbot-a", "rag-b"})
	second := whitelist("second", [2]string{"chatbot-a", "rag-a"})
	legacy := &workflowv1alpha1.ResourcePolicy{ObjectMeta: metav1.ObjectMeta{Name: "legacy", Namespace: "default"}}
	unrelated := baselinePolicy("unrelated")
	unrelated.Spec.Ingress = append(unrelated.Spec.Ingress, namespaceAllowance("manual"))
	external := baselinePolicy("chatbot-a")
	external.Name = "allow-external-traffic"
	c, client, indexer := testNetworkAccessController(t, []runtime.Object{
		baselinePolicy("chatbot-a"), baselinePolicy("rag-a"), baselinePolicy("rag-b"), unrelated, external,
	}, first, second, legacy)
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, client, "chatbot-a", "rag-a", "rag-b")
	assertAllowed(t, client, "rag-a", "chatbot-a")
	assertAllowed(t, client, "rag-b", "chatbot-a")
	// Reversed pairs and overlapping policies must not produce duplicate rules or writes.
	client.ClearActions()
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	for _, action := range client.Actions() {
		if action.GetVerb() != "list" {
			t.Fatalf("unchanged whitelist caused an API write: %v", action)
		}
	}
	// Removing one policy keeps the pair still declared by the other.
	if err := indexer.Delete(first); err != nil {
		t.Fatal(err)
	}
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, client, "chatbot-a", "rag-a")
	assertAllowed(t, client, "rag-b")
	// Changing endpoints must revoke the old pair in both directions.
	if err := indexer.Update(whitelist("second", [2]string{"rag-a", "rag-b"})); err != nil {
		t.Fatal(err)
	}
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, client, "chatbot-a")
	assertAllowed(t, client, "rag-a", "rag-b")
	assertAllowed(t, client, "rag-b", "rag-a")
	if err := indexer.Delete(second); err != nil {
		t.Fatal(err)
	}
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	for _, namespace := range []string{"chatbot-a", "rag-a", "rag-b"} {
		assertAllowed(t, client, namespace)
	}
	if got := getNetworkPolicy(t, client, "unrelated"); !equality.Semantic.DeepEqual(got, unrelated) {
		t.Fatal("changed an unreferenced, unmanaged policy")
	}
	got, err := client.NetworkingV1().NetworkPolicies("chatbot-a").Get(context.Background(), external.Name, metav1.GetOptions{})
	if err != nil || !equality.Semantic.DeepEqual(got, external) {
		t.Fatal("changed allow-external-traffic")
	}
}

func TestNetworkAccessPreservesCustomRules(t *testing.T) {
	policy := baselinePolicy("chatbot-a")
	policy.Labels = map[string]string{"keep": "label"}
	policy.Annotations = map[string]string{"keep": "annotation"}
	policy.Spec.PodSelector.MatchLabels = map[string]string{"keep": "selector"}
	policy.Spec.PolicyTypes = []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress}
	policy.Spec.Egress = []networkingv1.NetworkPolicyEgressRule{{}}
	allow := namespaceAllowance("old")
	// Similar-looking custom rules must survive; only the exact tool format is replaced.
	portRestricted := *allow.DeepCopy()
	port := intstr.FromInt(443)
	portRestricted.Ports = []networkingv1.NetworkPolicyPort{{Port: &port}}
	podRestricted := *allow.DeepCopy()
	podRestricted.From[0].PodSelector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": "special"}}
	multiplePeers := *allow.DeepCopy()
	multiplePeers.From = append(multiplePeers.From, networkingv1.NetworkPolicyPeer{PodSelector: &metav1.LabelSelector{}})
	multipleNamespaces := *allow.DeepCopy()
	multipleNamespaces.From[0].NamespaceSelector.MatchExpressions[0].Values = []string{"one", "two"}
	policy.Spec.Ingress = append(policy.Spec.Ingress, portRestricted, podRestricted, multiplePeers, multipleNamespaces,
		networkingv1.NetworkPolicyIngressRule{From: []networkingv1.NetworkPolicyPeer{{IPBlock: &networkingv1.IPBlock{CIDR: "10.0.0.0/8"}}}},
		networkingv1.NetworkPolicyIngressRule{},
	)
	want := policy.DeepCopy()
	want.Labels[networkAccessManagedLabel] = "true"
	want.Spec.Ingress = append(want.Spec.Ingress, namespaceAllowance("new"))
	policy.Spec.Ingress = append(policy.Spec.Ingress, allow)
	before := policy.DeepCopy()
	got := networkPolicyWithAllowedNamespaces(policy, sets.NewString("new"))
	if !equality.Semantic.DeepEqual(got, want) {
		t.Fatal("failed to preserve custom rules or unrelated fields")
	}
	if !equality.Semantic.DeepEqual(before, policy) {
		t.Fatal("mutated the original object")
	}
	if !equality.Semantic.DeepEqual(got, networkPolicyWithAllowedNamespaces(got, sets.NewString("new"))) {
		t.Fatal("not idempotent")
	}
}

func TestNetworkAccessInvalidPairDoesNotBlockReconciliation(t *testing.T) {
	for _, invalidPair := range [][2]string{{"INVALID", "rag-a"}, {"chatbot-a", "INVALID"}, {"", "rag-a"}} {
		t.Run(invalidPair[0]+"-"+invalidPair[1], func(t *testing.T) {
			active := whitelist("access", [2]string{"chatbot-a", "rag-a"})
			c, client, indexer := testNetworkAccessController(t, []runtime.Object{
				baselinePolicy("chatbot-a"), baselinePolicy("rag-a"), baselinePolicy("rag-b"),
			}, active)
			if err := c.syncNetworkAccess(); err != nil {
				t.Fatal(err)
			}
			// An invalid pair must not prevent the valid replacement or old-pair cleanup.
			if err := indexer.Update(whitelist("access", invalidPair, [2]string{"chatbot-a", "rag-b"})); err != nil {
				t.Fatal(err)
			}
			if err := c.syncNetworkAccess(); err == nil {
				t.Fatal("expected a missing baseline policy error")
			}
			assertAllowed(t, client, "chatbot-a", "rag-b")
			assertAllowed(t, client, "rag-b", "chatbot-a")
			assertAllowed(t, client, "rag-a")
		})
	}
}

func TestNetworkAccessMissingBaselineDoesNotBlockCleanup(t *testing.T) {
	active := whitelist("active", [2]string{"chatbot-a", "rag-a"})
	c, client, indexer := testNetworkAccessController(t, []runtime.Object{
		baselinePolicy("chatbot-a"), baselinePolicy("rag-a"), baselinePolicy("chatbot-b"),
	}, active)
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	if err := indexer.Delete(active); err != nil {
		t.Fatal(err)
	}
	if err := indexer.Add(whitelist("pending", [2]string{"chatbot-b", "rag-b"})); err != nil {
		t.Fatal(err)
	}
	// Missing rag-b must neither grant half a pair nor block removal of the old pair.
	if err := c.syncNetworkAccess(); err == nil {
		t.Fatal("expected an error for the missing rag-b baseline policy")
	}
	assertAllowed(t, client, "chatbot-a")
	assertAllowed(t, client, "rag-a")
	assertAllowed(t, client, "chatbot-b")
	// Once the prerequisite exists, the next pass can apply the pending pair.
	if _, err := client.NetworkingV1().NetworkPolicies("rag-b").Create(context.Background(), baselinePolicy("rag-b"), metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, client, "chatbot-b", "rag-b")
	assertAllowed(t, client, "rag-b", "chatbot-b")
}

func TestNetworkAccessRepairsDriftAndCleansUp(t *testing.T) {
	policy := whitelist("access", [2]string{"chatbot-a", "rag-a"})
	c, client, indexer := testNetworkAccessController(t, []runtime.Object{
		baselinePolicy("chatbot-a"), baselinePolicy("rag-a"),
	}, policy)
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	// Simulate a manual CLI edit, then reconcile the declared whitelist again.
	drifted := getNetworkPolicy(t, client, "rag-a")
	drifted.Spec.Ingress = append(baselinePolicy("rag-a").Spec.Ingress, namespaceAllowance("manual"))
	if _, err := client.NetworkingV1().NetworkPolicies("rag-a").Update(context.Background(), drifted, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, client, "rag-a", "chatbot-a")
	// An empty whitelist revokes its permissions while keeping the ResourcePolicy.
	if err := indexer.Update(whitelist("access")); err != nil {
		t.Fatal(err)
	}
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, client, "chatbot-a")
	assertAllowed(t, client, "rag-a")
	if err := indexer.Update(policy); err != nil {
		t.Fatal(err)
	}
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	// After an offline deletion, a fresh controller has only the marked NetworkPolicies.
	c, client, _ = testNetworkAccessController(t, []runtime.Object{
		getNetworkPolicy(t, client, "chatbot-a"), getNetworkPolicy(t, client, "rag-a"),
	})
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, client, "chatbot-a")
	assertAllowed(t, client, "rag-a")
}

func TestNetworkAccessRecoversFromAPIErrors(t *testing.T) {
	c, client, _ := testNetworkAccessController(t, []runtime.Object{
		baselinePolicy("chatbot-a"), baselinePolicy("rag-a"),
	}, whitelist("access", [2]string{"chatbot-a", "rag-a"}))
	failed := false
	client.PrependReactor("update", "networkpolicies", func(action clienttesting.Action) (bool, runtime.Object, error) {
		if !failed {
			failed = true
			return true, nil, context.DeadlineExceeded
		}
		return false, nil, nil
	})
	if err := c.syncNetworkAccess(); err == nil {
		t.Fatal("expected the failed update to be reported")
	}
	// A later pass must finish the partially applied whitelist.
	if err := c.syncNetworkAccess(); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, client, "chatbot-a", "rag-a")
	assertAllowed(t, client, "rag-a", "chatbot-a")
}
