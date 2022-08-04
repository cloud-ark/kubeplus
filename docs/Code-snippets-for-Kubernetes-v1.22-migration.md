Code snippets for migrating to Kubernetes v1.22+
=================================================

Recently we have been working on migrating KubePlus to the latest version of Kubernetes (1.24+). This has been an involved process, as several key APIs and procedures that KubePlus depended on have been graduated or changed. These include, APIs and steps related to creating Kubernetes Mutating Webhook, APIs related to working with Kubernetes Custom Resource Definitions (CRDs), steps to create clients, listers, informers for your CRDs, and steps involved in creating ServiceAccount secret tokens. 

While broad-level suggestions about this migration were already [published on Kubernetes blog](https://kubernetes.io/blog/2021/07/14/upcoming-changes-in-kubernetes-1-22/), we realized that when it came to actual code-level details, these suggestions were not sufficient. After going through several Github issues, Kubernetes code, and StackOverflow posts, we felt that a document that shows code snippets for each of the above problems can be useful for others who are going through this migration. Hopefully, this saves you some time as you work on your migration. 


Q1. How to register a CustomResourceDefinition (CRD) object?

A. The main issue in registring a CRD object through code is how to create the OpenAPISchema definition, which is compulsory in Kubernetes versions 1.22+.

Suppose you want to register a CRD that has following spec and status fields:
```
apiVersion: platformapi.kubeplus/v1alpha1
kind: HelloWorldService
metadata:
  name: hs1
spec:
  greeting: Hello hello hello
  language: English
  country: USA
status:
  fullgreeting: Greeting is Hello hello hello 
```

Here is Golang code to achieve that.


```
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
    apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
    "k8s.io/client-go/rest"
)


 func() createCRD() {

        cfg, err := rest.InClusterConfig()
        if err != nil {
                panic(err.Error())
        }

        crdClient, _ := apiextensionsclientset.NewForConfig(cfg)

        properties := [3]string{"greeting", "language", "country"}

        specProperties := make(map[string]apiextensionsv1.JSONSchemaProps)
        for i:=0; i<len(properties); i++ {
                fieldName := properties[i]
                specProperties[fieldName] = apiextensionsv1.JSONSchemaProps{Type: "string"}
        }
        var jsonSchemaInner apiextensionsv1.JSONSchemaProps
        jsonSchemaInner.Type = "object"
        jsonSchemaInner.Properties = specProperties

        statusProperties := make(map[string]apiextensionsv1.JSONSchemaProps)
        statusProperties["fullgreeting"] = apiextensionsv1.JSONSchemaProps{Type: "string"}
        var jsonSchemaStatus apiextensionsv1.JSONSchemaProps
        jsonSchemaStatus.Type = "object"
        jsonSchemaStatus.Properties = statusProperties

        specField := make(map[string]apiextensionsv1.JSONSchemaProps)
        specField["spec"] = jsonSchemaInner
        specField["status"] = jsonSchemaStatus
        var jsonSchemaOuter apiextensionsv1.JSONSchemaProps
        jsonSchemaOuter.Type = "object"
        jsonSchemaOuter.Properties = specField

        crd := &apiextensionsv1.CustomResourceDefinition{
                ObjectMeta: metav1.ObjectMeta{
                        Name: "helloworldservices.platformapi.kubeplus",
                },
                Spec: apiextensionsv1.CustomResourceDefinitionSpec{
                        Group: "platformapi.kubeplus",
                        Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
                                {
                                        Name: "v1alpha1",
                                        Served: true,
                                        Storage: true,
                                        Schema: &apiextensionsv1.CustomResourceValidation{OpenAPIV3Schema: &jsonSchemaOuter},
                                },
                        },
                        Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
                                Plural: "helloworldservices",
                                Kind: "HelloWorldService",
                        },
                        Scope: "Namespaced",
                },
        }
         _, err1 := crdClient.CustomResourceDefinitions().Create(context.Background(), crd, metav1.CreateOptions{})
  }
```

Q2. How create AdmissionReview response in your mutating webhook?

A. The main thing to note here is that the AdmissionReview response object requires following three fields set - Kind (AdmissionReview), APIVersion (``admission.k8s.io/v1``), Response.UID.


```
import (
    "fmt"
	"net/http"
    "strings"
    "io/ioutil"
    "k8s.io/api/admission/v1"
    admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1"
    "k8s.io/apimachinery/pkg/runtime/serializer"
)

func() mutate(ar *v1.AdmissionReview, httpMethod string) *v1.AdmissionResponse {
        return &v1.AdmissionResponse{
                Allowed: true,
        }
}

func() serve(w http.ResponseWriter, r *http.Request) {

		var body []byte
        if r.Body != nil {
                if data, err := ioutil.ReadAll(r.Body); err == nil {
                        body = data
                }
        }

		ar := v1.AdmissionReview{}
		deserializer.Decode(body, nil, &ar)
		method := string(ar.Request.Operation)

		admissionResponse := mutate(&ar, method)
        admissionReview := v1.AdmissionReview{}
        if admissionResponse != nil {
                admissionReview.Response = admissionResponse
                if ar.Request != nil {
                        admissionReview.Response.UID = ar.Request.UID
                        admissionReview.APIVersion = "admission.k8s.io/v1"
                        admissionReview.Kind = "AdmissionReview"
                }
                resp, err := json.Marshal(admissionReview)
                if err != nil {
                        fmt.Printf("Can't encode response: %v", err)
                        http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
                }
                if _, err := w.Write(resp); err != nil {
                        fmt.Printf("Can't write response: %v", err)
                        http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
                }
        }
}
```

Q3. How to register a mutating webhook?

A. Check [sample-mutatingwebhookserver](https://github.com/cloud-ark/sample-mutatingwebhook).


Q4. How to create client, lister, informer for your CRDs?

A. Check the steps [here](https://github.com/cloud-ark/kubeplus/issues/14#issuecomment-1197339771).


Q5. How to create a ServiceAccount secret token?

A. The main thing to note here is that a Secret object is no longer created by default for ServiceAccount. We have to create the Secret ourselves. First create a ServiceAccount, then create a secret with ServiceAccount name set as an annotation on the Secret. And make sure that you set the Secret type to ``kubernetes.io/service-account-token``. The Secret will be populated with the token that you can retrieve using ``kubectl describe secret``.

```
$ kubectl create sa abc
$ cat abc-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret-sa-sample
  annotations:
    kubernetes.io/service-account.name: "abc"
type: kubernetes.io/service-account-token
$ kubect create secret -f abc-secret.yaml

```

