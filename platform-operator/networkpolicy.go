package main

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"

	platformworkflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
)

const (
	networkPolicyProviderAnnotation = "kubeplus.io/provider-ref"
	networkPolicyPartOfLabel        = "partof"
	networkPolicyNamespaceLabel     = "kubernetes.io/metadata.name"
)

// builds one access rule; aggregate multiple rules when reconciliation is added.
func buildNetworkPolicy(rule platformworkflowv1alpha1.NetworkAccessRule, service corev1.Service) (*networkingv1.NetworkPolicy, error) {
	if rule.ProviderRef.Kind == "" || rule.ProviderRef.Name == "" {
		return nil, fmt.Errorf("providerRef kind and name are required")
	}
	if rule.ServiceRef.Name == "" {
		return nil, fmt.Errorf("serviceRef name is required")
	}
	if service.Name != rule.ServiceRef.Name {
		return nil, fmt.Errorf("serviceRef %q does not identify service %q", rule.ServiceRef.Name, service.Name)
	}
	if rule.ServiceRef.Namespace != "" && rule.ServiceRef.Namespace != service.Namespace {
		return nil, fmt.Errorf("serviceRef namespace %q does not match service namespace %q", rule.ServiceRef.Namespace, service.Namespace)
	}
	if len(service.Spec.Selector) == 0 {
		return nil, fmt.Errorf("service %q has no selector", service.Name)
	}
	if len(rule.Consumers) == 0 {
		return nil, fmt.Errorf("at least one consumer is required")
	}
	if len(rule.Ports) == 0 {
		return nil, fmt.Errorf("at least one port is required")
	}

	policyName := rule.Name
	if policyName == "" {
		policyName = "kubeplus-" + service.Name
	}
	if errs := validation.IsDNS1123Subdomain(policyName); len(errs) != 0 {
		return nil, fmt.Errorf("invalid network policy name %q: %s", policyName, strings.Join(errs, "; "))
	}

	from := make([]networkingv1.NetworkPolicyPeer, 0, len(rule.Consumers))
	for _, consumer := range rule.Consumers {
		if consumer.Kind == "" || consumer.Name == "" || consumer.Namespace == "" {
			return nil, fmt.Errorf("consumer kind, name, and namespace are required")
		}
		from = append(from, networkingv1.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
				networkPolicyNamespaceLabel: consumer.Namespace,
			}},
			PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
				networkPolicyPartOfLabel: strings.ToLower(consumer.Kind + "-" + consumer.Name),
			}},
		})
	}

	ports := make([]networkingv1.NetworkPolicyPort, 0, len(rule.Ports))
	for _, port := range rule.Ports {
		if port.Port < 1 || port.Port > 65535 {
			return nil, fmt.Errorf("port %d is outside the valid range 1-65535", port.Port)
		}
		policyPort := networkingv1.NetworkPolicyPort{Port: &intstr.IntOrString{Type: intstr.Int, IntVal: port.Port}}
		if port.Protocol != "" {
			protocol, err := networkProtocol(port.Protocol)
			if err != nil {
				return nil, err
			}
			policyPort.Protocol = &protocol
		}
		ports = append(ports, policyPort)
	}

	selector := make(map[string]string, len(service.Spec.Selector))
	for key, value := range service.Spec.Selector {
		selector[key] = value
	}

	return &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "NetworkPolicy"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyName,
			Namespace: service.Namespace,
			Labels: map[string]string{
				CREATED_BY_KEY: CREATED_BY_VALUE,
			},
			Annotations: map[string]string{
				networkPolicyProviderAnnotation: resourceReferenceString(rule.ProviderRef),
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: selector},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			Ingress: []networkingv1.NetworkPolicyIngressRule{{
				From:  from,
				Ports: ports,
			}},
		},
	}, nil
}

func networkProtocol(protocol string) (corev1.Protocol, error) {
	switch strings.ToUpper(protocol) {
	case string(corev1.ProtocolTCP):
		return corev1.ProtocolTCP, nil
	case string(corev1.ProtocolUDP):
		return corev1.ProtocolUDP, nil
	case string(corev1.ProtocolSCTP):
		return corev1.ProtocolSCTP, nil
	default:
		return "", fmt.Errorf("unsupported network protocol %q", protocol)
	}
}

func resourceReferenceString(ref platformworkflowv1alpha1.ResourceReference) string {
	return fmt.Sprintf("%s/%s/%s", ref.Namespace, ref.Kind, ref.Name)
}
