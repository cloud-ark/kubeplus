package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	guuid "github.com/google/uuid"

	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kubernetes/pkg/apis/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

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
	customKindPluralMap map[string]string
	customAPIInstanceUIDMap map[string]string
	kindPluralMap map[string]string
	resourcePolicyMap map[string]interface{}
	resourceNameObjMap map[string]interface{}
	namespaceHelmAnnotationMap map[string]string
	kindReqMap map[string]interface{}

)

type WebhookServer struct {
	server *http.Server
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
	customAPIInstanceUIDMap = make(map[string]string,0)
	customKindPluralMap = make(map[string]string,0)
	kindPluralMap = make(map[string]string,0)
	resourcePolicyMap = make(map[string]interface{}, 0)
	resourceNameObjMap = make(map[string]interface{}, 0)
	namespaceHelmAnnotationMap = make(map[string]string, 0)
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

	saveResource(ar)

	var patchOperations []patchOperation
	patchOperations = make([]patchOperation, 0)

	if req.Kind.Kind == "ResourcePolicy" {
		saveResourcePolicy(ar)
	}

	if req.Kind.Kind == "ResourceComposition" {
		trackCustomAPIs(ar)
	}

	var pacAnnotationMap map[string]string
	if req.Kind.Kind == "CustomResourceDefinition" {
		pacAnnotationMap = getPaCAnnotation(ar)
	}

	accountIdentityAnnotationMap := getAccountIdentityAnnotation(ar)
	allAnnotations := mergeMaps(pacAnnotationMap, accountIdentityAnnotationMap)
	annotationPatch := getAnnotationPatch(allAnnotations)

	patchOperations = append(patchOperations, annotationPatch)

	errResponse := handleCustomAPIs(ar)
	if errResponse != nil {
		return errResponse
	}

	if req.Kind.Kind == "Pod" {
		customAPI, rootKind, rootName, rootNamespace := checkServiceLevelPolicyApplicability(ar)
		var podResourcePatches []patchOperation
		if customAPI != "" {
			podResourcePatches = applyPolicies(ar, customAPI, rootKind, rootName, rootNamespace)
		} else {
			// Check if Namespace-level policy is applicable.
			podResourcePatches = checkAndApplyNSPolicies(ar)
		}

		for _, podPatch := range podResourcePatches {
			patchOperations = append(patchOperations, podPatch)
		}
	}

	if req.Kind.Kind == "Namespace" {
		fmt.Printf("Recording Namespace...\n")
		releaseName := getReleaseName(ar)
		fmt.Printf("DEF Release name:%s\n", releaseName)
		if releaseName != "" {
			_, nsName, _ := getObjectDetails(ar)
			namespaceHelmAnnotationMap[nsName] = releaseName
		}
	}

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

func checkAndApplyNSPolicies(ar *v1beta1.AdmissionReview) []patchOperation {

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
	fmt.Printf("Pod Namespace:%s\n", podNamespace)
	releaseName := namespaceHelmAnnotationMap[podNamespace]
	fmt.Printf("Release Name:%s\n", releaseName)

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
	podPolicy1 := podPolicy.(platformworkflowv1alpha1.Pol)
	scope := podPolicy1.PolicyResources.Scope
	fmt.Printf("Scope:%s\n",scope)
	if scope == "Namespace" {
		resKindAndName := serviceKind + "-" + serviceInstance
		resAR := resourceNameObjMap[resKindAndName].(*v1beta1.AdmissionReview)
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

	return patchOperations
}

func getReleaseName(ar *v1beta1.AdmissionReview) string {
	req := ar.Request
	annotations1 := make(map[string]string, 0)
	allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")

	if err != nil {
		fmt.Printf("Error in parsing existing annotations")
	} else {
		json.Unmarshal(allAnnotations, &annotations1)
		fmt.Printf("All Annotations:%v\n", annotations1)
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

func saveResource(ar *v1beta1.AdmissionReview) {
	kind, resName, _ := getObjectDetails(ar)
	//key := kind + "/" + namespace + "/" + resName
	key := kind + "-" + resName
	fmt.Printf("Res Key:%s\n", key)
	resourceNameObjMap[key] = ar
}

func saveResourcePolicy(ar *v1beta1.AdmissionReview) {
	req := ar.Request
	body := req.Object.Raw

	resPolicyName, err := jsonparser.GetUnsafeString(body, "metadata", "name")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Resource Policy Name:%s\n", resPolicyName)

	var resourcePolicy platformworkflowv1alpha1.ResourcePolicy
	err = json.Unmarshal(body, &resourcePolicy)
	if err != nil {
	    fmt.Println(err)	
	}

	kind := resourcePolicy.Spec.Resource.Kind
	lowercaseKind := strings.ToLower(kind)
	group := resourcePolicy.Spec.Resource.Group
	version := resourcePolicy.Spec.Resource.Version
	fmt.Printf("Kind:%s, Group:%s, Version:%s\n", kind, group, version)

	podPolicy := resourcePolicy.Spec.Policy
	fmt.Printf("Pod Policy:%v\n", podPolicy)

 	customAPI := group + "/" + version + "/" + lowercaseKind
 	resourcePolicyMap[customAPI] = podPolicy
 	fmt.Printf("Resource Policy Map:%v\n", resourcePolicyMap)
}

func checkServiceLevelPolicyApplicability(ar *v1beta1.AdmissionReview) (string, string, string, string) {
	fmt.Printf("Inside checkServiceLevelPolicyApplicability")

	req := ar.Request
	body := req.Object.Raw

	//fmt.Printf("Body:%v\n", body)
	namespace := req.Namespace
	fmt.Println("Namespace:%s\n",namespace)

	// TODO: looks like we can just keep one - namespace or namespace1
	namespace1, _, _, err := jsonparser.Get(body, "metadata", "namespace")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Namespace1:%s\n", namespace1)
	}

	ownerKind, _, _, err1 := jsonparser.Get(body, "metadata", "ownerReferences", "[0]", "kind")
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
	} else {
		fmt.Printf("ownerKind:%v\n", string(ownerKind))
	}

	ownerName, _, _, err1 := jsonparser.Get(body, "metadata", "ownerReferences", "[0]", "name")
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
	} else {
		fmt.Printf("ownerName:%v\n", string(ownerName))
	}

	ownerAPIVersion, _, _, err1 := jsonparser.Get(body, "metadata", "ownerReferences", "[0]", "apiVersion")
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
	} else {
		fmt.Printf("ownerAPIVersion:%v\n", string(ownerAPIVersion))
	}

	ownerKindS := string(ownerKind)
	ownerNameS := string(ownerName)
	ownerAPIVersionS := string(ownerAPIVersion)

	rootKind, rootName, rootAPIVersion := findRoot(namespace, ownerKindS, ownerNameS, ownerAPIVersionS)
	fmt.Printf("Root Kind:%s\n", rootKind)
	fmt.Printf("Root Name:%s\n", rootName)
	fmt.Printf("Root API Version:%s\n", rootAPIVersion)
	lowercaseKind := strings.ToLower(rootKind)

	// Check if the rootKind, rootName, rootAPIVersion is registered to be applied policies on.
 	customAPI := rootAPIVersion + "/" + lowercaseKind
 	fmt.Printf("Custom API:%s\n", customAPI)
 	if podPolicy, ok := resourcePolicyMap[customAPI]; ok {
	 	fmt.Printf("Resource Policy:%v\n", podPolicy)
	 	return customAPI, rootKind, rootName, string(namespace1)
 	}
	return "", "", "", ""
}

func findRoot(namespace, kind, name, apiVersion string) (string, string, string) {
	rootKind := ""
	rootName := ""
	rootAPIVersion := ""

	time.Sleep(10)

	parts := strings.Split(apiVersion, "/")
	group := ""
	version := ""
	if len(parts) == 2 {
		group = parts[0]
		version = parts[1]
	} else {
		version = parts[0]
	}
	fmt.Printf("Group:%s\n", group)
	fmt.Printf("Version:%s\n", version)
	fmt.Printf("ResName:%s\n", name)
	fmt.Printf("Namespace:%s\n", namespace)

	ownerResKindPlural, _, ownerResApiVersion, ownerResGroup := getKindAPIDetails(kind)

	fmt.Printf("ownerResKindPlural:%s\n", ownerResKindPlural)
	fmt.Printf("ownerResApiVersion:%s\n", ownerResApiVersion)
	fmt.Printf("ownerResGroup:%s\n", ownerResGroup)

	ownerRes := schema.GroupVersionResource{Group: ownerResGroup,
									 		Version: ownerResApiVersion,
									   		Resource: ownerResKindPlural}

	fmt.Printf("OwnerRes:%v\n", ownerRes)
	dynamicClient, err1 := getDynamicClient1()
	if err1 != nil {
		fmt.Printf("Error 1:%v\n", err1)
	    fmt.Println(err1)
		return rootKind, rootName, rootAPIVersion
	}
	instanceObj, err2 := dynamicClient.Resource(ownerRes).Namespace(namespace).Get(
																				   name,
																	   		 	   metav1.GetOptions{})
	if err2 != nil {
		fmt.Printf("Error 2:%v\n", err2)
	    fmt.Println(err2)
		return rootKind, rootName, rootAPIVersion
	}

	ownerReference := instanceObj.GetOwnerReferences()
	if len(ownerReference) == 0 {
		// Reached the root
		// Jump of from the Helm annotation; should be of type <plural>-<name>

		fmt.Printf("Intermediate Root kind:%s\n", kind)
		fmt.Printf("Intermediate Root name:%s\n", name)
		fmt.Printf("Intermediate Root APIVersion:%s\n", apiVersion)

		annotations := instanceObj.GetAnnotations()
		releaseName := annotations["meta.helm.sh/release-name"]
		parts := strings.Split(releaseName, "-")
		if len(parts) == 2 {
			okindLowerCase := parts[0]
			oinstance := parts[1]
			fmt.Printf("KindPluralMap2:%v\n", kindPluralMap)
			oplural := kindPluralMap[okindLowerCase]
			fmt.Printf("OPlural:%s OInstance:%s\n", oplural, oinstance)
			customAPI := ""
			for k, v := range customKindPluralMap {
				if v == oplural {
					customAPI = k
					break
				}
			}
			fmt.Printf("CustomAPI:%s\n", customAPI)
			capiParts := strings.Split(customAPI, "/")
			capiGroup := capiParts[0]
			capiVersion := capiParts[1]
			capiKind := capiParts[2]
			fmt.Printf("capiGroup:%s capiVersion:%s capiKind:%s\n", capiGroup, capiVersion, capiKind)
			return capiKind, oinstance, capiGroup + "/" + capiVersion
		} else {
			return "","",""
		}
	} else {
		owner := ownerReference[0]
		ownerKind := owner.Kind
		ownerName := owner.Name
		ownerAPIVersion := owner.APIVersion
		rootKind, rootName, rootAPIVersion := findRoot(namespace, ownerKind, ownerName, ownerAPIVersion)
		return rootKind, rootName, rootAPIVersion
	}
}

func applyPolicies(ar *v1beta1.AdmissionReview, customAPI, rootKind, rootName, rootNamespace string) []patchOperation {
	req := ar.Request
	body := req.Object.Raw

	podName, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Pod Name:%s\n", podName)

	// TODO: Defaulting to the first container. Take input for additional containers
	res, dataType, _, err1 := jsonparser.Get(body, "spec", "containers", "[0]", "resources")
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
	} else {
		fmt.Printf("Resources:%v\n", string(res))
	}

	var operation string
	fmt.Printf("DataType:%s\n", dataType)
	operation = "add"

	podPolicy := resourcePolicyMap[customAPI]
	fmt.Printf("PodPolicy:%v\n", podPolicy)

	xType := fmt.Sprintf("%T", podPolicy)
	fmt.Printf("Pod Policy type:%s\n", xType) // "[]int"

	patchOperations := make([]patchOperation, 0)

	podPolicy1 := podPolicy.(platformworkflowv1alpha1.Pol)

	// 1. Requests
	cpuRequest := podPolicy1.PolicyResources.Requests.CPU
	memRequest := podPolicy1.PolicyResources.Requests.Memory
	if cpuRequest != "" && memRequest != "" {
		fmt.Printf("CPU Request:%s\n", cpuRequest)
		if strings.Contains(cpuRequest, "values") {
			cpuRequest = getFieldValueFromInstance(cpuRequest,rootKind, rootName)
		}
		fmt.Printf("CPU Request1:%s\n", cpuRequest)

		fmt.Printf("Mem Request:%s\n", memRequest)
		if strings.Contains(memRequest, "values") {
			memRequest = getFieldValueFromInstance(memRequest,rootKind, rootName)
		}
		fmt.Printf("Mem Request1:%s\n", memRequest)

		podResRequest := make(map[string]string,0)
		podResRequest["cpu"] = cpuRequest
		podResRequest["memory"] = memRequest

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
		fmt.Printf("CPU Limit:%s\n", cpuLimit)
		fmt.Printf("Mem Limit:%s\n", memLimit)

		podResLimits := make(map[string]string,0)
		podResLimits["cpu"] = cpuLimit
		podResLimits["memory"] = memLimit

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
		fmt.Printf("Node Selector:%s\n", nodeSelector)
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
	fmt.Printf("Field:%s\n", field)

		//kind, resName, namespace := getObjectDetails(ar)
	lowercaseRootKind := strings.ToLower(rootKind)
		//rootkey := lowercaseRootKind + "/" + rootNamespace + "/" + rootName
	rootkey := lowercaseRootKind + "-" + rootName
	fmt.Printf("Root Key:%s\n", rootkey)
	arSaved := resourceNameObjMap[rootkey].(*v1beta1.AdmissionReview)
	reqObject := arSaved.Request
	reqspec := reqObject.Object.Raw

	fieldValue, _, _, err2 := jsonparser.Get(reqspec, "spec", field)
	fieldValueS := string(fieldValue)
	if err2 != nil {
		fmt.Printf("Error:%v\n", err2)
	} else {
		fmt.Printf("Fields:%v\n", string(fieldValue))
	}
	return fieldValueS
}

func getObjectDetails(ar *v1beta1.AdmissionReview) (string, string, string) {

	req := ar.Request
	body := req.Object.Raw

	kind := req.Kind.Kind
	lowercaseKind := strings.ToLower(kind)

	resName, err := jsonparser.GetUnsafeString(body, "metadata", "name")
	if err != nil {
		fmt.Println(err)
	}

	namespace, err := jsonparser.GetUnsafeString(body, "metadata", "namespace")
	if err != nil {
		fmt.Println(err)
	}
	if namespace == "" {
		namespace = "default"
	}
	return lowercaseKind, resName, namespace
}

func trackCustomAPIs(ar *v1beta1.AdmissionReview) {
	req := ar.Request
	body := req.Object.Raw

	var platformWorkflow platformworkflowv1alpha1.ResourceComposition
	err := json.Unmarshal(body, &platformWorkflow)
	if err != nil {
	    fmt.Println(err)	
	}

	platformWorkflowName, err := jsonparser.GetUnsafeString(body, "metadata", "name")
	if err != nil {
		fmt.Println(err)
	}

	namespace1, _, _, err := jsonparser.Get(body, "metadata", "namespace")
	namespace := string(namespace1)
	if namespace == "" {
		namespace = "default"
	}

	fmt.Printf("ResourceComposition:%s\n", platformWorkflowName)
	kind := platformWorkflow.Spec.NewResource.Resource.Kind
	group := platformWorkflow.Spec.NewResource.Resource.Group
	version := platformWorkflow.Spec.NewResource.Resource.Version
	plural := platformWorkflow.Spec.NewResource.Resource.Plural
	chartURL := platformWorkflow.Spec.NewResource.ChartURL
	chartName := platformWorkflow.Spec.NewResource.ChartName
 	fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s, ChartURL:%s ChartName:%s\n", kind, group, version, plural, chartURL, chartName)
 	customAPI := group + "/" + version + "/" + kind
 	customAPIPlatformWorkflowMap[customAPI] = platformWorkflowName
 	customKindPluralMap[customAPI] = plural
}

func registerManPage(kind, platformworkflow, namespace string) string {
	lowercaseKind := strings.ToLower(kind)

    valuesYamlbytes := GetValuesYaml(platformworkflow, namespace)
    valuesYaml := string(valuesYamlbytes)
    introString := "Here is the values.yaml for the underlying Helm chart representing this resource.\n"
    introString = introString + "The attributes in values.yaml become the Spec properties of the resource.\n\n"
    valuesYaml = introString + valuesYaml
    fmt.Printf("Values YAML:%s\n",valuesYaml)

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

	configMap1, err1 := kubeClient.CoreV1().ConfigMaps(namespace).Create(configMap)

	if err1 != nil {
		fmt.Printf("Error:%s\n", err1.Error())
		return ""
	} else {
		fmt.Printf("Config Map created:%v\n",configMap1)
	}

	usageAnnotationValue := configMapName + ".spec"
	fmt.Printf("Usage Annotation:%s\n", usageAnnotationValue)
	return usageAnnotationValue
}

func getPaCAnnotation(ar *v1beta1.AdmissionReview) map[string]string {
	// Add crd annotation
	annotations1 := make(map[string]string, 0)

	req := ar.Request
	body := req.Object.Raw
	crdkind, _ := jsonparser.GetUnsafeString(body, "spec", "names", "kind")
	crdplural, _ := jsonparser.GetUnsafeString(body, "spec", "names", "plural")
	crdversion, _ := jsonparser.GetUnsafeString(body, "spec", "version")
	crdgroup, _ := jsonparser.GetUnsafeString(body, "spec", "group")
	fmt.Printf("CRDKind:%s, CRDPlural:%s, CRDVersion:%s\n", crdkind, crdplural, crdversion)
	customAPI := crdgroup + "/" + crdversion + "/" + crdkind
	platformWorkflowName, ok := customAPIPlatformWorkflowMap[customAPI]
	chartKinds := ""
	if ok {

			namespace := "default"
	 		chartKindsB := DryRunChart(platformWorkflowName, namespace)
	 		chartKinds = string(chartKindsB)
	 		fmt.Printf("Chart Kinds:%v\n", chartKinds)

	 		// If no kinds are found in the dry run then there is nothing to be done.
	 		if chartKinds == "" {
	 			return annotations1
	 		}

	 	fmt.Printf("Annotating %s\n", chartKinds)
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

	 	fmt.Printf("Unique kinds:%v\n", uniqueKinds)
	 	chartKinds = strings.Join(uniqueKinds, ";")
	  	fmt.Printf("Annotating %s\n", chartKinds)

		allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")
		if err != nil {
			fmt.Printf("Error in parsing existing annotations")
		} else {
			json.Unmarshal(allAnnotations, &annotations1)
			fmt.Printf("All Annotations:%v\n", annotations1)
		}
		annotateRel := "resource/annotation-relationship"
		lowercaseKind := strings.ToLower(crdkind)
		kindPluralMap[lowercaseKind] = crdplural
		fmt.Printf("KindPluralMap1:%v\n", kindPluralMap)
	 	annotationValue := lowercaseKind + "-INSTANCE.metadata.name"
	 	fmt.Printf("Annotation value:%s\n", annotationValue)

		annotateVal := "on:" + chartKinds + ", key:meta.helm.sh/release-name, value:" + annotationValue

		annotations1[annotateRel] = annotateVal

	 	namespace = "default"
	 	manpageConfigMapName := registerManPage(crdkind, platformWorkflowName, namespace)
	 	fmt.Printf("### ManPage ConfigMap Name:%s ####\n", manpageConfigMapName)

	    manPageAnnotation := "resource/usage"
	    manPageAnnotationValue := manpageConfigMapName

	    annotations1[manPageAnnotation] = manPageAnnotationValue
 	}

	fmt.Printf("All Annotations:%v\n", annotations1)

	return annotations1
}

func handleCustomAPIs(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	fmt.Printf("Inside handleCustomAPIs...")
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
	fmt.Printf("Namespace:%s\n", namespace)
	crname, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
	fmt.Printf("CR Name:%s\n", crname)

	//cruid, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "uid")
	// We have to generate a uid as when the request is received there is no uid yet.
	// When the object is persisted Kubernetes will overwrite the uid with a new value - that is okay.
	id := guuid.New()
	cruid := id.String()
	fmt.Printf("CR Uid:%s\n", cruid)

	overridesBytes, _, _, _ := jsonparser.Get(req.Object.Raw, "spec")
	overrides := string(overridesBytes)
	//fmt.Printf("Overrides:%s\n", overrides)

	customAPI := apiVersion + "/" + kind
	fmt.Printf("CustomAPI:%s\n", customAPI)

	// Save name:uid mapping
	customAPIInstance := customAPI + "/" + namespace + "/" + crname

	customAPIInstanceUIDMap[customAPIInstance] = cruid

	platformWorkflowName := customAPIPlatformWorkflowMap[customAPI]
	if platformWorkflowName != "" {
		fmt.Printf("ResourceComposition:%s\n", platformWorkflowName)

		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		var sampleclientset platformworkflowclientset.Interface
		sampleclientset = platformworkflowclientset.NewForConfigOrDie(config)

		platformWorkflow1, err := sampleclientset.WorkflowsV1alpha1().ResourceCompositions(namespace).Get(platformWorkflowName, metav1.GetOptions{})
		fmt.Printf("ResourceComposition:%v\n", platformWorkflow1)
		if err != nil {
			fmt.Errorf("Error:%s\n", err)
		}

		kind := platformWorkflow1.Spec.NewResource.Resource.Kind
		group := platformWorkflow1.Spec.NewResource.Resource.Group
		version := platformWorkflow1.Spec.NewResource.Resource.Version
		plural := platformWorkflow1.Spec.NewResource.Resource.Plural
		chartURL := platformWorkflow1.Spec.NewResource.ChartURL
		chartName := platformWorkflow1.Spec.NewResource.ChartName
 		fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s, ChartURL:%s, ChartName:%s\n", kind, group, version, plural, chartURL, chartName)
 		QueryDeployEndpoint(platformWorkflowName, crname, namespace, overrides)
	}
	return nil
}

// Sets Owner Reference on an object after it has been created. 
// We initially started with this approach but are not using it anymore as the ResourceComposition instance is not technically
// owner of instances of the Custom API. Instead, we are using PaC annotation relationship
// to track this relation. The specific annotation that we look for is the Helm release name annotation.
func setOwnerReference(ar *v1beta1.AdmissionReview) {
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
	cplural := string(GetPlural(ckind, cgroup))
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

					oplural := string(GetPlural(okind, ogroup))
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
		cobj, err1 := dynamicClient.Resource(cres).Namespace(namespace).Get(cinstance, metav1.GetOptions{})
		oobj, err2 := dynamicClient.Resource(ores).Namespace(namespace).Get(oinstance, metav1.GetOptions{})

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
			dynamicClient.Resource(cres).Namespace(namespace).Update(cobj, metav1.UpdateOptions{})
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

func getAccountIdentityAnnotation(ar *v1beta1.AdmissionReview) map[string]string {

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
	return annotations1
}

// This method resolves binding functions - ImportValue, AddLabel, AddAnnotations in the Spec.
// Currently handling such Spec is turned off (there is no reference to this method in the main flow.)
// Leaving this method around for reference.
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
