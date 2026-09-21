package main

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	platformworkflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
	platformlisters "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/listers/workflowcontroller/v1alpha1"
)

func TestBuildNetworkPolicy(t *testing.T) {
	rule := platformworkflowv1alpha1.NetworkAccessRule{
		Name: "chatbot-to-rag",
		ProviderRef: platformworkflowv1alpha1.ResourceReference{
			Kind: "RAGService", Namespace: "rag-services", Name: "shared-rag",
		},
		ServiceRef: platformworkflowv1alpha1.ResourceReference{
			Kind: "Service", Namespace: "rag-services", Name: "rag-api",
		},
		Consumers: []platformworkflowv1alpha1.ResourceReference{{
			Kind: "Chatbot", Namespace: "team-a", Name: "chatbot",
		}},
		Ports: []platformworkflowv1alpha1.NetworkPort{{Protocol: "TCP", Port: 8080}},
	}

	policy, err := buildNetworkPolicy(rule, corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "rag-services", Name: "rag-api"},
		Spec:       corev1.ServiceSpec{Selector: map[string]string{"partof": "ragservice-shared-rag"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if policy.Name != "chatbot-to-rag" || policy.Namespace != "rag-services" {
		t.Fatalf("unexpected policy identity: %s/%s", policy.Namespace, policy.Name)
	}
	if policy.Spec.PodSelector.MatchLabels["partof"] != "ragservice-shared-rag" {
		t.Fatalf("unexpected target selector: %#v", policy.Spec.PodSelector.MatchLabels)
	}
	if len(policy.Spec.Ingress) != 1 || len(policy.Spec.Ingress[0].From) != 1 || len(policy.Spec.Ingress[0].Ports) != 1 {
		t.Fatalf("unexpected ingress: %#v", policy.Spec.Ingress)
	}
	peer := policy.Spec.Ingress[0].From[0]
	if peer.NamespaceSelector.MatchLabels[networkPolicyNamespaceLabel] != "team-a" || peer.PodSelector.MatchLabels[networkPolicyPartOfLabel] != "chatbot-chatbot" {
		t.Fatalf("unexpected consumer selector: %#v", peer)
	}
	if policy.Spec.Ingress[0].Ports[0].Protocol == nil || *policy.Spec.Ingress[0].Ports[0].Protocol != corev1.ProtocolTCP || policy.Spec.Ingress[0].Ports[0].Port.IntVal != 8080 {
		t.Fatalf("unexpected port: %#v", policy.Spec.Ingress[0].Ports[0])
	}
	if policy.Labels[CREATED_BY_KEY] != CREATED_BY_VALUE {
		t.Fatalf("missing KubePlus ownership label: %#v", policy.Labels)
	}
}

func TestBuildNetworkPolicyRejectsInvalidPort(t *testing.T) {
	rule := platformworkflowv1alpha1.NetworkAccessRule{
		ProviderRef: platformworkflowv1alpha1.ResourceReference{Kind: "RAGService", Name: "shared-rag"},
		ServiceRef:  platformworkflowv1alpha1.ResourceReference{Name: "rag-api"},
		Consumers:   []platformworkflowv1alpha1.ResourceReference{{Kind: "Chatbot", Namespace: "team-a", Name: "chatbot"}},
		Ports:       []platformworkflowv1alpha1.NetworkPort{{Protocol: "TCP", Port: 0}},
	}
	_, err := buildNetworkPolicy(rule, corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "rag-services", Name: "rag-api"},
		Spec:       corev1.ServiceSpec{Selector: map[string]string{"partof": "ragservice-shared-rag"}},
	})
	if err == nil {
		t.Fatal("expected invalid port to be rejected")
	}
}

func TestSyncResourcePolicyUpsertsNetworkPolicy(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "rag-services", Name: "rag-api"},
		Spec:       corev1.ServiceSpec{Selector: map[string]string{"partof": "ragservice-shared-rag"}},
	}
	policy := &platformworkflowv1alpha1.ResourcePolicy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "provider", Name: "rag-access"},
		Spec: platformworkflowv1alpha1.ResourcePolicySpec{Policy: platformworkflowv1alpha1.Pol{
			Network: platformworkflowv1alpha1.NetworkAccessPolicy{Access: []platformworkflowv1alpha1.NetworkAccessRule{{
				Name:        "chatbot-to-rag",
				ProviderRef: platformworkflowv1alpha1.ResourceReference{Kind: "RAGService", Name: "shared-rag"},
				ServiceRef:  platformworkflowv1alpha1.ResourceReference{Kind: "Service", Namespace: "rag-services", Name: "rag-api"},
				Consumers:   []platformworkflowv1alpha1.ResourceReference{{Kind: "Chatbot", Namespace: "team-a", Name: "chatbot"}},
				Ports:       []platformworkflowv1alpha1.NetworkPort{{Protocol: "TCP", Port: 8080}},
			}}},
		}},
	}

	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	if err := indexer.Add(policy); err != nil {
		t.Fatal(err)
	}
	controller := &Controller{
		kubeclientset:          fake.NewSimpleClientset(service),
		resourcePoliciesLister: platformlisters.NewResourcePolicyLister(indexer),
	}

	if err := controller.syncHandler(resourcePolicyQueuePrefix + "provider/rag-access"); err != nil {
		t.Fatal(err)
	}
	policies := controller.kubeclientset.NetworkingV1().NetworkPolicies("rag-services")
	created, err := policies.Get(context.Background(), "chatbot-to-rag", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if created.Spec.Ingress[0].Ports[0].Port.IntVal != 8080 {
		t.Fatalf("unexpected created port: %d", created.Spec.Ingress[0].Ports[0].Port.IntVal)
	}

	policy.Spec.Policy.Network.Access[0].Ports[0].Port = 9090
	if err := indexer.Update(policy); err != nil {
		t.Fatal(err)
	}
	if err := controller.syncHandler(resourcePolicyQueuePrefix + "provider/rag-access"); err != nil {
		t.Fatal(err)
	}
	updated, err := policies.Get(context.Background(), "chatbot-to-rag", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Spec.Ingress[0].Ports[0].Port.IntVal != 9090 {
		t.Fatalf("unexpected updated port: %d", updated.Spec.Ingress[0].Ports[0].Port.IntVal)
	}
}
