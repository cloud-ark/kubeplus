package main

import (
	"fmt"
	//"os"
	//"flag"
	"time"
	"context"
	//"path/filepath"
	//"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"

	//admissionregistration "k8s.io/api/admissionregistration/v1beta1"
	admissionregistrationclientset "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
)

const (
	// Annotations to check on CRDs.
	CREATED_BY_KEY = "created-by"
	CREATED_BY_VALUE = "kubeplus"
)

func main() {

        cfg, err := rest.InClusterConfig()
        if err != nil {
                panic(err.Error())
        }

	crdClient, _ := apiextensionsclientset.NewForConfig(cfg)

	for {

		crds, _ := crdClient.CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})

		for _, crd := range crds.Items {
			//crdName := "moodles.moodlecontroller.kubeplus"
			crdName := crd.Name
			crdObj, err := crdClient.CustomResourceDefinitions().Get(context.Background(), crdName, metav1.GetOptions{})
			annotations := crdObj.ObjectMeta.Annotations
			val, ok := annotations[CREATED_BY_KEY]
			if !ok || val != CREATED_BY_VALUE {
				continue;
			}

			resourceAPIGroup := crdObj.Spec.Group
			fmt.Printf("Custom Resource API Group:%s\n", resourceAPIGroup)

			resourcePlural := crdObj.Spec.Names.Plural
			fmt.Printf("Custom Resource Plural:%s\n", resourcePlural)

			resourceAPIVersion := crdObj.Spec.Versions[0].Name

			if err != nil {
				fmt.Errorf("Error:%s\n", err)
			}
			//fmt.Printf("CRD Object:%v\n", crdObj)

			//fmt.Println("====================================")

			admissionRegClient, _ := admissionregistrationclientset.NewForConfig(cfg)
			mutatingWebhookConfigName := "platform-as-code.crd-binding"
			mutatingWebhookObj, err := admissionRegClient.MutatingWebhookConfigurations().Get(context.Background(), mutatingWebhookConfigName, metav1.GetOptions{})

			if err != nil {
				fmt.Errorf("Error:%s\n", err)
			}
			fmt.Printf("MutatingWebhook Object:%v\n", mutatingWebhookObj)

			fmt.Println("============= Updating the Rules ==============")

			webhooks := mutatingWebhookObj.Webhooks

			if len(webhooks) == 1 {

				webhook := webhooks[0]
				rules := webhook.Rules
				currentRule := rules[0]

				if !checkExists(currentRule.APIGroups, resourceAPIGroup) {
					currentRule.APIGroups = append(currentRule.APIGroups, resourceAPIGroup)
				}
				if !checkExists(currentRule.Resources, resourcePlural) {
					currentRule.Resources = append(currentRule.Resources, resourcePlural)
				}
				if !checkExists(currentRule.APIVersions, resourceAPIVersion) {
					currentRule.APIVersions = append(currentRule.APIVersions, resourceAPIVersion)
				}

				rules[0] = currentRule
				webhook.Rules = rules
				webhooks[0] = webhook
				mutatingWebhookObj.Webhooks = webhooks

				fmt.Printf("MutatingWebhook Object:%v\n", mutatingWebhookObj)

				_, err = admissionRegClient.MutatingWebhookConfigurations().Update(context.Background(), mutatingWebhookObj, metav1.UpdateOptions{})
				if err != nil {
					fmt.Errorf("Error1:%s\n", err)
				}
		   }
		}
	}
	time.Sleep(1 * time.Second)
}

func checkExists(rule []string, s string) bool {
	for _, r := range rule {
		if r == s {
			return true
		}
	}
	return false
}

func extra() {

		/*
		var kubeconfig *string
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	*/

	// use the current context in kubeconfig
	//cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)


/*
	for _, webhook := range webhooks {
		name := webhook.Name
		fmt.Printf("Webhook Name:%s\n", name)
		rules := webhook.Rules
		currentRule := rules[0]
		currentRule.APIGroups = append(currentRule.APIGroups, "abc.def")
		currentRule.Resources = append(currentRule.Resources, "abc.def.ghi")
		rules[0] = currentRule
		webhook.Rules = rules

		for _, rule := range webhook.Rules {
			actualRule := rule.Rule
			fmt.Printf("Actual Rule:%s\n", actualRule)
			apiGroups := actualRule.APIGroups
			apiVersions := actualRule.APIVersions
			resources := actualRule.Resources
			fmt.Printf("APIGroups:%s, APIVersions:%s, Resources:%s\n", apiGroups, apiVersions, resources)
		}
	}
*/

}
