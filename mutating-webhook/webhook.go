package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"context"
	"os"

	"github.com/buger/jsonparser"
	//guuid "github.com/google/uuid"

	"k8s.io/api/admission/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	//"k8s.io/kubernetes/pkg/apis/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	platformworkflowclientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/clientset/versioned"
	platformworkflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
)

type WebhookServer struct {
	server *http.Server
}


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

type ResourceComposition struct {
    CPU_limits    string
    Mem_limits string
    CPU_requests string
    Mem_requests string
    Name string
    Namespace string
    ChartName string
    ChartURL string
    Group string
    Kind string
    Plural string
    Version string
    Policy platformworkflowv1alpha1.Pol
}

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	accountidentity = "accountidentity"
	webhook_namespace = GetNamespace()

	annotations StoredAnnotations = StoredAnnotations{}
	customAPIPlatformWorkflowMap map[string]string
	customKindPluralMap map[string]string
	customAPIInstanceUIDMap map[string]string
	customAPIQuotaMap map[string]interface{}
	kindPluralMap map[string]string
	chartKindMap map[string]string
	resourcePolicyMap map[string]interface{}
	resourceNameObjMap map[string]interface{}
	namespaceHelmAnnotationMap map[string]string

	maxAllowedLength int
	maxLengthKind int

	resCompositions []ResourceComposition
)


func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
	annotations.KindToEntry = make(map[string][]Entry, 0)

	customAPIPlatformWorkflowMap = make(map[string]string,0)
	customAPIInstanceUIDMap = make(map[string]string,0)
	customKindPluralMap = make(map[string]string,0)
	customAPIQuotaMap = make(map[string]interface{}, 0)
	kindPluralMap = make(map[string]string,0)
	chartKindMap = make(map[string]string,0)
	resourcePolicyMap = make(map[string]interface{}, 0)
	resourceNameObjMap = make(map[string]interface{}, 0)
	namespaceHelmAnnotationMap = make(map[string]string, 0)

	maxAllowedLength = 30
	maxLengthKind = 28

	go setup()
}

func setup() {
	fmt.Println("Inside setup")

	var resCompositionsArr []byte
	var resString string
	for {
		resCompositionsArr = GetExistingResourceCompositions()
		resString = string(resCompositionsArr)
		fmt.Printf("ExistingResCompositions:%s\n", resString)
		if strings.Contains(resString, "connection refused") {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	fmt.Printf("ExistingResCompositions1:%s\n", resString)
	err := json.Unmarshal([]byte(resString), &resCompositions)
	if err != nil {
		fmt.Printf("Unmarshalling error:%s\n", err.Error())
	}
	fmt.Printf("Rescompositions:%v\n", resCompositions)

	for _, res := range resCompositions {
		name := res.Name
		ns := res.Namespace
		group := res.Group
		version := res.Version
		kind := res.Kind
		plural := res.Plural
		chartName := res.ChartName
		chartURL := res.ChartURL
		cpu_limits := res.CPU_limits
		mem_limits := res.Mem_limits
		cpu_requests := res.CPU_requests
		mem_requests := res.Mem_requests
		podPolicy := res.Policy

		customAPI := group + "/" + version + "/" + kind
		fmt.Printf("%s %s %s %s %s %s %s %s %s %s %s %s %s\n", name, ns, group, version, kind, plural, chartName, chartURL, cpu_limits, mem_limits, cpu_requests, mem_requests, customAPI)

		customAPIPlatformWorkflowMap[customAPI] = name
		customKindPluralMap[customAPI] = plural

		var quota_map map[string]string
        	quota_map = make(map[string]string)
        	quota_map["requests.cpu"] = cpu_requests
        	quota_map["limits.cpu"] = cpu_limits
        	quota_map["requests.memory"] = mem_requests
        	quota_map["limits.memory"] = mem_limits
        	customAPIQuotaMap[customAPI] = quota_map

                lowercaseKind := strings.ToLower(kind)
                kindPluralMap[lowercaseKind] = plural

	        customAPI1 := group + "/" + version + "/" + lowercaseKind
        	resourcePolicyMap[customAPI1] = podPolicy
	}
}

// main mutation process
func (whsvr *WebhookServer) mutate(ar *v1.AdmissionReview, httpMethod string) *v1.AdmissionResponse {
	req := ar.Request

	//fmt.Println("=== Request ===")
	//fmt.Println(req.Kind.Kind)
	//fmt.Println(req.Name)
	//fmt.Println(req.Namespace)
	//fmt.Println(httpMethod)
	//fmt.Printf("%s\n", string(req.Object.Raw))
	//fmt.Println("=== Request ===")

	//fmt.Println("=== User ===")
	//fmt.Println(req.UserInfo.Username)
	//fmt.Println("=== User ===")

	user := req.UserInfo.Username

	var patchOperations []patchOperation
	patchOperations = make([]patchOperation, 0)

	if httpMethod == http.MethodDelete {
		resp := handleDelete(ar)
		return resp
	}

	saveResource(ar)

	if req.Kind.Kind == "ResourcePolicy" {
		saveResourcePolicy(ar)
	}

	if req.Kind.Kind == "ResourceComposition"  {
		if strings.Contains(user, "kubeplus-saas-provider") {

			crdNameCheck := checkCRDNameValidity(ar)
			if crdNameCheck != "" {
                        	return &v1.AdmissionResponse{
                                	Result: &metav1.Status{
                                        Message: crdNameCheck,
                                	},
                        	}
			}

			message := checkChartExists(ar)
			if message != "" {
                        	return &v1.AdmissionResponse{
                                	Result: &metav1.Status{
                                        Message: message,
                                	},
                        	}
			}

			statusMessageToCheck := "New CRD defined in ResourceComposition created successfully."
			statusMessage := getStatusMessage(ar)
			if statusMessageToCheck == statusMessage {
				fmt.Printf("Intercepted call from platform-controller. Nothing to do for this call..\n")
                        	return &v1.AdmissionResponse{
					Allowed: true,
                        	}
			}

			errResponse := trackCustomAPIs(ar)
			if errResponse != nil {
				//fmt.Printf("111222333")
				return errResponse
			}
		} else {
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: "ResourceComposition instance can only be created by Provider.",
				},
			}
		}
	}
	var pacAnnotationMap map[string]string
	if req.Kind.Kind == "CustomResourceDefinition" {
		pacAnnotationMap = getPaCAnnotation(ar)
	}

	accountIdentityAnnotationMap := getAccountIdentityAnnotation(ar)
	allAnnotations := mergeMaps(pacAnnotationMap, accountIdentityAnnotationMap)
	annotationPatch := getAnnotationPatch(allAnnotations)

	patchOperations = append(patchOperations, annotationPatch)

	errResponse := handleCustomAPIs(ar, httpMethod)
	if errResponse != nil {
		return errResponse
	}

	/*
	// Deprecating support for Pod-level policies.
	if req.Kind.Kind == "Pod" {
		customAPI, rootKind, rootName, rootNamespace := checkServiceLevelPolicyApplicability(ar)
		var podResourcePatches []patchOperation
		if customAPI != "" && strings.Contains(customAPI, "platformapi.kubeplus") {
			podResourcePatches = applyPolicies(ar, customAPI, rootKind, rootName, rootNamespace)

                        // Add label if the Pod belongs to custom resource
                        labels := make(map[string]string, 0)
                        allLabels, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "labels")
                        if err == nil {
                                json.Unmarshal(allLabels, &labels)
                                //fmt.Printf("Pod all labels:%v\n", labels)
                        }
                        labels["partof"] = strings.ToLower(rootKind + "-" + rootName)
                        //fmt.Printf("All labels:%v\n", labels)

                        podLabelPatch := patchOperation{
                                Op:    "add",
                                Path:  "/metadata/labels",
                                Value: labels,
                        }
                        patchOperations = append(patchOperations, podLabelPatch)

		} else {
			// Check if Namespace-level policy is applicable.
			podResourcePatches = checkAndApplyNSPolicies(ar)
		}

		for _, podPatch := range podResourcePatches {
			patchOperations = append(patchOperations, podPatch)
		}
	}
	*/

	if req.Kind.Kind == "Namespace" {

		if strings.Contains(user, "kubeplus-saas-provider") || strings.Contains(user, "kubeplus-saas-consumer") {
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: "Permission denied: Namespace cannot be created.",
				},
			}
		}

		if !strings.Contains(webhook_namespace, "default") {
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: "Permission denied: Namespace can be created only if KubePlus is deployed in default Namespace.",
				},
			}
		}

		fmt.Printf("Recording Namespace...\n")
		releaseName := getReleaseName(ar)
		fmt.Printf("DEF Release name:%s\n", releaseName)
		if releaseName != "" {
			_, nsName, _ := getObjectDetails(ar)
			namespaceHelmAnnotationMap[nsName] = releaseName
		}
	}

	//fmt.Printf("PatchOperations:%v\n", patchOperations)
	patchBytes, _ := json.Marshal(patchOperations)
	//fmt.Printf("---------------------------------\n")
	// marshal the struct into bytes to pass into AdmissionResponse
	return &v1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func getStatusMessage(ar *v1.AdmissionReview) string {
	//fmt.Println("Inside getStatusMessage")
	req := ar.Request
        body := req.Object.Raw
        status, err := jsonparser.GetUnsafeString(body, "status","status")
        if err != nil {
                fmt.Errorf("Error:%s\n", err)
        }
	//fmt.Printf("getStatusMessage:%s\n", status)
	return status
}

func handleDelete(ar *v1.AdmissionReview) *v1.AdmissionResponse {
	fmt.Println("Inside handleDelete...")
	req := ar.Request
	//fmt.Printf("%v\n---",req)
	//body := req.Object.Raw
	namespace := req.Namespace
	resName := req.Name

	//fmt.Printf("Body:%v\n", body)
	apiv1 := req.Kind
	apiv2 := req.Resource
	//fmt.Printf("&&&&&\n")
	fmt.Printf("APIv1:%s, APIv2:%s\n", apiv1, apiv2)
	group := req.Resource.Group
	version := req.Resource.Version
	kind := req.Kind.Kind

	fmt.Printf("Group:%s, version:%s\n", group, version)
	plural := string(GetPlural(kind))
	apiVersion := group + "/" + version

	fmt.Printf("NS:%s, Kind:%s, apiVersion:%s, group:%s, version:%s plural:%s resName:%s\n",
		namespace, kind, apiVersion, group, version, plural, resName)

	fmt.Println("=== User ===")
	fmt.Println(req.UserInfo.Username)
	fmt.Println("=== User ===")

	user := req.UserInfo.Username

	if user == "system:serviceaccount:default:kubeplus" {
		return &v1.AdmissionResponse{
			Allowed: true,
		}
	}

	if kind == "ResourceComposition" {
		if !strings.Contains(user, "kubeplus-saas-provider") {
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: "ResourceComposition instance can only be deleted by Provider.",
				},
			}
		}
		// check status of helm releases for this CRD
		fmt.Println("Delete ResourceComposition")

		config, err := rest.InClusterConfig()
		if err != nil {
			fmt.Printf("Error:%s\n", err.Error())
		}

		var sampleclientset platformworkflowclientset.Interface
		sampleclientset, err = platformworkflowclientset.NewForConfig(config)
		if err != nil {
			fmt.Printf("Error:%s\n", err.Error())
		}

		platformWorkflow, err := sampleclientset.WorkflowsV1alpha1().ResourceCompositions(namespace).Get(context.Background(), resName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Error:%s\n", err.Error())
		}

		crdnamespace := platformWorkflow.ObjectMeta.Namespace
		fmt.Printf("Namespace: %s\n", crdnamespace)

		crdkind := platformWorkflow.Spec.NewResource.Resource.Kind
		crdgroup := platformWorkflow.Spec.NewResource.Resource.Group
		crdversion := platformWorkflow.Spec.NewResource.Resource.Version
		crdplural := platformWorkflow.Spec.NewResource.Resource.Plural
		fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s\n", crdkind, crdgroup, crdversion, crdplural)

		ownerRes := schema.GroupVersionResource{Group: crdgroup,
							Version: crdversion,
							Resource: crdplural}
		fmt.Printf("GVR: %v\n", ownerRes)


		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}

		crdObjList, err := dynamicClient.Resource(ownerRes).Namespace(crdnamespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error:%v\n...checking at cluster-scope", err)
			crdObjList, err = dynamicClient.Resource(ownerRes).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Error:%v\n", err)
			}
		}

		if crdObjList != nil {
			for _, instanceObj := range crdObjList.Items {
				objData := instanceObj.UnstructuredContent()
				status := objData["status"]
				labels := instanceObj.GetLabels()
				forcedDelete, _ := labels["delete"] 
				if status == nil && forcedDelete == "" {
					return &v1.AdmissionResponse{
						Result: &metav1.Status{
							Message: "Error: ResourceComposition instance cannot be deleted. It has an application instance starting up.",
						},
					}
				}
			}
		}

	}

	if kind == "Namespace" {
		if (strings.Contains(user, "kubeplus-saas-provider") || strings.Contains(user, "kubeplus-saas-consumer")) {
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: "Permission denied: Namespace cannot be deleted.",
				},
			}
		}
	}

	fmt.Printf("Calling DeleteCRDInstances...")
	resp := DeleteCRDInstances(kind, group, version, plural, namespace, resName)
	fmt.Println("After calling DeleteCRDInstances...")
	respString := string(resp)
	if strings.Contains(respString, "Error") {
		fmt.Println(respString)
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: respString,
			},
		}
	}

	return &v1.AdmissionResponse{
		Allowed: true,
	}
}

func checkAndApplyNSPolicies(ar *v1.AdmissionReview) []patchOperation {

	req := ar.Request

	// Get Pod's Namespace; Check if Namespace has Helm release annotation; 
	// Check if there is a Policy to be applied for that kind.
	// Assumption: A Given Kind name can exist in only one API group.

	// TODO:
	// Look up apiVersion + "/" + kind from arSaved
	// Use above to find out resourcePolicy from resourcePolicyMap
	// For applying polices - the spec comes from resourcePolicy, the values 
	// come from arSaved.

	podNamespace := req.Namespace
	//fmt.Printf("Pod Namespace:%s\n", podNamespace)
	releaseName := namespaceHelmAnnotationMap[podNamespace]
	//fmt.Printf("Release Name:%s\n", releaseName)

	customAPI := ""
	serviceKind := ""
	serviceInstance := ""
	var podPolicy interface{}

	if releaseName != "" {
		parts := strings.Split(releaseName, "-")
		if len(parts) > 0 {
			serviceKind = parts[0]
			serviceInstance = parts[1]
			for key, value := range resourcePolicyMap {
				parts1 := strings.Split(key, "/")
				if len(parts1) == 3 {
					targetKind := parts1[2]
					fmt.Printf("TargetKind:%s\n", targetKind)
					if targetKind == serviceKind {
						customAPI = key
						podPolicy = value
						fmt.Printf("Custom API for Policy application:%s\n", customAPI)
					}
				}
			}
		}
	}

	patchOperations := make([]patchOperation, 0)

	// Check if this is Namespace-scoped policy
	if podPolicy != nil {
		podPolicy1 := podPolicy.(platformworkflowv1alpha1.Pol)
		scope := podPolicy1.PolicyResources.Scope
		fmt.Printf("Scope:%s\n",scope)
		if scope == "Namespace" {
			resKindAndName := serviceKind + "-" + serviceInstance
			resAR := resourceNameObjMap[resKindAndName].(*v1.AdmissionReview)
			req := resAR.Request
			body := resAR.Request.Object.Raw

			serviceKindCanonical := req.Kind.Kind
			serviceKindNamespace, _ := jsonparser.GetUnsafeString(body, "metadata", "namespace")
			if serviceKindNamespace == "" {
				serviceKindNamespace = "default"
			}
			fmt.Printf("serviceKindCanonical:%s ServiceInstance:%s ServiceNamespace:%s\n", serviceKindCanonical, serviceInstance, serviceKindNamespace)
			patchOperations = applyPolicies(ar, customAPI, serviceKindCanonical, serviceInstance, serviceKindNamespace)
		}
	}

	return patchOperations
}

func getReleaseName(ar *v1.AdmissionReview) string {
	req := ar.Request
	annotations1 := make(map[string]string, 0)
	allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")

	if err == nil {
		json.Unmarshal(allAnnotations, &annotations1)
		//fmt.Printf("All Annotations:%v\n", annotations1)
	}

	for key, value := range annotations1 {
		if key == "meta.helm.sh/release-name" {
			releaseName := value
			fmt.Printf("ABC --- Release name:%s\n", releaseName)
			return releaseName
		}
	}
	return ""
}

func saveResource(ar *v1.AdmissionReview) {
	//fmt.Printf("Inside saveResource\n")
	kind, resName, _ := getObjectDetails(ar)
	//key := kind + "/" + namespace + "/" + resName
	key := kind + "-" + resName
	//fmt.Printf("Res Key:%s\n", key)
	_, ok := resourceNameObjMap[key]
	if !ok {
		resourceNameObjMap[key] = ar
	} 
	/*else {
		fmt.Printf("Key %s already present in resourceNameObjMap\n", key)
		//fmt.Printf("%v\n", val)
	}*/
}

func saveResourcePolicy(ar *v1.AdmissionReview) {
	req := ar.Request
	body := req.Object.Raw

	/*resPolicyName, _ := jsonparser.GetUnsafeString(body, "metadata", "name")
	fmt.Printf("Resource Policy Name:%s\n", resPolicyName)
	*/

	var resourcePolicy platformworkflowv1alpha1.ResourcePolicy
	json.Unmarshal(body, &resourcePolicy)
	/*if err != nil {
	    fmt.Println(err)	
	}*/

	kind := resourcePolicy.Spec.Resource.Kind
	lowercaseKind := strings.ToLower(kind)
	group := resourcePolicy.Spec.Resource.Group
	version := resourcePolicy.Spec.Resource.Version
	//fmt.Printf("Kind:%s, Group:%s, Version:%s\n", kind, group, version)

	podPolicy := resourcePolicy.Spec.Policy
	//fmt.Printf("Pod Policy:%v\n", podPolicy)

 	customAPI := group + "/" + version + "/" + lowercaseKind
 	resourcePolicyMap[customAPI] = podPolicy
 	//fmt.Printf("Resource Policy Map:%v\n", resourcePolicyMap)
}

func checkServiceLevelPolicyApplicability(ar *v1.AdmissionReview) (string, string, string, string) {
	//fmt.Printf("Inside checkServiceLevelPolicyApplicability")

	req := ar.Request
	body := req.Object.Raw

	//fmt.Printf("Body:%v\n", body)
	namespace := req.Namespace
	//fmt.Println("Namespace:%s\n",namespace)

	// TODO: looks like we can just keep one - namespace or namespace1
	namespace1, _, _, _ := jsonparser.Get(body, "metadata", "namespace")
	/*if err == nil {
		fmt.Printf("Namespace1:%s\n", namespace1)
	}*/

	ownerKind, _, _, err1 := jsonparser.Get(body, "metadata", "ownerReferences", "[0]", "kind")
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
	} /*else {
		fmt.Printf("ownerKind:%v\n", string(ownerKind))
	}*/

	ownerName, _, _, err1 := jsonparser.Get(body, "metadata", "ownerReferences", "[0]", "name")
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
	} /*else {
		fmt.Printf("ownerName:%v\n", string(ownerName))
	}*/

	ownerAPIVersion, _, _, err1 := jsonparser.Get(body, "metadata", "ownerReferences", "[0]", "apiVersion")
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
	} /*else {
		fmt.Printf("ownerAPIVersion:%v\n", string(ownerAPIVersion))
	}*/

	ownerKindS := string(ownerKind)
	ownerNameS := string(ownerName)
	ownerAPIVersionS := string(ownerAPIVersion)

	rootKind := ""
	rootName := ""
	rootAPIVersion := ""
	if ownerKindS == "" && ownerNameS == "" && ownerAPIVersionS == "" {
		annotations1 := make(map[string]string, 0)
		allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")
		if err == nil {
			json.Unmarshal(allAnnotations, &annotations1)
			//fmt.Printf("All Annotations:%v\n", annotations1)
		}
		releaseName := annotations1["meta.helm.sh/release-name"]
		//fmt.Printf("Helm release name:%s\n", releaseName)
		capiGroup := ""
		capiVersion := ""
        rootKind, rootName, capiGroup, capiVersion = getAPIDetailsFromHelmReleaseAnnotation(releaseName)
        rootAPIVersion = capiGroup + "/" + capiVersion
        //fmt.Printf("RK:%s, RN:%s, RAPI:%s\n", rootKind, rootName, rootAPIVersion)
	} else {
		rootKind, rootName, rootAPIVersion = findRoot(namespace, ownerKindS, ownerNameS, ownerAPIVersionS)
	}

	/*fmt.Printf("Root Kind:%s\n", rootKind)
	fmt.Printf("Root Name:%s\n", rootName)
	fmt.Printf("Root API Version:%s\n", rootAPIVersion)*/

	lowercaseKind := strings.ToLower(rootKind)

	// Check if the rootKind, rootName, rootAPIVersion is registered to be applied policies on.
 	customAPI := rootAPIVersion + "/" + lowercaseKind
 	//fmt.Printf("Custom API:%s\n", customAPI)
 	if _, ok := resourcePolicyMap[customAPI]; ok {
	 	//fmt.Printf("Resource Policy:%v\n", podPolicy)
	 	return customAPI, rootKind, rootName, string(namespace1)
 	}
	return "", "", "", ""
}

func getGroupVersion(apiVersion string) (string, string) {
	parts := strings.Split(apiVersion, "/")
	group := ""
	version := ""
	if len(parts) == 2 {
		group = parts[0]
		version = parts[1]
	} else {
		version = parts[0]
	}
	return group, version
}

func findRoot(namespace, kind, name, apiVersion string) (string, string, string) {
	rootKind := ""
	rootName := ""
	rootAPIVersion := ""

	time.Sleep(10)

	/*group, version := getGroupVersion(apiVersion)
	fmt.Printf("Group:%s\n", group)
	fmt.Printf("Version:%s\n", version)
	fmt.Printf("ResName:%s\n", name)
	fmt.Printf("Namespace:%s\n", namespace)
	*/
	ownerResKindPlural, _, ownerResApiVersion, ownerResGroup := getKindAPIDetails(kind)

	/*fmt.Printf("ownerResKindPlural:%s\n", ownerResKindPlural)
	fmt.Printf("ownerResApiVersion:%s\n", ownerResApiVersion)
	fmt.Printf("ownerResGroup:%s\n", ownerResGroup)
	*/
	ownerRes := schema.GroupVersionResource{Group: ownerResGroup,
									 		Version: ownerResApiVersion,
									   		Resource: ownerResKindPlural}

	//fmt.Printf("OwnerRes:%v\n", ownerRes)
	dynamicClient, err1 := getDynamicClient1()
	if err1 != nil {
		fmt.Printf("Error 1:%v\n", err1)
	    fmt.Println(err1)
		return rootKind, rootName, rootAPIVersion
	}
	instanceObj, err2 := dynamicClient.Resource(ownerRes).Namespace(namespace).Get(context.Background(),
																				   name,
																	   		 	   metav1.GetOptions{})
	if err2 != nil {
		//fmt.Printf("Error 2:%v\n", err2)
	    //fmt.Println(err2)
		return rootKind, rootName, rootAPIVersion
	}

	ownerReference := instanceObj.GetOwnerReferences()
	if len(ownerReference) == 0 {
		// Reached the root
		// Jump of from the Helm annotation; should be of type <plural>-<name>

		/*fmt.Printf("Intermediate Root kind:%s\n", kind)
		fmt.Printf("Intermediate Root name:%s\n", name)
		fmt.Printf("Intermediate Root APIVersion:%s\n", apiVersion)*/

		annotations := instanceObj.GetAnnotations()
		releaseName := annotations["meta.helm.sh/release-name"]

        capiKind, oinstance, capiGroup, capiVersion := getAPIDetailsFromHelmReleaseAnnotation(releaseName)
		return capiKind, oinstance, capiGroup + "/" + capiVersion
	} else {
		owner := ownerReference[0]
		ownerKind := owner.Kind
		ownerName := owner.Name
		ownerAPIVersion := owner.APIVersion
		rootKind, rootName, rootAPIVersion := findRoot(namespace, ownerKind, ownerName, ownerAPIVersion)
		return rootKind, rootName, rootAPIVersion
	}
}

func getAPIDetailsFromHelmReleaseAnnotation(releaseName string) (string, string, string, string) {
	capiKind := ""
	oinstance := ""
	capiGroup := ""
	capiVersion := ""

	parts := strings.Split(releaseName, "-")
	if len(parts) >= 2 {
		okindLowerCase := parts[0]
		if len(parts) > 2 {
			oinstance = strings.Join(parts[1:],"-")
		} else if len(parts) == 2 {
			oinstance = parts[1]
		}
		//fmt.Printf("KindPluralMap2:%v\n", kindPluralMap)
		oplural := kindPluralMap[okindLowerCase]
		//fmt.Printf("OPlural:%s OInstance:%s\n", oplural, oinstance)
		customAPI := ""
		for k, v := range customKindPluralMap {
			if v == oplural {
				customAPI = k
				break
			}
		}
		//fmt.Printf("CustomAPI:%s\n", customAPI)
		if customAPI != "" {
			capiParts := strings.Split(customAPI, "/")
			capiGroup = capiParts[0]
			capiVersion = capiParts[1]
			capiKind = capiParts[2]
			//fmt.Printf("capiGroup:%s capiVersion:%s capiKind:%s\n", capiGroup, capiVersion, capiKind)
		}
	} else {
		return "","","",""
	}
	return capiKind, oinstance, capiGroup, capiVersion
}

func applyPolicies(ar *v1.AdmissionReview, customAPI, rootKind, rootName, rootNamespace string) []patchOperation {
	//req := ar.Request
	//body := req.Object.Raw

	/*
	podName, _ := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
	fmt.Printf("Pod Name:%s\n", podName)*/

	/*
	res1, _, _, _ := jsonparser.Get(body, "spec", "containers")
	var containers map[string]any
	json.Unmarshal(res1, &containers)
	for key, value := range containers {
		mapval := value.(map[string]any)
		fmt.Printf("%v %v\n", key, mapval["name"])
	}*/

	// TODO: Defaulting to the first container. Take input for additional containers
	/*
	_, _, _, err1 := jsonparser.Get(body, "spec", "containers", "[0]", "resources")
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
	} */

	var operation string
	//fmt.Printf("DataType:%s\n", dataType)
	operation = "add"

	podPolicy := resourcePolicyMap[customAPI]
	//fmt.Printf("PodPolicy:%v\n", podPolicy)

	/*xType := fmt.Sprintf("%T", podPolicy)
	fmt.Printf("Pod Policy type:%s\n", xType) // "[]int"*/

	patchOperations := make([]patchOperation, 0)

	podPolicy1 := podPolicy.(platformworkflowv1alpha1.Pol)

	// 1. Requests
	cpuRequest := podPolicy1.PolicyResources.Requests.CPU
	memRequest := podPolicy1.PolicyResources.Requests.Memory
	if cpuRequest != "" && memRequest != "" {
		//fmt.Printf("CPU Request:%s\n", cpuRequest)
		if strings.Contains(cpuRequest, "values") {
			cpuRequest = getFieldValueFromInstance(cpuRequest,rootKind, rootName)
		}
		//fmt.Printf("CPU Request1:%s\n", cpuRequest)

		//fmt.Printf("Mem Request:%s\n", memRequest)
		if strings.Contains(memRequest, "values") {
			memRequest = getFieldValueFromInstance(memRequest,rootKind, rootName)
		}
		//fmt.Printf("Mem Request1:%s\n", memRequest)

		podResRequest := make(map[string]string,0)
		podResRequest["cpu"] = cpuRequest
		podResRequest["memory"] = memRequest

		//TODO: Defaulting to the first container. Take input for additional containers
		patch1 := patchOperation{
			Op:    operation,
			Path:  "/spec/containers/0/resources/requests",
			Value: podResRequest,
		}
		patchOperations = append(patchOperations, patch1)
	}

	// 2. Limits
	cpuLimit := podPolicy1.PolicyResources.Limits.CPU
	memLimit := podPolicy1.PolicyResources.Limits.Memory
	if cpuLimit != "" && memLimit != "" {
		//fmt.Printf("CPU Limit:%s\n", cpuLimit)
		//fmt.Printf("Mem Limit:%s\n", memLimit)

		podResLimits := make(map[string]string,0)
		podResLimits["cpu"] = cpuLimit
		podResLimits["memory"] = memLimit

		//TODO: Defaulting to the first container. Take input for additional containers
		patch2 := patchOperation{
			Op:    operation,
			Path:  "/spec/containers/0/resources/limits",
			Value: podResLimits,
		}
		patchOperations = append(patchOperations, patch2)
	}

	// 3. Node Selector
	nodeSelector := podPolicy1.PolicyResources.NodeSelector
	if nodeSelector != "" {
		//fmt.Printf("Node Selector:%s\n", nodeSelector)
		fieldValueS := getFieldValueFromInstance(nodeSelector, rootKind, rootName)
		if fieldValueS != "" {
			podNodeSelector := make(map[string]string,0)
			podNodeSelector["kubernetes.io/hostname"] = fieldValueS

			patch3 := patchOperation{
				Op:    operation,
				Path:  "/spec/nodeSelector",
				Value: podNodeSelector,
			}
			patchOperations = append(patchOperations, patch3)
		}
	}

	return patchOperations
}

func getFieldValueFromInstance(fieldName, rootKind, rootName string) string {
	parts := strings.Split(fieldName, ".")
	field := parts[1]
	field = strings.TrimSpace(field)
	//fmt.Printf("Field:%s\n", field)

		//kind, resName, namespace := getObjectDetails(ar)
	lowercaseRootKind := strings.ToLower(rootKind)
		//rootkey := lowercaseRootKind + "/" + rootNamespace + "/" + rootName
	rootkey := lowercaseRootKind + "-" + rootName
	//fmt.Printf("Root Key:%s\n", rootkey)
	/*fmt.Printf("Printing resourceNameObjMap -- \n")
	for key, value := range resourceNameObjMap {
        	fmt.Println(key, ":", value)
    	}
	fmt.Printf("--\n")*/

	val := resourceNameObjMap[rootkey]
	fieldValueS := ""
	if val != nil {
		arSaved := resourceNameObjMap[rootkey].(*v1.AdmissionReview)
		reqObject := arSaved.Request
		reqspec := reqObject.Object.Raw
		fieldValue, _, _, _ := jsonparser.Get(reqspec, "spec", field)
		fieldValueS = string(fieldValue)
	}
	/*if err2 != nil {
		fmt.Printf("Error:%v\n", err2)
	} else {
		fmt.Printf("Fields:%v\n", string(fieldValue))
	}*/
	return fieldValueS
}

func getObjectDetails(ar *v1.AdmissionReview) (string, string, string) {

	req := ar.Request
	body := req.Object.Raw

	kind := req.Kind.Kind
	lowercaseKind := strings.ToLower(kind)

	resName, _ := jsonparser.GetUnsafeString(body, "metadata", "name")

	namespace, _ := jsonparser.GetUnsafeString(body, "metadata", "namespace")

	if namespace == "" {
		namespace = "default"
	}
	return lowercaseKind, resName, namespace
}

func trackCustomAPIs(ar *v1.AdmissionReview) *v1.AdmissionResponse {
	fmt.Printf("Inside trackCustomAPIs...")
	req := ar.Request
	body := req.Object.Raw

	var platformWorkflow platformworkflowv1alpha1.ResourceComposition
	json.Unmarshal(body, &platformWorkflow)
	/*if err != nil {
	    fmt.Println(err)
	}*/

	platformWorkflowName, _ := jsonparser.GetUnsafeString(body, "metadata", "name")

	namespace1, _, _, _ := jsonparser.Get(body, "metadata", "namespace")
	namespace := string(namespace1)
	if namespace == "" {
		namespace = "default"
	}

	// Ensure that ResourceComposition is created in the same Namespace
	// where KubePlus is deployed.
	kubePlusNS := GetNamespace()
	if namespace != kubePlusNS {
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: "ResourceComposition instance should be created in the same Namespace as KubePlus Namespace (" + kubePlusNS + ")",
			},
		}
	}

	fmt.Printf("ResourceComposition:%s\n", platformWorkflowName)
	kind := platformWorkflow.Spec.NewResource.Resource.Kind
	group := platformWorkflow.Spec.NewResource.Resource.Group
	plural := platformWorkflow.Spec.NewResource.Resource.Plural
	version := platformWorkflow.Spec.NewResource.Resource.Version
	chartURL := platformWorkflow.Spec.NewResource.ChartURL
	chartName := platformWorkflow.Spec.NewResource.ChartName
	fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s, ChartURL:%s ChartName:%s\n", kind, group, version, plural, chartURL, chartName)

	if len(kind) > maxLengthKind {
		msg := fmt.Sprintf("Kind name cannot be more than %d characters. Consider changing Kind name.\n", maxLengthKind)
		fmt.Printf(msg + "\n")
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: msg,
			},
		}
	}

	customAPI := group + "/" + version + "/" + kind
	customAPIPlatformWorkflowMap[customAPI] = platformWorkflowName
	customKindPluralMap[customAPI] = plural

	// Ensure that the consumer Kind name does not already exist in the cluster.
	failed := checkResourceExists(kind, plural)
	if failed != "" {
		message := ""
		if failed == kind {
			message = "Resource with Kind Name " + kind + " exists in the cluster."
		}
		if failed == plural {
			message = "Resource with Plural Name " + plural + " exists in the cluster."
		}
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: message,
			},
		}
	}

	//lint_chart := os.Getenv("LINT_CHART")
	//if strings.EqualFold(lint_chart, "yes") {
		message1 := string(LintChart(chartURL))
		//fmt.Printf("After LintChart - message:%s\n", message1)
		if !strings.Contains(message1, "Chart is good") {
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: message1,
				},
			}
		}
	//}

	parts := strings.Split(message1, "\n")
	kindString := parts[1]
	//fmt.Printf("Kind string:%s\n", kindString)
	chartKindMap[platformWorkflowName] = kindString

	check_kyverno_policies := os.Getenv("CHECK_KYVERNO_POLICIES")
	if strings.EqualFold(check_kyverno_policies, "yes") {
		message := string(TestChartDeployment(kind, namespace, chartName, chartURL))
		fmt.Printf("After TestChartDeployment - message:%s\n", message)
		if strings.Contains(message, "Error") {
			fmt.Printf("99999999\n")
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: message,
				},
			}
		}
	}

	quota_cpu_requests, _ := jsonparser.GetUnsafeString(body, "spec","respolicy","spec","policy","quota","requests.cpu")
	//fmt.Printf("CPU requests(quota):%s\n", quota_cpu_requests)

	quota_memory_requests, _ := jsonparser.GetUnsafeString(body, "spec","respolicy","spec","policy","quota","requests.memory")
	//fmt.Printf("Memory requests(quota):%s\n", quota_memory_requests)

	quota_cpu_limits, _ := jsonparser.GetUnsafeString(body, "spec","respolicy","spec","policy","quota","limits.cpu")
	//fmt.Printf("CPU limits(quota):%s\n", quota_cpu_limits)

	quota_memory_limits, _ := jsonparser.GetUnsafeString(body, "spec","respolicy","spec","policy","quota","limits.memory")
	//fmt.Printf("Memory limits(quota):%s\n", quota_memory_limits)

	empty_quota_fields := 0
	if (quota_cpu_requests == "") {
		empty_quota_fields = empty_quota_fields + 1
	}

	if (quota_memory_requests == "") {
		empty_quota_fields = empty_quota_fields + 1
	}

	if (quota_cpu_limits == "") {
		empty_quota_fields = empty_quota_fields + 1
	}

	if (quota_memory_limits == "") {
		empty_quota_fields = empty_quota_fields + 1
	}

	if ( empty_quota_fields < 4 && empty_quota_fields > 0 ) {
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: "If quota is specified, specify all four values: requests.cpu, requests.memory, limits.cpu, limits.memory",
			},
		}
	}

	_, message2 := CheckClusterCapacity(quota_cpu_requests, quota_cpu_limits, quota_memory_requests, quota_memory_limits)
	//fmt.Printf("After CheckClusterCapacity - message:%s\n", message2)
	if !strings.Contains(message2, "Quota is within limits.") {
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: message2,
			},
		}
	}

	var quota_map map[string]string
	quota_map = make(map[string]string)
	quota_map["requests.cpu"] = quota_cpu_requests
	quota_map["limits.cpu"] = quota_cpu_limits
	quota_map["requests.memory"] = quota_memory_requests
	quota_map["limits.memory"] = quota_memory_limits

	customAPIQuotaMap[customAPI] = quota_map

	return nil
}

func checkResourceExists(kind, plural string) string {
	kindPlural := string(CheckResource(kind, plural))
	return kindPlural
}

func registerManPage(kind, apiVersion, platformworkflow, namespace string) string {
	lowercaseKind := strings.ToLower(kind)

    valuesYamlbytes := GetValuesYaml(platformworkflow, namespace)
    valuesYamlStr := string(valuesYamlbytes)
    /*introString := "Here is the values.yaml for the underlying Helm chart representing this resource.\n"
    introString = introString + "The attributes in values.yaml become the Spec properties of the resource.\n\n"
    valuesYaml = introString + valuesYaml*/

    /*valuesYamlDict := make(map[string]interface{})
    metadata := make(map[string]string)
    spec := make(map[string]interface{})
    valuesYamlDict["apiVersion"] = apiVersion
    valuesYamlDict["kind"] = kind
    metadata["name"] = "sample-" + lowercaseKind
    spec["spec"] = valuesYamlStr
    valuesYamlDict["metadata"] = metadata
    valuesYamlDict["spec"] = spec

    valuesYaml := fmt.Sprintf("%v", valuesYamlDict)*/

    prefix := "apiVersion: " + apiVersion + "\n"
    prefix = prefix + "kind: " + kind + "\n"
    prefix = prefix + "metadata:\n"
    prefix = prefix + "  name: sample-" + lowercaseKind + "\n"
    prefix = prefix + "spec:\n"

    lines := strings.Split(valuesYamlStr, "\n")
    for _, line := range lines {
	    prefix = prefix + "  " + line + "\n"
    }

    valuesYaml := prefix

    //fmt.Printf("Values YAML:%s\n",valuesYaml)

	configMapName := lowercaseKind + "-usage"

	cfg, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Error:%s\n", err.Error())
		return ""
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("Error:%s\n", err.Error())
		return ""
	}

    var yamlDataMap map[string]string
    yamlDataMap = make(map[string]string,0)
    yamlDataMap["spec"] = valuesYaml

	configMap := &corev1.ConfigMap{
		 ObjectMeta: metav1.ObjectMeta{
        	Name:      configMapName,
        	Namespace: namespace,
    	},
		Data: yamlDataMap,
	}

	_, err1 := kubeClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMap, metav1.CreateOptions{})

	if err1 != nil {
		fmt.Printf("Error:%s\n", err1.Error())
	} else {
		fmt.Println("Usage Config Map created:")
	}

	usageAnnotationValue := configMapName + ".spec"
	//fmt.Printf("Usage Annotation:%s\n", usageAnnotationValue)
	return usageAnnotationValue
}

func getPaCAnnotation(ar *v1.AdmissionReview) map[string]string {
	// Add crd annotation
	annotations1 := make(map[string]string, 0)

	req := ar.Request
	body := req.Object.Raw
	crdkind, _ := jsonparser.GetUnsafeString(body, "spec", "names", "kind")
	crdplural, _ := jsonparser.GetUnsafeString(body, "spec", "names", "plural")
	crdversion, _ := jsonparser.GetUnsafeString(body, "spec", "versions","[0]","name")
	crdgroup, _ := jsonparser.GetUnsafeString(body, "spec", "group")
	//fmt.Printf("CRDKind:%s, CRDPlural:%s, CRDVersion:%s\n", crdkind, crdplural, crdversion)
	customAPI := crdgroup + "/" + crdversion + "/" + crdkind
	apiVersion := crdgroup + "/" + crdversion
	platformWorkflowName, ok := customAPIPlatformWorkflowMap[customAPI]
	//fmt.Printf("PlatformWorkflowName:%s, ok:%v\n", platformWorkflowName, ok)
	chartKinds := ""
	if ok {

			namespace := GetNamespace()
			/**
	 		chartKindsB := DryRunChart(platformWorkflowName, namespace)
	 		chartKinds = string(chartKindsB)*/
			chartKinds = chartKindMap[platformWorkflowName]
	 		//fmt.Printf("Chart Kinds:%v\n", chartKinds)

	 		// If no kinds are found in the dry run then there is nothing to be done.
	 		if chartKinds == "" {
	 			return annotations1
	 		}

	 	//fmt.Printf("Annotating %s\n", chartKinds)
	 	parts := strings.Split(chartKinds, "-")
	 	uniqueKinds := make([]string,0)
	 	for _, p := range parts {
	 		found := false
	 		for _, u := range uniqueKinds {
	 			if p == u {
	 				found = true
	 			}
	 		}
	 		if !found {
	 			uniqueKinds = append(uniqueKinds, p)
	 		}
	 	}

	 	//fmt.Printf("Unique kinds:%v\n", uniqueKinds)
	 	chartKinds = strings.Join(uniqueKinds, ";")
	  	//fmt.Printf("Annotating %s\n", chartKinds)

		allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")
		if err == nil {
			json.Unmarshal(allAnnotations, &annotations1)
			//fmt.Printf("All Annotations:%v\n", annotations1)
		}
		annotateRel := "resource/annotation-relationship"
		lowercaseKind := strings.ToLower(crdkind)
		kindPluralMap[lowercaseKind] = crdplural
		//fmt.Printf("KindPluralMap1:%v\n", kindPluralMap)
	 	annotationValue := lowercaseKind + "-INSTANCE.metadata.name"
	 	//fmt.Printf("Annotation value:%s\n", annotationValue)

		annotateVal := "on:" + chartKinds + ", key:meta.helm.sh/release-name, value:" + annotationValue

		annotations1[annotateRel] = annotateVal

	 	namespace = GetNamespace()
	 	manpageConfigMapName := registerManPage(crdkind, apiVersion, platformWorkflowName, namespace)
	 	//fmt.Printf("### ManPage ConfigMap Name:%s ####\n", manpageConfigMapName)

	    manPageAnnotation := "resource/usage"
	    manPageAnnotationValue := manpageConfigMapName

	    annotations1[manPageAnnotation] = manPageAnnotationValue
 	}

	//fmt.Printf("All Annotations:%v\n", annotations1)

	return annotations1
}


func checkCRDNameValidity(ar *v1.AdmissionReview) string {
	//fmt.Printf("Inside checkCRDNameValidity...\n")

        req := ar.Request
        body := req.Object.Raw
        kind, err := jsonparser.GetUnsafeString(body, "spec","newResource","resource","kind")
        if err != nil {
                fmt.Errorf("Error:%s\n", err)
        }

	//crname, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
        //fmt.Printf("CR Name:%s\n", crname)
        //fmt.Printf("Kind:%s\n", kind)

	message1 := "";
	if strings.Contains(kind, ".") {
		message1 = "Kind name " + kind + " invalid. Cannot contain period (.)\n"
	}
		return message1
}


func checkChartExists(ar *v1.AdmissionReview) string {
	//fmt.Printf("Inside checkChartExists...\n")

        req := ar.Request
        body := req.Object.Raw
        _, err := jsonparser.GetUnsafeString(body, "kind")
        if err != nil {
                fmt.Errorf("Error:%s\n", err)
        }

	//_, err = jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
        //fmt.Printf("CR Name:%s\n", crname)
        //fmt.Printf("Kind:%s\n", kind)

	chartURL, err := jsonparser.GetUnsafeString(req.Object.Raw, "spec", "newResource", "chartURL")

	//fmt.Printf("%%%%% CHART URL:%s %%%%%\n", chartURL)

	message1 := string(CheckChartExists(chartURL))
        //fmt.Printf("After CheckChartExists - message:%s\n", message1)

	return message1
}


func handleCustomAPIs(ar *v1.AdmissionReview, httpMethod string) *v1.AdmissionResponse {
	req := ar.Request
	body := req.Object.Raw
	//fmt.Printf("%v\n", req)
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
	//fmt.Printf("Namespace:%s\n", namespace)
	crname, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")

	//cruid, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "uid")
	// We have to generate a uid as when the request is received there is no uid yet.
	// When the object is persisted Kubernetes will overwrite the uid with a new value - that is okay.
	//id := guuid.New()
	//cruid := id.String()
	//fmt.Printf("CR Uid:%s\n", cruid)

	labelsBytes, _, _, _ := jsonparser.Get(req.Object.Raw, "metadata", "labels")
	labels := string(labelsBytes)

	annotations1 := make(map[string]string, 0)
        allAnnotations, _, _, _ := jsonparser.Get(req.Object.Raw, "metadata", "annotations")
        json.Unmarshal(allAnnotations, &annotations1)
        sourceKind := ""
	sourceKindPlural := ""
        for key, value := range annotations1 {
                if key == "kubeplus/migrate-from" {
                        sourceKind = strings.TrimSpace(string(value))
			sourceKindPlural = string(GetPlural(sourceKind))
			sourceKindPlural = strings.TrimSpace(sourceKindPlural)
			fmt.Printf("Migration specificed from %s %s\n", sourceKind, sourceKindPlural)
			// Check nwhether the sourceKind exists in the cluster.
			if sourceKindPlural == "" {
				msg := fmt.Sprintf("Invalid kind specified in migrate-from annotation %s.\n", sourceKind)
				return &v1.AdmissionResponse{
					Result: &metav1.Status{
						Message: msg,
					},
				}
			}
                }
        }

	overridesBytes, _, _, _ := jsonparser.Get(req.Object.Raw, "spec")
	overrides := string(overridesBytes)

	nodeName, err := jsonparser.GetUnsafeString(req.Object.Raw, "spec", "nodeName")

	if nodeName != "" {
		fmt.Printf("nodeName in Spec:%s\n", nodeName)
		validNodeName := CheckApplicationNodeName(nodeName)
		if !validNodeName {
			msg := fmt.Sprintf("Invalid node name specified %s\n", nodeName)
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: msg,
				},
			}
		}
	}

	customAPI := apiVersion + "/" + kind
	//fmt.Printf("CustomAPI:%s\n", customAPI)

	// No longer being used - runs into concurrent map writes error (#1385)
	// Save name:uid mapping 
	//customAPIInstance := customAPI + "/" + namespace + "/" + crname
	//customAPIInstanceUIDMap[customAPIInstance] = cruid

	platformWorkflowName := customAPIPlatformWorkflowMap[customAPI]

	if platformWorkflowName != "" {
		fmt.Printf("=========\n")
		fmt.Printf("Inside handleCustomAPIs...\n")
		fmt.Printf("ResourceComposition:%s\n", platformWorkflowName)
		fmt.Printf("CR Name:%s\n", crname)
		fmt.Printf("Kind:%s\n", kind)
		fmt.Printf("labels:%s\n", labels)
		fmt.Printf("Overrides:%s\n", overrides)
		fmt.Printf("HTTP method:%v\n", httpMethod)

		// Check license
		if httpMethod == "CREATE" {
			license_ok := CheckLicense(kind, webhook_namespace)
			if license_ok != "" {
           		msg := fmt.Sprintf("%s Update license for %s and then re-try.\n", license_ok, kind)
           		return &v1.AdmissionResponse{
                		  Result: &metav1.Status{
                        		    Message: msg,
                  		},
           		}
			}
		}

		// Check if Namespace corresponding to crname is not in Terminating state
		config, err := rest.InClusterConfig()
		if err != nil {
                	fmt.Printf("Error:%s\n", err.Error())
			panic(err.Error())
		}

		kubeClient, err := kubernetes.NewForConfig(config)
        	if err != nil {
                	fmt.Printf("Error:%s\n", err.Error())
			panic(err.Error())
        	}

		nsObj, nsGetErr := kubeClient.CoreV1().Namespaces().Get(context.Background(), crname, metav1.GetOptions{})
		if nsGetErr == nil {
			nsPhase := nsObj.Status.Phase
			fmt.Printf("Namespace for %s exists. Current status is: %s\n", crname, nsPhase)
			if nsPhase == "Terminating" {
				msg := fmt.Sprintf("Previous Namespace for custom resource %s is in terminating state. Wait for it to terminate and then re-deploy\n", crname)
				return &v1.AdmissionResponse{
					Result: &metav1.Status{
						Message: msg,
					},
				}
			}
			nsAnnotations := nsObj.Annotations
			releaseNameAnnotation := ""
   			if nsAnnotations != nil {
        			fmt.Printf("Annotations of namespace %s:\n", crname)
        			for key, value := range nsAnnotations {
            				fmt.Printf("%s: %s\n", key, value)
					if key == "meta.helm.sh/release-name" {
						releaseNameAnnotation = value
					}
        			}
				if releaseNameAnnotation != "" {
					releaseNameAnnotation = strings.Replace(releaseNameAnnotation, `"`,"",-1)
					releaseNameAnnotation = strings.Replace(releaseNameAnnotation, `\`, "",-1)
					fmt.Printf("Release name annotation:%s\n", releaseNameAnnotation)
					prts := strings.Split(releaseNameAnnotation, "-")
					kindInRelease := ""
        				if len(prts) > 0 {
                				kindInRelease = prts[0]
						fmt.Printf("Kind in release:%s\n", kindInRelease)
						fmt.Printf("Kind in Rescomp:%s\n", kind)
						if strings.ToLower(kindInRelease) != strings.ToLower(kind) {
			                                msg := fmt.Sprintf("App with name %s already exists for Kind %s. Use different app name.\n", crname, kindInRelease)
                        			        return &v1.AdmissionResponse{
                                        				Result: &metav1.Status{
                                                			Message: msg,
                                        			},
                                			}
						}
        				}
				}
    			} else {
        			fmt.Println("No annotations found.")
    			}

		}

		lengthCheck := kind + "-" + crname
		if len(lengthCheck) > maxAllowedLength {
			kindLength := len(kind)
			// helm releases are named: <kind>-<instance>; maxAllowedLength is this combined length
			allowedInstanceLength := maxAllowedLength - kindLength - 1
			msg := fmt.Sprintf("Instance name exceeds the allowed limit of %d. Consider changing instance name.\n", allowedInstanceLength)
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: msg,
				},
			}
		}


		var sampleclientset platformworkflowclientset.Interface
		sampleclientset = platformworkflowclientset.NewForConfigOrDie(config)

		platformWorkflow1, err := sampleclientset.WorkflowsV1alpha1().ResourceCompositions(namespace).Get(context.Background(), platformWorkflowName, metav1.GetOptions{})
		//fmt.Printf("ResourceComposition:%v\n", platformWorkflow1)
		if err != nil {
			fmt.Errorf("Error:%s\n", err)
		}

		chartURL := platformWorkflow1.Spec.NewResource.ChartURL
		/*kind := platformWorkflow1.Spec.NewResource.Resource.Kind
		group := platformWorkflow1.Spec.NewResource.Resource.Group
		version := platformWorkflow1.Spec.NewResource.Resource.Version
		plural := platformWorkflow1.Spec.NewResource.Resource.Plural
		chartName := platformWorkflow1.Spec.NewResource.ChartName
 		//fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s, ChartURL:%s, ChartName:%s\n", kind, group, version, plural, chartURL, chartName)*/

		// If chart is local, check if it exists - it will not if KubePlus Pod has restarted due to cluster restart
                message1 := string(CheckChartExists(chartURL))
                //fmt.Printf("After CheckChartExists - message:%s\n", message1)
                if message1 != "" {
                        return &v1.AdmissionResponse{
                                Result: &metav1.Status{
                                        Message: message1,
                                },
                        }
                }

		cpu_requests_q := ""
		cpu_limits_q := ""
		mem_requests_q := ""
		mem_limits_q := ""
		if customAPIQuotaMap[customAPI] != nil {

		quota_map := customAPIQuotaMap[customAPI].(map[string]string)
		cpu_requests_q = quota_map["requests.cpu"]
		cpu_limits_q = quota_map["limits.cpu"]
		mem_requests_q = quota_map["requests.memory"]
		mem_limits_q = quota_map["limits.memory"]

		//fmt.Printf("cpu_req:%s cpu_lim:%s mem_req:%s mem_lim:%s\n", cpu_requests_q, cpu_limits_q, mem_requests_q, mem_limits_q)
	}

		//Save raw bytes of the request; We will create overrides in kubeconfiggenerator
	        //encodedOverrides := url.QueryEscape(overrides)
        	fp, _ := os.Create("/crdinstances/" + platformWorkflowName + "-" + crname + ".raw")
        	fp.Write(req.Object.Raw)
        	fp.Close()

		deploymentStatus := QueryDeployEndpoint(platformWorkflowName, crname, namespace, overrides, cpu_requests_q,
	                           cpu_limits_q, mem_requests_q, mem_limits_q, labels, sourceKind, sourceKindPlural)


		if string(deploymentStatus) != "" {
			msg := fmt.Sprintf("Error in deploying instance: %s\n", string(deploymentStatus))
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: msg,
				},
			}
		}


	}
	return nil
}

// Sets Owner Reference on an object after it has been created. 
// We initially started with this approach but are not using it anymore as the ResourceComposition instance is not technically
// owner of instances of the Custom API. Instead, we are using PaC annotation relationship
// to track this relation. The specific annotation that we look for is the Helm release name annotation.
func setOwnerReference(ar *v1.AdmissionReview) {
	req := ar.Request
	body := req.Object.Raw
	ckind, err := jsonparser.GetUnsafeString(body, "kind")
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
	capiVersion, err := jsonparser.GetUnsafeString(body, "apiVersion")
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
	}
	prts := strings.Split(capiVersion, "/")
	cgroup := ""
	cversion := prts[0]
	if len(prts) == 2 {
		cgroup = prts[0]
		cversion = prts[1]
	} 

	cname, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
	fmt.Printf("CR Name:%s\n", cname)
	cplural := string(GetPlural(ckind))
	fmt.Printf("Child Plural:%s\n",cplural)

	annotations1 := make(map[string]string, 0)
	allAnnotations, _, _, _ := jsonparser.Get(req.Object.Raw, "metadata", "annotations")
	json.Unmarshal(allAnnotations, &annotations1)
	helmReleaseName := ""
	helmReleaseNamespace := ""
	for key, value := range annotations1 {
		if key == "meta.helm.sh/release-name" {
			helmReleaseName = string(value)
		}
		if key == "meta.helm.sh/release-namespace" {
			helmReleaseNamespace = string(value)
		}
	}

	if helmReleaseName != "" && helmReleaseNamespace != "" {
	for registeredCustomResourceInstance, _ := range customAPIInstanceUIDMap {
		//uid := customAPIInstanceUIDMap[registeredCustomResourceInstance]
		parts := strings.Split(registeredCustomResourceInstance, "/")
		//fmt.Printf("Parts:%v\n", parts)
		if len(parts) >= 5 {
		apiVersion := parts[0] + "/" + parts[1]
		ogroup := parts[0]
		oversion := parts[1]
		okind := parts[2]
		namespace := parts[3]
		oinstance := parts[4]

		for customAPI, _ := range customAPIPlatformWorkflowMap {
			parts1 := strings.Split(customAPI, "/")
			customAPIKind := parts1[2]
			//fmt.Printf("Custom API Kind:%s\n", customAPIKind)

			if helmReleaseNamespace == namespace && helmReleaseName == oinstance && okind == customAPIKind {
					fmt.Printf("CR API Version:%s", apiVersion)
					fmt.Printf("CR API kind:%s", okind)
					fmt.Printf("CR API Namespace:%s", namespace)
					fmt.Printf("CR API Name:%s", oinstance)

					oplural := string(GetPlural(okind))
					fmt.Printf("OwnerPlural:%s\n",oplural)

					//if uid != "" && apiVersion != "" && kind != "" && name != "" {
					fmt.Printf("Setting owner reference...")

		 			go updateOwnerReference(cgroup, cversion, cplural, cname, ogroup, oversion, okind, oplural, oinstance, namespace)
					}
				}
			}
		}
	}
}

func updateOwnerReference(cgroup, cversion, cplural, cinstance, ogroup, oversion, okind, oplural, oinstance, namespace string) {

	fmt.Printf("Inside updateOwnerReference")
	cres := schema.GroupVersionResource{Group: cgroup,
									   Version: cversion,
									   Resource: cplural}
	fmt.Printf("CRes:%v\n", cres)

	ores := schema.GroupVersionResource{Group: ogroup,
									   Version: oversion,
									   Resource: oplural}
	fmt.Printf("ORes:%v\n", ores)

	dynamicClient, err1 := getDynamicClient1()
	if err1 != nil {
		fmt.Printf("Error in getting dynamic client:%v\n", err1)
		return
	}

	for {
		cobj, err1 := dynamicClient.Resource(cres).Namespace(namespace).Get(context.Background(), cinstance, metav1.GetOptions{})
		oobj, err2 := dynamicClient.Resource(ores).Namespace(namespace).Get(context.Background(), oinstance, metav1.GetOptions{})

		if err1 == nil && err2 == nil {
			oapiVersion := oobj.GetAPIVersion()
			ouid := oobj.GetUID()
			ref := metav1.OwnerReference{
				APIVersion: oapiVersion,
				Kind: okind,
				Name: oinstance, 
				UID: ouid,
			}
			refList := make([]metav1.OwnerReference, 0)
			refList = append(refList, ref)
			cobj.SetOwnerReferences(refList)
			dynamicClient.Resource(cres).Namespace(namespace).Update(context.Background(), cobj, metav1.UpdateOptions{})
			// break out of the for loop
			break 
		} else {
			time.Sleep(2 * time.Second)
		}
	}
	fmt.Printf("Done updating ownerReference of the CR kind:%s instance:%s\n", cplural, cinstance)
}

func mergeMaps(map1, map2 map[string]string) map[string]string {
	retmap := make(map[string]string,0)
	for k, v := range map1 {
		retmap[k] = v
	}
	for k, v := range map2 {
		retmap[k] = v
	}
	return retmap
}

func getAnnotationPatch(allAnnotations map[string]string) patchOperation {
	patch := patchOperation{
		Op:    "add",
		Path:  "/metadata/annotations",
		Value: allAnnotations,
	}
	return patch
}

func getAccountIdentityAnnotation(ar *v1.AdmissionReview) map[string]string {
	//fmt.Println("Inside getAccountIdentityAnnotation...")
	req := ar.Request

	/*
	kind := req.Kind.Kind
	name := req.Name
	namespace := req.Namespace

	fmt.Println(kind)
	fmt.Println(name)
	fmt.Println(namespace)
	*/

	// Add user identity annotation
	annotations1 := make(map[string]string, 0)
	allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")
	if err == nil {
		json.Unmarshal(allAnnotations, &annotations1)
		//fmt.Printf("All Annotations:%v\n", annotations1)
	}
	delete(annotations1, accountidentity)
	annotations1[accountidentity] = req.UserInfo.Username
	//fmt.Printf("All Annotations:%v\n", annotations1)
	return annotations1
}

// This method resolves binding functions - ImportValue, AddLabel, AddAnnotations in the Spec.
// Currently handling such Spec is turned off (there is no reference to this method in the main flow.)
// Leaving this method around for reference.
func getSpecResolvedPatch(ar *v1.AdmissionReview) ([]patchOperation, *v1.AdmissionResponse) {
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
				return patchOperations, &v1.AdmissionResponse{
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
				return patchOperations, &v1.AdmissionResponse{
					Result: &metav1.Status{
						Message: err.Error(),
					},
				}
			}
		}
		if resolveObj.FunctionType == AddAnnotation {
		 	fmt.Printf("Path to resolve:%s\n",resolveObj.JSONTreePath)
		 	fmt.Printf("Value to resolve:%s\n", resolveObj.Value)
			_, err := AddResourceAnnotation(resolveObj.Value)
			if err != nil {
				fmt.Printf("Could not add annotation to: %s", resolveObj.Value)
				return patchOperations, &v1.AdmissionResponse{
					Result: &metav1.Status{
						Message: err.Error(),
					},
				}
			}
		}
	}
	return patchOperations, nil
}

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
	//fmt.Print("## Received request ##")
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

	var admissionResponse *v1.AdmissionResponse
	ar := v1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		fmt.Printf("Can't decode body: %v", err)
		admissionResponse = &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		//fmt.Printf("%v\n", ar.Request)
		//fmt.Printf("####### METHOD:%s #######\n", ar.Request.Operation)
		//fmt.Println(r.URL.Path)
		if r.URL.Path == "/mutate" {
			method := string(ar.Request.Operation)
			admissionResponse = whsvr.mutate(&ar, method)
		}
	}

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
		//fmt.Println("Ready to write reponse ...")
		if _, err := w.Write(resp); err != nil {
			fmt.Printf("Can't write response: %v", err)
			http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		}
	}
}
