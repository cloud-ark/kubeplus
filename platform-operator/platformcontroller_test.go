package main

import (
	"context"
	"reflect"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	platformworkflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
	platformfake "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/clientset/versioned/fake"
	platformlisters "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/listers/workflowcontroller/v1alpha1"
)

func policyComposition() *platformworkflowv1alpha1.ResourceComposition {
	return &platformworkflowv1alpha1.ResourceComposition{
		ObjectMeta: metav1.ObjectMeta{Name: "c", Namespace: "provider", UID: "composition-uid"},
		Spec: platformworkflowv1alpha1.ResourceCompositionSpec{
			ResPolicy: platformworkflowv1alpha1.ResourcePolicy{
				ObjectMeta: metav1.ObjectMeta{Name: "p1"},
				Spec: platformworkflowv1alpha1.ResourcePolicySpec{Policy: platformworkflowv1alpha1.Pol{
					Network: platformworkflowv1alpha1.NetworkAccessPolicy{Access: []platformworkflowv1alpha1.NetworkAccessRule{{
						Name: "access", Consumers: []platformworkflowv1alpha1.ResourceReference{{Namespace: "namespace-a", Name: "app"}},
					}}},
				}},
			},
		},
	}
}

func TestSyncCompositionResourcePolicyLifecycle(t *testing.T) {
	ctx := context.Background()
	composition := policyComposition()
	standalone := composition.Spec.ResPolicy.DeepCopy()
	standalone.Name, standalone.Namespace = "p2", composition.Namespace
	clients := platformfake.NewSimpleClientset(standalone)
	controller := &Controller{platformStackclientset: clients}
	policies := clients.WorkflowsV1alpha1().ResourcePolicies(composition.Namespace)

	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	policy, err := policies.Get(ctx, "p1", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !metav1.IsControlledBy(policy, composition) || !reflect.DeepEqual(policy.Spec, composition.Spec.ResPolicy.Spec) {
		t.Fatalf("expected P1 owned by C with the nested spec: %#v", policy)
	}
	originalHash := policy.Annotations[resourcePolicyTemplateHashAnnotation]
	if originalHash == "" || composition.Spec.ResPolicy.Annotations != nil {
		t.Fatal("expected a persisted template hash without modifying the cached composition")
	}

	policy.Spec.Policy.Network.Access[0].Consumers = append(policy.Spec.Policy.Network.Access[0].Consumers,
		platformworkflowv1alpha1.ResourceReference{Namespace: "namespace-b", Name: "app"})
	policy.Labels = map[string]string{"user-label": "keep"}
	policy.Annotations["user-note"] = "keep"
	policy.Status.Status = "keep"
	if _, err := policies.Update(ctx, policy, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	composition.Spec.NewResource.ChartURL = "updated-chart"
	composition.Status.Status = "updated-status"
	// A new controller must recover the baseline from P1 rather than its own memory.
	controller = &Controller{platformStackclientset: clients}
	clients.ClearActions()
	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	for _, action := range clients.Actions() {
		if action.GetVerb() != "get" {
			t.Fatalf("unchanged nested policy should not write P1: %#v", action)
		}
	}
	preserved, err := policies.Get(ctx, "p1", metav1.GetOptions{})
	if err != nil || !reflect.DeepEqual(preserved.Spec, policy.Spec) {
		t.Fatalf("direct P1 edits were lost: %#v, %v", preserved, err)
	}

	composition.Spec.ResPolicy.Spec.Policy.Network.Access[0].Consumers = append(
		composition.Spec.ResPolicy.Spec.Policy.Network.Access[0].Consumers,
		platformworkflowv1alpha1.ResourceReference{Namespace: "namespace-c", Name: "app"})
	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	replaced, err := policies.Get(ctx, "p1", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(replaced.Spec, composition.Spec.ResPolicy.Spec) || replaced.Annotations[resourcePolicyTemplateHashAnnotation] == originalHash {
		t.Fatalf("expected the new nested spec and checkpoint to replace P1: %#v", replaced)
	}
	if replaced.Labels["user-label"] != "keep" || replaced.Annotations["user-note"] != "keep" || replaced.Status.Status != "keep" {
		t.Fatal("replacing P1's spec must preserve its other metadata and status")
	}

	if err := policies.Delete(ctx, "p1", metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	recreated, err := policies.Get(ctx, "p1", metav1.GetOptions{})
	if err != nil || !reflect.DeepEqual(recreated.Spec, composition.Spec.ResPolicy.Spec) || !metav1.IsControlledBy(recreated, composition) {
		t.Fatalf("expected P1 recreated from C's current policy: %#v, %v", recreated, err)
	}
	p2, err := policies.Get(ctx, "p2", metav1.GetOptions{})
	if err != nil || !reflect.DeepEqual(p2, standalone) {
		t.Fatalf("standalone P2 was changed: %#v, %v", p2, err)
	}
}

func TestSyncCompositionResourcePolicyAdoptsLegacyPolicy(t *testing.T) {
	composition := policyComposition()
	legacy := composition.Spec.ResPolicy.DeepCopy()
	legacy.Namespace = composition.Namespace
	legacy.Spec.Policy.Network.Access[0].Consumers[0].Namespace = "direct-edit"
	clients := platformfake.NewSimpleClientset(legacy)
	controller := &Controller{platformStackclientset: clients}
	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	adopted, err := clients.WorkflowsV1alpha1().ResourcePolicies(composition.Namespace).Get(context.Background(), "p1", metav1.GetOptions{})
	if err != nil || !reflect.DeepEqual(adopted.Spec, legacy.Spec) || !metav1.IsControlledBy(adopted, composition) || adopted.Annotations[resourcePolicyTemplateHashAnnotation] == "" {
		t.Fatalf("expected legacy P1 adopted without resetting its edits: %#v, %v", adopted, err)
	}
}

func TestSyncCompositionResourcePolicyRejectsForeignOwner(t *testing.T) {
	composition := policyComposition()
	policy := composition.Spec.ResPolicy.DeepCopy()
	policy.Namespace = composition.Namespace
	other := composition.DeepCopy()
	other.UID = "other-composition-uid"
	policy.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(other, platformworkflowv1alpha1.SchemeGroupVersion.WithKind("ResourceComposition"))}
	clients := platformfake.NewSimpleClientset(policy)
	controller := &Controller{platformStackclientset: clients}
	if err := controller.syncCompositionResourcePolicy(composition); err == nil {
		t.Fatal("expected a conflicting owner to be rejected")
	}
	for _, action := range clients.Actions() {
		if action.GetVerb() != "get" {
			t.Fatal("must not modify a policy owned by another composition")
		}
	}
}

func TestSyncCompositionResourcePolicyRetriesConflict(t *testing.T) {
	composition := policyComposition()
	clients := platformfake.NewSimpleClientset()
	controller := &Controller{platformStackclientset: clients}
	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	policies := clients.WorkflowsV1alpha1().ResourcePolicies(composition.Namespace)
	before, err := policies.Get(context.Background(), "p1", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	composition.Spec.ResPolicy.Spec.Policy.Network.Access[0].Consumers[0].Namespace = "namespace-c"
	conflict := true
	clients.PrependReactor("update", "resourcepolicies", func(action clienttesting.Action) (bool, runtime.Object, error) {
		if conflict {
			conflict = false
			return true, nil, apierrors.NewConflict(platformworkflowv1alpha1.Resource("resourcepolicies"), "p1", nil)
		}
		return false, nil, nil
	})
	if err := controller.syncCompositionResourcePolicy(composition); !apierrors.IsConflict(err) {
		t.Fatalf("expected conflict to reach the worker: %v", err)
	}
	after, err := policies.Get(context.Background(), "p1", metav1.GetOptions{})
	if err != nil || !reflect.DeepEqual(after, before) {
		t.Fatalf("failed update must not advance the checkpoint: %#v, %v", after, err)
	}
	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	after, err = policies.Get(context.Background(), "p1", metav1.GetOptions{})
	if err != nil || !reflect.DeepEqual(after.Spec, composition.Spec.ResPolicy.Spec) {
		t.Fatalf("retry did not apply C's new policy: %#v, %v", after, err)
	}
}

func TestHandleResourcePolicyDeletion(t *testing.T) {
	for _, name := range []string{"owned", "tombstone", "standalone", "stale-owner", "missing-composition", "deleting-composition"} {
		t.Run(name, func(t *testing.T) {
			composition := policyComposition()
			policy := composition.Spec.ResPolicy.DeepCopy()
			policy.Namespace = composition.Namespace
			policy.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(composition, platformworkflowv1alpha1.SchemeGroupVersion.WithKind("ResourceComposition"))}
			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			if name == "deleting-composition" {
				now := metav1.Now()
				composition.DeletionTimestamp = &now
			}
			if name != "missing-composition" {
				if err := indexer.Add(composition); err != nil {
					t.Fatal(err)
				}
			}
			if name == "standalone" {
				policy.OwnerReferences = nil
			} else if name == "stale-owner" {
				policy.OwnerReferences[0].UID = "old-composition-uid"
			}
			queue := workqueue.NewRateLimitingQueue(workqueue.NewItemFastSlowRateLimiter(0, 0, 1))
			defer queue.ShutDown()
			controller := &Controller{platformStacksLister: platformlisters.NewResourceCompositionLister(indexer), workqueue: queue}
			var obj interface{} = policy
			if name == "tombstone" {
				obj = cache.DeletedFinalStateUnknown{Key: "provider/p1", Obj: policy}
			}
			controller.handleObject(obj)
			want := 0
			if name == "owned" || name == "tombstone" {
				want = 1
			}
			if queue.Len() != want {
				t.Fatalf("queued %d compositions, want %d", queue.Len(), want)
			}
		})
	}
}

func TestSyncCompositionResourcePolicySkipsMissingTemplateOrDeletion(t *testing.T) {
	composition := policyComposition()
	clients := platformfake.NewSimpleClientset()
	controller := &Controller{platformStackclientset: clients}
	now := metav1.Now()
	composition.DeletionTimestamp = &now
	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	composition.DeletionTimestamp = nil
	composition.Spec.ResPolicy.Name = ""
	if err := controller.syncCompositionResourcePolicy(composition); err != nil {
		t.Fatal(err)
	}
	if len(clients.Actions()) != 0 {
		t.Fatal("must not create a policy without an active composition and named template")
	}
}
