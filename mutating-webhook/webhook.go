package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/buger/jsonparser"

	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kubernetes/pkg/apis/core/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"

	platformworkflowclientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/clientset/versioned"
	platformworkflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"

)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)

	accountidentity = "accountidentity"
	accountidentities = "accountidentities"
	webhook_namespace = "default"
 
	kubeclientset *kubernetes.Clientset

	helper chan string

	customAPIPlatformWorkflowMap map[string]string
)

type WebhookServer struct {
	server *http.Server
	// client *kubernetes.ClientSet
}

var annotations StoredAnnotations = StoredAnnotations{}

// WhSvrParameters ...
// Webhook Server parameters
type WhSvrParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
	alsoLogToStderr bool 
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
type label struct {
	key   string
	value string
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
	annotations.KindToEntry = make(map[string][]Entry, 0)

	cfg, _ := rest.InClusterConfig()
	kubeclientset, _ = kubernetes.NewForConfig(cfg)

	customAPIPlatformWorkflowMap = make(map[string]string,0)

	//TODO: Helper that tracks resource creation and puts labels/annotations
	//once they are available.
	//helper = make(chan string)
	//go helperHandler()
}

func helperHandler() {
	for {
		val := <- helper
		fmt.Printf("..helperHandler value:%s\n", val)

		time.Sleep(10*time.Second)
	}
}

// main mutation process
func (whsvr *WebhookServer) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request

	fmt.Println("=== Request ===")
	fmt.Println(req.Kind.Kind)
	fmt.Println(req.Name)
	fmt.Println(req.Namespace)
	fmt.Println("=== Request ===")

	fmt.Println("=== User ===")
	fmt.Println(req.UserInfo.Username)
	fmt.Println("=== User ===")

	if req.Kind.Kind == "PlatformWorkflow" {
		trackCustomAPIs(ar)
	}

	handleCustomAPIs(ar)

	// TODO: Check if dependent resources have been created or not
	// checkDependency(ar)

	var patchOperations []patchOperation
	patchOperations = make([]patchOperation, 0)

	specResolvedPatches, errResponse := getSpecResolvedPatch(ar)
	if errResponse != nil {
		return errResponse
	}
	for _, specResolvedPatch := range specResolvedPatches {
		patchOperations = append(patchOperations, specResolvedPatch)
	}

	annotationPatch := getAccountIdentityAnnotationPatch(ar)
	patchOperations = append(patchOperations, annotationPatch)

	fmt.Printf("PatchOperations:%v\n", patchOperations)
	patchBytes, _ := json.Marshal(patchOperations)
	fmt.Printf("---------------------------------\n")
	// marshal the struct into bytes to pass into AdmissionResponse
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func trackCustomAPIs(ar *v1beta1.AdmissionReview) {
	req := ar.Request
	body := req.Object.Raw

	var platformWorkflow platformworkflowv1alpha1.PlatformWorkflow
	err := json.Unmarshal(body, &platformWorkflow)
	if err != nil {
	    fmt.Println(err)	
	}

	platformWorkflowName, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("PlatformWorkflow:%s\n", platformWorkflowName)
    customAPIs := platformWorkflow.Spec.CustomAPI
    for _, customAPI := range customAPIs {
    	kind := customAPI.Kind
    	group := customAPI.Group
    	version := customAPI.Version
    	plural := customAPI.Plural
    	chartURL := customAPI.ChartURL
 		fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s, ChartURL:%s\n", kind, group, version, plural, chartURL)
 		customAPI := group + "/" + version + "/" + kind
 		customAPIPlatformWorkflowMap[customAPI] = platformWorkflowName
    }
}

func handleCustomAPIs(ar *v1beta1.AdmissionReview) {
	fmt.Printf("Inside handleCustomAPIs...")
	req := ar.Request
	body := req.Object.Raw
	kind, err := jsonparser.GetUnsafeString(body, "kind")
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
	apiVersion, err := jsonparser.GetUnsafeString(body, "apiVersion")
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
	namespace := "default"
	ns, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "namespace")
	if ns != "" {
		namespace = ns
	}
	fmt.Printf("Namespace:%s\n", namespace)
	crname, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
	fmt.Printf("CR Name:%s\n", crname)

	overridesBytes, _, _, _ := jsonparser.Get(req.Object.Raw, "spec")
	overrides := string(overridesBytes)
	fmt.Printf("Overrides:%s\n", overrides)

	customAPI := apiVersion + "/" + kind
	fmt.Printf("CustomAPI:%s\n", customAPI)
	platformWorkflowName := customAPIPlatformWorkflowMap[customAPI]
	if platformWorkflowName != "" {
		fmt.Printf("PlatformWorkflow:%s\n", platformWorkflowName)

		config, err := rest.InClusterConfig()
	//	config, err := clientcmd.BuildConfigFromFlags("", "")
		if err != nil {
			panic(err.Error())
		}

		var sampleclientset platformworkflowclientset.Interface
		sampleclientset = platformworkflowclientset.NewForConfigOrDie(config)

		platformWorkflow1, err := sampleclientset.WorkflowsV1alpha1().PlatformWorkflows(namespace).Get(platformWorkflowName, metav1.GetOptions{})
		fmt.Printf("PlatformWorkflow:%v\n", platformWorkflow1)
		if err != nil {
			fmt.Errorf("Error:%s\n", err)
		}

	    customAPIs := platformWorkflow1.Spec.CustomAPI
    	for _, customAPI := range customAPIs {
    		kind := customAPI.Kind
    		group := customAPI.Group
    		version := customAPI.Version
    		plural := customAPI.Plural
    		chartURL := customAPI.ChartURL
    		chartName := customAPI.ChartName
 			fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s, ChartURL:%s, ChartName:%s\n", kind, group, version, plural, chartURL, chartName)
 			QueryDeployEndpoint(platformWorkflowName, crname, namespace, overrides)
    	}
	}
}

/*
func checkDependency(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {

	req := ar.Request
	kind := req.Kind.Kind
	name := req.Name
	namespace := req.Namespace

	//dependencyCreated, dependentElements := CheckDependency(kind, name, namespace, req.Object.Raw)
    dependencyCreated := false
	dependentElements := make([]string,0)
	fmt.Printf("DependencyCreated:%v, dependencyElements:%v\n", dependencyCreated, dependentElements)

	if !dependencyCreated {
		errorMessage := "Dependent Resources not created:\n"
		for _, elem := range dependentElements {
			depName := elem.Name
			depNamespace := elem.Namespace
			depKind := elem.Kind
			msg := fmt.Sprintf("   %s %s %s\n", depKind, depName, depNamespace)
			errorMessage = errorMessage + msg
		}
		fmt.Printf("Error:%s\n", errorMessage)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: errorMessage,
				},
			}
	}
	return nil
}
*/

func getAccountIdentityAnnotationPatch(ar *v1beta1.AdmissionReview) patchOperation {

	req := ar.Request

	kind := req.Kind.Kind
	name := req.Name
	namespace := req.Namespace

	fmt.Println(kind)
	fmt.Println(name)
	fmt.Println(namespace)

	// Add user identity annotation
	annotations1 := make(map[string]string, 0)
	allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")
	if err != nil {
		fmt.Printf("Error in parsing existing annotations")
	} else {
		json.Unmarshal(allAnnotations, &annotations1)
		fmt.Printf("All Annotations:%v\n", annotations1)
	}
	delete(annotations1, accountidentity)
	annotations1[accountidentity] = req.UserInfo.Username
	fmt.Printf("All Annotations:%v\n", annotations1)

	//updateConfigMap(req.UserInfo.Username)
	//userIdentity := map[string]string{"useridentity": req.UserInfo.Username}
	//annotations1 = append(annotations1, userIdentity)
	//userIdentityJSON, _ := json.Marshal(userIdentity)
	//allAnnotations = append(allAnnotations, userIdentityJSON)
	//fmt.Printf("All Annotations:%s", fmt.Sprintf("%v", annotations1))
	
	patch := patchOperation{
		Op:    "replace",
		Path:  "/metadata/annotations",
		Value: annotations1,
	}

	return patch
}

func getSpecResolvedPatch(ar *v1beta1.AdmissionReview) ([]patchOperation, *v1beta1.AdmissionResponse) {
	fmt.Printf("Inside getSpecResolvedPatch...\n")
	var patchOperations []patchOperation
	patchOperations = make([]patchOperation, 0)

	req := ar.Request
	forResolve := ParseRequest(req.Object.Raw)
	fmt.Printf("Objects To Resolve: %v\n", forResolve)
	for i := 0; i < len(forResolve); i++ {
		var resolveObj ResolveData
		resolveObj = forResolve[i]
		if resolveObj.FunctionType == ImportValue {
			importString := resolveObj.ImportString
			fmt.Printf("Import String: %s\n", importString)
			value, err := ResolveImportString(importString)
			fmt.Printf("ImportString:%s, Resolved ImportString value:%s", importString, value)
			if err != nil {
				return patchOperations, &v1beta1.AdmissionResponse{
					Result: &metav1.Status{
						Message: err.Error(),
					},
				}
			}
			patch := patchOperation{
				Op:    "replace",
				Path:  resolveObj.JSONTreePath,
				Value: value,
			}
			patchOperations = append(patchOperations, patch)
		}
		if resolveObj.FunctionType == AddLabel {
		 	fmt.Printf("Path to resolve:%s\n",resolveObj.JSONTreePath)
		 	fmt.Printf("Value to resolve:%s\n", resolveObj.Value)
			_, err := AddResourceLabel(resolveObj.Value)
			if err != nil {
				fmt.Printf("Could not add Label to: %s", resolveObj.Value)
				return patchOperations, &v1beta1.AdmissionResponse{
					Result: &metav1.Status{
						Message: err.Error(),
					},
				}
			}
			// TODO: Helper to put the label if the resource is not yet created.
		 	//helper <- resolveObj.Value
		}
		if resolveObj.FunctionType == AddAnnotation {
		 	fmt.Printf("Path to resolve:%s\n",resolveObj.JSONTreePath)
		 	fmt.Printf("Value to resolve:%s\n", resolveObj.Value)
			_, err := AddResourceAnnotation(resolveObj.Value)
			if err != nil {
				fmt.Printf("Could not add annotation to: %s", resolveObj.Value)
				return patchOperations, &v1beta1.AdmissionResponse{
					Result: &metav1.Status{
						Message: err.Error(),
					},
				}
			}
			// TODO: Helper to put the label if the resource is not yet created.
		 	//helper <- resolveObj.Value
		}
	}
	return patchOperations, nil
}

/*
func (whsvr *WebhookServer) mutate_prev(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	// Work-in-Progress: Below code is currently Work-in-progress hence
	// returning above.

	req := ar.Request

	var entry Entry
	var kind string
	name, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
	if err != nil {
		fmt.Printf("Error in parsing metadata.name. Key does not exist.")
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	namespace, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "namespace")
	if err != nil {
		fmt.Printf("Error in parsing metadata.namespace. Key does not exist.")
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	kind, err = jsonparser.GetUnsafeString(req.Object.Raw, "kind")
	if err != nil {
		fmt.Printf("Error in parsing kind. Key does not exist.")
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	fmt.Printf("Kind:%s, Name:%s, Namespace:%s\n", kind, name, namespace)
 
	if kind == "PlatformStack" {
		//UpdatePlatformStacks(name, namespace, req.Object.Raw)
	} else {
		//dependencyCreated, dependentElements := CheckDependency(kind, name, namespace, req.Object.Raw)
		dependencyCreated := false
		dependentElements := make([]string,0)
		fmt.Printf("DependencyCreated:%v, dependencyElements:%v\n", dependencyCreated, dependentElements)

		if !dependencyCreated {
			errorMessage := "Dependent Resources not created:\n"
			for _, elem := range dependentElements {
				depName := elem.Name
				depNamespace := elem.Namespace
				depKind := elem.Kind
				msg := fmt.Sprintf("   %s %s %s\n", depKind, depName, depNamespace)
				errorMessage = errorMessage + msg
			}
			fmt.Printf("Error:%s\n", errorMessage)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: errorMessage,
				},
			}
		}
	}

	allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")

	fmt.Println("--- Annotation Values: ---")
	jsonparser.ObjectEach(allAnnotations, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {

		val := strings.TrimSpace(string(value))
		fmt.Printf("--- value: %s\n", val)
		hasLabelFunc := strings.Contains(val, "Fn::AddLabel")

		if hasLabelFunc {
			_, err := AddResourceLabel(val)
			if err != nil {
				fmt.Printf("Could not add Label: %s", val)
			}
		}
		return nil
	})

	var patchOperations []patchOperation
	patchOperations = make([]patchOperation, 0)

	//fmt.Println("----- Stored data: -----")
	//fmt.Printf("Data: %v\n", annotations.KindToEntry)
	fmt.Printf("Api Request!: %s\n", string(req.Object.Raw))

	forResolve := ParseRequest(req.Object.Raw)

	fmt.Printf("Objects To Resolve: %v\n", forResolve)
	for i := 0; i < len(forResolve); i++ {
		var resolveObj ResolveData
		resolveObj = forResolve[i]

		// Skip processing if resolveObj is for metadata.annotations.
		// This is because we would have already processed that.
		partOfMetaData := strings.Contains(resolveObj.JSONTreePath, "/metadata/annotations")
		if partOfMetaData {
			continue
		}

		if resolveObj.FunctionType == ImportValue {

			importString := resolveObj.ImportString
			fmt.Printf("Import String: %s\n", importString)

			value, err := ResolveImportString(importString)
			fmt.Printf("ImportString:%s, Resolved ImportString value:%s", importString, value)
			if err != nil {
				// Because we could not resolve one of the Fn::<path>
				// we want to roll back our data structure by deleting the entry
				// we just added. The store and resolve should be an atomic
				// operation
				deleted := annotations.Delete(entry, kind)
				fmt.Printf("The data was deleted : %t", deleted)
				//fmt.Println(annotations)
				return &v1beta1.AdmissionResponse{
					Result: &metav1.Status{
						Message: err.Error(),
					},
				}
			}
			patch := patchOperation{
				Op:    "replace",
				Path:  resolveObj.JSONTreePath,
				Value: value,
			}
			patchOperations = append(patchOperations, patch)
		} else if resolveObj.FunctionType == AddLabel {

			patch := patchOperation{
				Op:    "replace",
				Path:  resolveObj.JSONTreePath,
				Value: resolveObj.Value,
			}
			patchOperations = append(patchOperations, patch)
		}

	}
	fmt.Printf("PatchOperations:%v\n", patchOperations)
	patchBytes, _ := json.Marshal(patchOperations)
	// marshal the struct into bytes to pass into AdmissionResponse
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func updateConfigMap(user string) {

	configMap, err1 := kubeclientset.CoreV1().ConfigMaps(webhook_namespace).Get(accountidentities, metav1.GetOptions{})
	if err1 != nil {
		fmt.Printf("ConfigMap Get Error:%s\n", err1.Error())
		userList := make([]string, 1)
		userList = append(userList, user)
		userListString := strings.Join(userList, ", ")
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: accountidentities,
			},
			Data: map[string]string{
				accountidentities: userListString,
			},
		}
		_, err2 := kubeclientset.CoreV1().ConfigMaps(webhook_namespace).Create(configMap)
		if err2 != nil {
			fmt.Printf("ConfigMap create Error:%s\n", err2.Error())
			return
		}
	} else {
		existingIdentitiesMap := configMap.Data
		userListString := existingIdentitiesMap[accountidentities]
		userList := strings.Split(userListString, ", ")
		present := false
		for _, u := range userList {
			if u == user {
				present = true
			}
		}
		if !present {
			userList = append(userList, user)
		}
		userListString = strings.Join(userList, ", ")

		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: accountidentities,
			},
			Data: map[string]string{
				accountidentities: userListString,
			},
		}
		_, err3 := kubeclientset.CoreV1().ConfigMaps(webhook_namespace).Update(configMap)
		if err3 != nil {
			fmt.Printf("ConfigMap update Error:%s\n", err3.Error())
			return
		}
	}
}
*/

func searchAnnotation(entries []Entry, instanceName, namespace, key string) (string, error) {
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if strings.EqualFold(entry.InstanceName, instanceName) &&
			strings.EqualFold(entry.Key, key) &&
			strings.EqualFold(entry.Namespace, namespace) {
			return entry.Value, nil
		}
	}
	// Could not find the data
	fmt.Printf("instance name: %s key: %s \n", instanceName, key)
	return "", fmt.Errorf("annotation data was not found. Check path")
}

// Serve method for webhook server
func (whsvr *WebhookServer) serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		fmt.Println("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		fmt.Printf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		fmt.Printf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		fmt.Println(r.URL.Path)
		if r.URL.Path == "/mutate" {
			admissionResponse = whsvr.mutate(&ar)
		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		fmt.Printf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	fmt.Println("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		fmt.Printf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
