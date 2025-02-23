package main

import (
	"crypto/tls"
	cert "crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"strconv"
	"context"

	"k8s.io/client-go/util/retry"
	"encoding/json"
	"path/filepath"

	// apiv1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/buger/jsonparser"

	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/dynamic"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	//platformstackclientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/clientset/versioned"
	//platformstackv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/platformstackcontroller/v1alpha1"

)

var (
	cfg *rest.Config
	err error
	dynamicClient dynamic.Interface
	platformStackMap map[string]PlatformStackData
	serviceHost, servicePort, verificationServicePort string
)

type ClusterCapacity struct {
	nodes string `json:"nodes"`
	total_allocatable_cpu string `json:"total_allocatable_cpu"`
	total_allocatable_memory string `json:"total_allocatable_memory"`
	total_allocatable_memory_gb string `json:"total_allocatable_memory_gb"`
}

func init() {
	cfg, err = buildConfig()
	if err != nil {
		panic(err.Error())
	}
	platformStackMap = make(map[string]PlatformStackData)
	serviceHost = "localhost"
	servicePort = "8090"
	verificationServicePort = "5005"
}

func buildConfig() (*rest.Config, error) {
	if home := homeDir(); home != "" {
		kubeconfig := filepath.Join(home, ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Printf("kubeconfig error:%s\n", err.Error())
			fmt.Printf("Trying inClusterConfig..")
			cfg, err = rest.InClusterConfig()
			if err != nil {
				return nil, err
			}
		}
	}
	return cfg, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getDynamicClient() (dynamic.Interface, error) {
	if dynamicClient == nil {
		dynamicClient, err = dynamic.NewForConfig(cfg)
	}
	return dynamicClient, err
}

func getResourceLabels1(req []byte) map[string]string {
	labelMap := make(map[string]string)
	return labelMap
}

func checkIfLabelsExist(req []byte) (bool) {
	propertyValues := make([]string, 0)
	propertyValuesPtr := &propertyValues
	specProperty := "labels"
	checkOnlyExistence := true
	parsePropertyValue(req, specProperty, checkOnlyExistence, propertyValuesPtr)
	fmt.Printf("PropertyValues:%v\n", propertyValues)
	for _, propertyValue := range propertyValues {
		fmt.Printf("Property Value:%s\n", propertyValue)
		if propertyValue == "exists" {
			return true
		}
	}
	return false
}

func getResourceLabels(req []byte) map[string]string {
	labelMap := make(map[string]string)
	fmt.Printf("ResourceDetails:%s\n", string(req))

	hasLabels := checkIfLabelsExist(req)

	if !hasLabels {
		fmt.Println("No labels found")
		return labelMap
	} else {
		fmt.Println("Found labels")
		labels, _, _, err1 := jsonparser.Get(req, "metadata", "labels")
		if err1 != nil {
			fmt.Printf("Error:%s\n", err1)
			return labelMap
		}
		fmt.Printf("Labels:%s\n", labels)
		if err := json.Unmarshal(labels, &labelMap); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		fmt.Printf("Label Map:%v\n", labelMap)
		return labelMap
	}
}

func checkIfResourceCreated(kind, name, namespace string) bool {
	resourceCreated := true
	resourceDetails := queryResourceDetailsEndpoint(kind, name, namespace)
	propertyValues := make([]string, 0)
	propertyValuesPtr := &propertyValues
	checkOnlyExistence := false
	parsePropertyValue(resourceDetails, "reason", checkOnlyExistence, propertyValuesPtr)
	fmt.Printf("PropertyValues:%v\n", propertyValues)
	for _, propertyValue := range propertyValues {
		fmt.Printf("Property Value:%s\n", propertyValue)
		if strings.Contains(propertyValue, "NotFound") {
			resourceCreated = false
		}
	}
	return resourceCreated
}

// From: https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
func AddDeploymentLabel(key, value string, name, namespace string) {
	deployClient := kubeClient.AppsV1().Deployments(namespace)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := deployClient.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", err))
		}
		fmt.Println(deployment.ObjectMeta.Labels)
		if len(deployment.ObjectMeta.Labels) == 0 {
			newLabelsMap := make(map[string]string, 0)
			newLabelsMap[key] = value
			deployment.ObjectMeta.Labels = newLabelsMap
		} else {
			deployment.ObjectMeta.Labels[key] = value
		}
		_, updateErr := deployClient.Update(context.Background(), deployment, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
}

func getLabelAnnotationKeyVal(addLabelDefinition string) (string, string, string, error) {
	//Fn::AddLabel(application/moodle1, MysqlCluster:default.cluster1:Service(filter=master))
	start := strings.Index(addLabelDefinition, "(")
	end := strings.LastIndex(addLabelDefinition, ")")
	args := strings.Split(addLabelDefinition[start+1:end], ",")

	fmt.Printf("key, val: %s, %s\n", args[0], args[1])

	keyValLabel := strings.TrimSpace(args[0])
	keyVal := strings.Split(keyValLabel, "/")
	key := keyVal[0]
	val := keyVal[1]
	resourceString := args[1]
	if len(keyVal) > 2 {
		return "", "", "", fmt.Errorf("Value cannot contain '/' %s", keyVal[2])
	}

	return key, val, resourceString, nil
}

func AddResourceAnnotation(addAnnotationDefinition string) (string, error) {

	key, val, resourceString, err := getLabelAnnotationKeyVal(addAnnotationDefinition)

	if err != nil {
		return "", fmt.Errorf("Annotation value cannot contain '/' %s", val)
	}

	kind, namespace, resourceName, subKind, nameFilterPredicate, _, _ := parseImportString(resourceString)

	//namespace, kind, crdKindName, subKind, err := ParseCompositionPath(args[1])
	fmt.Printf("Parsed Composition Path: %s, %s, %s, %s, %s\n", namespace, kind, resourceName, subKind, nameFilterPredicate)
	jsonData := QueryCompositionEndpoint(kind, namespace, resourceName)
	fmt.Printf("Queried KubeDiscovery: %s\n", string(jsonData))
	if string(jsonData) == "[]" {
		return "", fmt.Errorf("Resource for adding annotation not found.\n")
	}

	// Annotation values cannot contain '/'
	// We should reject annotation values if they contain '/'
	AddLabelAnnotationSubresources(AddAnnotation, jsonData, subKind, nameFilterPredicate, key, val, kind, resourceName, namespace)
	return val, nil
}

func AddResourceLabel(addLabelDefinition string) (string, error) {

	key, val, resourceString,  err := getLabelAnnotationKeyVal(addLabelDefinition)

	if err != nil {
		return "", fmt.Errorf("Label value cannot contain '/' %s", val)
	}

	kind, namespace, resourceName, subKind, nameFilterPredicate, _, _ := parseImportString(resourceString)

	//namespace, kind, crdKindName, subKind, err := ParseCompositionPath(args[1])
	fmt.Printf("Parsed Composition Path: %s, %s, %s, %s, %s\n", namespace, kind, resourceName, subKind, nameFilterPredicate)
	jsonData := QueryCompositionEndpoint(kind, namespace, resourceName)
	fmt.Printf("Queried KubeDiscovery: %s\n", string(jsonData))
	if string(jsonData) == "[]" {
		return "", fmt.Errorf("Resource for adding label not found.\n")
	}

	// Label values cannot contain '/'
	// We should reject label values if they contain '/'
	AddLabelAnnotationSubresources(AddLabel, jsonData, subKind, nameFilterPredicate, key, val, kind, resourceName, namespace)

	return val, nil
}

// This method resolves the annotation value
func ResolveAnnotationValue(val, crName, namespace string) string {
	fmt.Println("Inside ResolveAnnotationValue")

	var resolvedValue string

	start := strings.Index(val, "(")
	end := strings.LastIndex(val, ")")
	args := strings.Split(val[start+1:end], ":")

	if len(args) == 5 {
		fmt.Printf("resolution:%s, name:%s, annotations:%s, name:%s, property:%s\n", 
					args[0], args[1], args[2], args[3], args[4])

		hasCRD := strings.Contains(args[0], "crd")
		if hasCRD {
			crdName := args[1]
			annotationName := args[3]
			propertyName := args[4]
			var err error
			resolvedValue, err = parseCRDAnnotation(crdName, crName, annotationName, propertyName)

			if err != nil {
				fmt.Printf("Error:%s\n", err.Error())
				return ""
			}
		}
	} else {
		resolvedValue = args[0]
	}
	fmt.Printf("Resolved Value:%s\n", resolvedValue)
	return resolvedValue
}


//1. Query CRD to get all annotations
//2. From the list of annotations, find the annotation that matches annotationName
//3. Find the value of that annotation; the value is name of the ConfigMap.key
//4. Read the ConfigMap with above name and use the key to get the data.
//5. Search the data for propertyName
//6. Get the value corresponding to the propertyName
//7. If the propertyValue contains some functions like Fn::Join() resolve them
//   before constructing the value
func parseCRDAnnotation(crdName, crName, annotationName, propertyName string) (string, error) {

	var propertyValue string
	cfg, err := rest.InClusterConfig()
	crdClient, _ := apiextensionsclientset.NewForConfig(cfg)

	crdObj, err := crdClient.CustomResourceDefinitions().Get(context.Background(), crdName, metav1.GetOptions{})
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
		return "", err
	}

	annotations := crdObj.ObjectMeta.Annotations
	fmt.Printf("CRD Annotations:%s\n", annotations)
	annotationValue := annotations[annotationName]
	fmt.Printf("CRD Annotations Value:%s\n", annotationValue)

	args := strings.Split(annotationValue, ".")

	var namespace, configMapName, configMapKey string
	namespace = GetNamespace()
	configMapName = strings.TrimSpace(args[0])
	configMapKey = strings.TrimSpace(args[1])

	if len(args) == 3 {
		namespace = strings.TrimSpace(args[0])
		configMapName = strings.TrimSpace(args[1])
		configMapKey = strings.TrimSpace(args[2])
	}

	fmt.Printf("Namespace:%s, ConfigMapName:%s, ConfigMapKey:%s\n", namespace, configMapName, configMapKey)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	configMapObj, _ := kubeClient.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})

	fmt.Printf("ConfigMapObj:%v\n", configMapObj)

	configMapData := configMapObj.Data
	fmt.Printf("ConfigMapData:%v\n", configMapData)
	configMapValue := configMapData["outputs"]
	fmt.Printf("Data:%s", configMapValue)

	propertyValue = configMapValue

	hasJoinFunc := strings.Contains(configMapValue, "Fn::Join")
	if hasJoinFunc {
		start := strings.Index(configMapValue, "(")
		end := strings.LastIndex(configMapValue, ")")
		args := strings.Split(configMapValue[start+1:end], ",")
		lhs := strings.TrimSpace(string(args[0]))
		rhs := strings.TrimSpace(string(args[1]))
		rhs = strings.Replace(rhs, "\"", "", 2)
		fmt.Printf("LHS:%s, RHS:%s\n", lhs, rhs)
		if lhs == "$instance.name" {
			propertyValue = crName + rhs
			fmt.Printf("PropertyValue:%s\n", propertyValue)
		}
	}
	fmt.Printf("Property Value:%s\n", propertyValue)
	return propertyValue, nil
}

// This is a method that initiates the recursion
// and creates the necessary arrays/data structs
// then calls ParseRequestHelper
func ParseRequest(data []byte) []ResolveData {
	needResolving := make([]ResolveData, 0)
	stringStack := StringStack{Data: "", Mutex: sync.Mutex{}}
	ParseRequestHelper(data, &needResolving, &stringStack)
	return needResolving
}

// This goes through the KubeDiscovery data, which is stored
// and then parses each json object child one by one.
func ParseRequestHelper(data []byte, needResolving *[]ResolveData, stringStack *StringStack) {
	jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
	    object := false
	    array := false
		if dataType.String() == "object" {
			object = true
			stringStack.Push(string(key))
			ParseRequestHelper(value, needResolving, stringStack)
			stringStack.Pop()
		} 
		if dataType.String() == "array" {
			array = true
			//fmt.Printf("Array Key:%s\n", key)
			//fmt.Printf("Array value value:%v\n", string(value))
			//fmt.Printf("StringStack:%s\n", string(stringStack))
			stringStack.Push(string(key))
			count := 0
			jsonparser.ArrayEach(value, func(value1 []byte, dataType1 jsonparser.ValueType, offset1 int, err error) {
				stringStack.Push(strconv.Itoa(count))
				count = count + 1
				//fmt.Printf("Value1:%v\n", string(value1))
				ParseRequestHelper(value1, needResolving, stringStack)
				stringStack.Pop()
				return
			})
			stringStack.Pop()
		}
		if !object && !array {
			//fmt.Printf("ParseRequestHelper Key:%s\n", string(key))
			stringStack.Push(string(key))
			jsonPath := stringStack.Peek()
			//fmt.Printf("ParseRequestHelper Jsonpath:%s\n", jsonPath)
			val := strings.TrimSpace(string(value))
			hasImportFunc := strings.Contains(val, "Fn::ImportValue")
			hasLabelFunc := strings.Contains(val, "Fn::AddLabel")
			hasAnnotationFunc := strings.Contains(val, "Fn::AddAnnotation")
			if hasImportFunc {
				start := strings.Index(val, "(")
				end := strings.LastIndex(val, ")")
				importString := val[start+1 : end]
				needResolve := ResolveData{
					JSONTreePath: jsonPath,
					ImportString: importString,
					FunctionType: ImportValue,
				}
				*needResolving = append(*needResolving, needResolve)
				stringStack.Pop()
			} else if hasLabelFunc {
				//val, err := AddResourceLabel(val)

				/*if err != nil {
					stringStack.Pop()
					return err
				}*/
				needResolve := ResolveData{
					JSONTreePath: jsonPath,
					Value:        val,
					FunctionType: AddLabel,
				}
				*needResolving = append(*needResolving, needResolve)
				stringStack.Pop()
				// I would use defer here but am unsure
				// how it works with all the recursion and returns..
			} else if hasAnnotationFunc {
				needResolve := ResolveData{
					JSONTreePath: jsonPath,
					Value:        val,
					FunctionType: AddAnnotation,
				}
				*needResolving = append(*needResolving, needResolve)
				stringStack.Pop()
				// I would use defer here but am unsure
				// how it works with all the recursion and returns..
			} else {
				stringStack.Pop()
			}
		}
		return nil
	})
}

// This is used to Resolve the Import String to its value
// importString can have one of the following structures:
//	1. Fully Qualified Resource Name syntax: <Kind>:<Namespace>.<Resource-Name>
//     -> Resolved Import String value in this case is the Resource name
//  2. Fully Qualified sub resource name: <Kind>:<Namespace>.<Resource-Name>:<Sub-kind>(filter=<filter-predicate>)
//     -> Resolved Import String value in this case is the name of the sub kind  resource that matches the filter predicate.
//        Filter predicate is optional. If not specified then the resolved value is the name of the first sub kind resource.
//  3. Fully Qualified Resource Spec property: <Kind>:<Namespace>.<Resource-Name>.<spec-properpty-name>
func ResolveImportString(importString string) (string, error) {
	parts := strings.Split(strings.TrimSpace(importString), ":")
	var resolvedValue string
	var err error
	//fmt.Printf("Length of ImportString parts:%d\n", len(parts))
	if len(parts) == 2 {
		resolvedValue, err = ResolveKind(importString)
	}
	if len(parts) == 3 {
		resolvedValue, err = ResolveSubKind(importString)
	}

	fmt.Printf("Resolved Value:%s\n", resolvedValue)
	return resolvedValue, err
}

func parseFilterPredicate(nameProperty string) (string, string) {
	filterPredicate := ""
	updatedNameProperty := nameProperty

	hasFilterPredicate := strings.Contains(nameProperty, "filter")
	if hasFilterPredicate {
		start := strings.Index(nameProperty, "(")
		end := strings.LastIndex(nameProperty, ")")
		args := strings.Split(nameProperty[start+1:end], "=")
		filterPredicate = args[1]
		fmt.Printf("Filter Predicate:%s\n", filterPredicate)
		updatedNameProperty = nameProperty[0:start] // remove the filter predicate
	} 
	return filterPredicate, updatedNameProperty
}

func parseImportString(importString string) (string, string, string, string, string, string, string) {

	parts := strings.Split(strings.TrimSpace(importString), ":")
	kind := parts[0]
	fqResourceName := parts[1]
	subKind := parts[2]
	args := strings.Split(fqResourceName, ".")
	namespace := args[0]
	resourceName := args[1]

	nameFilterPredicate := ""
	specPropertyFilterPredicate := ""
	// Fn::ImportValue(MysqlCluster:default.cluster1:Service(filter=master))
	// Service or Service(filter=master) or Service.volumemount(filter=abc) or Service(filter=a).label(filter=abc)

	// Fn::ImportValue(MysqlCluster:default.cluster1:Deployment.mountPath)
	hasSpecProperty := strings.Contains(subKind, ".")
	specProperty := ""
	if hasSpecProperty {
		args := strings.Split(subKind, ".")
		subKind = args[0]
		specProperty = args[1]
		specPropertyFilterPredicate, specProperty = parseFilterPredicate(specProperty)
	}

	nameFilterPredicate, subKind = parseFilterPredicate(subKind)

	return kind, namespace, resourceName, subKind, nameFilterPredicate, specProperty, specPropertyFilterPredicate
}

func ResolveSubKind(importString string) (string, error) {

	kind, namespace, resourceName, subKind, nameFilterPredicate, specProperty, specFilterPredicate := parseImportString(importString)

	fmt.Printf("Kind:%s, Namespace:%s, resourceName:%s, SubKind:%s, NameFilterPredicate:%s SpecProp:%s, SpecFilter:%s\n", 
		kind, namespace, resourceName, subKind, nameFilterPredicate, specProperty, specFilterPredicate)

	jsonData := QueryCompositionEndpoint(kind, namespace, resourceName)
	fmt.Printf(" ABC Queried KubeDiscovery: %s\n", string(jsonData))
	name, err := ParseDiscoveryJSON(jsonData, subKind, nameFilterPredicate)
	if err != nil {
		return "", err
	}
	fmt.Printf("Found name: %s\n", name)

	if specProperty != "" {
		specPropertyValue := resolveSpecProperty(namespace, subKind, name, specProperty, specFilterPredicate)
		return specPropertyValue, nil
	} else {
		return name, nil
	}
}

func ResolveKind(importString string) (string, error) {
	return "Not implemented yet.", nil
}

func resolveSpecProperty(namespace, subKind, name, specProperty, specFilterPredicate string) (string) {
	resourceDetails := queryResourceDetailsEndpoint(subKind, name, namespace)
	propertyValues := make([]string, 0)
	propertyValuesPtr := &propertyValues
	checkOnlyExistence := false
	parsePropertyValue(resourceDetails, specProperty, checkOnlyExistence, propertyValuesPtr)
	fmt.Printf("PropertyValues:%v\n", propertyValues)
	for _, propertyValue := range propertyValues {
		fmt.Printf("Property Value:%s\n", propertyValue)
		if strings.Contains(propertyValue, specFilterPredicate) {
			return propertyValue
		}
	}
	return ""
}

func parsePropertyValue(resourceDetails []byte, specProperty string, checkOnlyExistence bool, propertyValues *[]string) {
	jsonparser.ObjectEach(resourceDetails, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if dataType.String() == "string" {
			//fmt.Printf("Key:%s, Value:%s\n", key, value)
			if string(key) == specProperty {
				if checkOnlyExistence {
					*propertyValues = append(*propertyValues, "exists")
				} else {
					*propertyValues = append(*propertyValues, string(value))
				}
			}
		}
		if dataType.String() == "array" {
			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				parsePropertyValue(value, specProperty, checkOnlyExistence, propertyValues)
			})
		}
		if dataType.String() == "object" {
			if string(key) == specProperty && checkOnlyExistence {
				*propertyValues = append(*propertyValues, "exists")
				return nil
			} else {
				parsePropertyValue(value, specProperty, checkOnlyExistence, propertyValues)
			}
		}
		return nil
	})
	return
}

// This is to parse Annotation Paths of structure:
// [Namespace]?.[Kind].[InstanceName].[outputVariable]
// This is used in ImportValue Fn::
// namespace1.MysqlCluster.cluster1.service
func ParseAnnotationPath(annotationPath string) (string, string, string, string, error) {
	path := strings.Split(strings.TrimSpace(annotationPath), ".")
	if len(path) == 3 {
		return "default", path[0], path[1], path[2], nil
	}
	if len(path) != 4 {
		return "", "", "", "", fmt.Errorf("AnnotationPath inside function is not of len 3 or 4")
	}
	return path[0], path[1], path[2], path[3], nil
}

// This is to parse Composition Paths of structure:
// [Namespace]?.[Kind].[CrdKindName].[SubKind]
// ex: namespace1.Moodle.moodle1.Deployment
// returns [Namespace],[Kind],[CrdKindName],[SubKind]
func ParseCompositionPath(compositionPath string) (string, string, string, string, error) {
	path := strings.Split(strings.TrimSpace(compositionPath), ".")
	if len(path) == 3 { //can be len 3, namespace opttional
		return "default", path[0], path[1], path[2], nil
	}
	if len(path) != 4 {
		return "", "", "", "", fmt.Errorf("CompositionPath inside function is not of len 3")
	}
	return path[0], path[1], path[2], path[3], nil
}

func ParseDiscoveryJSON(composition []byte, subKind string, filterPredicate string) (string, error) {
	names := make([]string, 0)
	// note that for the recursive style I do, I must pass a ptr value to the function
	// the logic is that are that inside of ObjectEach or ArrayEach I cannot return from
	// ParseDiscoveryJSON, since I would actually be returning from a lamba func that is
	// defined to return err by the docs. Note you cannot have a string pointer in Go

	namesPtr := &names
	jsonparser.ArrayEach(composition, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		ParseDiscoveryJSONHelper(value, subKind, namesPtr)
	})
	// Since we are querying kubedisovery using name, namespace, kind (given by the yaml path)
	// There might be multiple instances of the subKind. If filterPredicate is specified, we return
	// the first instance whose name matches the filter predicate. If filterPredicate is empty string,
	// then we return the first name.
	var name string

	fmt.Printf("ParseDiscoveryJSON Names:%v\n", names)

	if len(names) > 0 {
		if filterPredicate == "" {
			return names[0], nil
		} else {
			for _, name := range names {
				// Return the first name that matches the filterPredicate
				if strings.Contains(name, filterPredicate) {
					return name, nil
				}
			}
		}

		if len(names) == 1 {
			name = names[0]
			return name, nil
		}
	}
	return "", fmt.Errorf("Resource not found. Kind:%s", subKind)
}

// Finds the name of the subresource of a CRD resource
func ParseDiscoveryJSONHelper(composition []byte, subKind string, subName *[]string) error {
	jsonparser.ObjectEach(composition, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if dataType.String() == "array" {
			// go through each children object.
			jsonparser.ArrayEach(value, func(value1 []byte, dataType jsonparser.ValueType, offset int, err error) {
				// Debug prints
				//kind1, _ := jsonparser.GetUnsafeString(value1, "Kind")
				//name1, _ := jsonparser.GetUnsafeString(value1, "Name")
				//fmt.Printf(" ParseDiscoveryJSONHelper kind: %s\n", kind1)
				//fmt.Printf(" ParseDiscoveryJSONHelper name: %s\n", name1)

				// each object in Children array is a json object
				// looking like {Kind: deployment, name: moodle1, level: 2, namespace: default}

				// if Kind key exists and it is equal to subKind, we found a resource name
				// that matches the type. It may be possible to have multiple Ingresses or Deployments?
				// so I return a list of strings, usually it will return one Name only. Handle by func caller.
				if kind, err := jsonparser.GetUnsafeString(value1, "Kind"); err == nil && kind == subKind {
					name, _ := jsonparser.GetUnsafeString(value1, "Name")
					*subName = append(*subName, name)
					ParseDiscoveryJSONHelper(value1, subKind, subName)
				} else { // Found a Kind that is not subKind ("Deployment") for example
					//Continue traversing
					ParseDiscoveryJSONHelper(value1, subKind, subName)
				}
			})
		} else { //if it is a
			return nil
		}
		return nil
	})
	return fmt.Errorf("Could not find a name for kind")
}

func AddLabelAnnotationSubresources(functype Function, composition []byte, subKind, nameFilterPredicate, labelkey, labelvalue, kind, resourceName, namespace string) error {
	//addLabel(labelkey, labelvalue, kind, resourceName, namespace)
	jsonparser.ArrayEach(composition, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		AddLabelAnnotationSubresourcesHelper(functype, value, subKind, nameFilterPredicate, labelkey, labelvalue, namespace)
	})
	return fmt.Errorf("Could not all label to all sub resources.")
}

func AddLabelAnnotationSubresourcesHelper(functype Function, composition []byte, subKind, nameFilterPredicate, labelkey, labelvalue, namespace string) error {
	jsonparser.ObjectEach(composition, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		//fmt.Printf("Datatype:%s key:%s\n", dataType.String(), string(key))
		if dataType.String() == "array" {
			// go through each children object.
			jsonparser.ArrayEach(value, func(value1 []byte, dataType jsonparser.ValueType, offset int, err error) {
				// Debug prints
				kind1, _ := jsonparser.GetUnsafeString(value1, "Kind")
				name1, _ := jsonparser.GetUnsafeString(value1, "Name")
				fmt.Printf(" AddLabelAnnotationSubresources kind:%s\n", kind1)
				fmt.Printf(" AddLabelAnnotationSubresources name:%s\n", name1)
				fmt.Printf(" AddLabelAnnotationSubresources subKind:%s\n", subKind)
				fmt.Printf(" AddLabelAnnotationSubresources filterPredicate:%s\n", nameFilterPredicate)
				if strings.Contains(subKind, kind1) {
					if nameFilterPredicate == "" || strings.Contains(name1, nameFilterPredicate) {
						if functype == AddLabel {
							addLabel(labelkey, labelvalue, kind1, name1, namespace)
						}
						if functype == AddAnnotation {
							addAnnotation(labelkey, labelvalue, kind1, name1, namespace)
						}
					}
				}
				AddLabelAnnotationSubresourcesHelper(functype, value1, subKind, nameFilterPredicate, labelkey, labelvalue, namespace)
			})
		} else {
			return nil
		}
		return nil
	})
	return fmt.Errorf("Could not all label to all sub resources.")
}

func addLabel(labelkey, labelvalue, kind, resource, namespace string) {
	fmt.Printf("Adding label kind:%s, resource:%s, namespace:%s\n", kind, resource, namespace)

	dynamicClient, err := getDynamicClient()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	resourceKindPlural, _, resourceApiVersion, resourceGroup := getKindAPIDetails(kind)
	res := schema.GroupVersionResource{Group: resourceGroup,
									   Version: resourceApiVersion,
									   Resource: resourceKindPlural}
	obj, err1 := dynamicClient.Resource(res).Namespace(namespace).Get(context.Background(), resource, metav1.GetOptions{})
	if err1 != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	objCopy := obj.DeepCopy()
	labelMap := objCopy.GetLabels()
	if labelMap == nil {
		labelMap = make(map[string]string)
	}
	labelMap[labelkey] = labelvalue
	objCopy.SetLabels(labelMap)

	fmt.Printf("Before adding label.\n")
	_, err = dynamicClient.Resource(res).Namespace(namespace).Update(context.Background(), objCopy, metav1.UpdateOptions{})
	fmt.Printf("Done adding label.\n")

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func addAnnotation(labelkey, labelvalue, kind, resource, namespace string) {
	fmt.Printf("Adding annotation kind:%s, resource:%s, namespace:%s\n", kind, resource, namespace)

	dynamicClient, err := getDynamicClient()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	resourceKindPlural, _, resourceApiVersion, resourceGroup := getKindAPIDetails(kind)
	res := schema.GroupVersionResource{Group: resourceGroup,
									   Version: resourceApiVersion,
									   Resource: resourceKindPlural}
	obj, err1 := dynamicClient.Resource(res).Namespace(namespace).Get(context.Background(), resource, metav1.GetOptions{})
	if err1 != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	objCopy := obj.DeepCopy()
	annotationMap := objCopy.GetAnnotations()
	if annotationMap == nil {
		annotationMap = make(map[string]string)
	}
	annotationMap[labelkey] = labelvalue
	objCopy.SetAnnotations(annotationMap)

	fmt.Printf("Before adding annotations.\n")
	_, err = dynamicClient.Resource(res).Namespace(namespace).Update(context.Background(), objCopy, metav1.UpdateOptions{})
	fmt.Printf("Done adding annotations.\n")

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func GetPlural(kind string) []byte {
	args := fmt.Sprintf("kind=%s", kind)
	fmt.Printf("Inside GetPlural...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/kubeplus/getPlural?%s", serviceHost, servicePort, args)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func CheckResource(kind, plural string) []byte {
	//fmt.Printf("Inside CheckResource...\n")
	args := fmt.Sprintf("kind=%s&plural=%s", kind, plural)
	url1 := fmt.Sprintf("http://%s:%s/apis/kubeplus/checkResource?%s", serviceHost, servicePort, args)
	//fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func CheckApplicationNodeName(nodeName string) bool  {
	fmt.Printf("Inside CheckApplicationNodeName...%s\n", nodeName)

	result := false
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/nodes", serviceHost, verificationServicePort)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	nodeNamesString := string(body)
	fmt.Printf("Cluster nodes:%s\n", nodeNamesString)
	parts := strings.Split(nodeNamesString, ",")
	if len(parts) > 0 {
		for _, n := range parts {
			if n != "" && nodeName == strings.TrimSpace(n) {
				fmt.Printf("Cluster Node Name:%s Spec Node Name:%s\n", n, nodeName)
				fmt.Printf("Valid Node name specified.\n")
				result = true
				break
			}
		}
	}

	if !result {
		fmt.Printf("Node name not valid:%s\n", nodeName)
	}

	return result
}

// http://10.80.10.160:90/apis/kubeplus/deploy?platformworkflow=hello-world-service-composition&customresource=hello-world-tenant1&namespace=default&overrides={"greeting":"Hello Kubernauts - This is KubePlus"}
// http://10.80.10.160:90/apis/kubeplus/deploy?platformworkflow=hello-world-service-composition&customresource=hello-world-tenant1&namespace=default&overrides={"greeting":"Hi"}
func QueryDeployEndpoint(platformworkflow, customresource, namespace, overrides, cpu_req, cpu_lim, mem_req, mem_lim, labels, sourceKind, sourceKindPlural string) []byte {
	encodedOverrides := url.QueryEscape(overrides)
	//fp, _ := os.Create("/crdinstances/" + platformworkflow + "-" + customresource)
	//fp.WriteString(encodedOverrides)
	//fp.Close()

        CreateOverrides(platformworkflow, customresource)

	args := fmt.Sprintf("platformworkflow=%s&customresource=%s&namespace=%s&overrides=%s&cpu_req=%s&cpu_lim=%s&mem_req=%s&mem_lim=%s&labels=%s&sourceKind=%s&sourceKindPlural=%s", platformworkflow, customresource, namespace, encodedOverrides, cpu_req, cpu_lim, mem_req, mem_lim,labels,sourceKind,sourceKindPlural)
	//fmt.Printf("Inside QueryDeployEndpoint...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/kubeplus/deploy?%s", serviceHost, servicePort, args)
	//fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func CheckLicense(kind, webhook_namespace string) string {
	//fmt.Printf("Inside CheckLicense...\n")
	args := fmt.Sprintf("kind=%s&namespace=%s", kind, webhook_namespace)
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/checklicense?%s", serviceHost, verificationServicePort, args)
	//fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	bodyString := string(body)
	return bodyString
}

func DryRunChart(platformworkflow, namespace string) []byte {
	args := fmt.Sprintf("platformworkflow=%s&namespace=%s&dryrun=true", platformworkflow, namespace)
	fmt.Printf("Inside DryRunChart...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/kubeplus/deploy?%s", serviceHost, servicePort, args)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func TestChartDeployment(kind, namespace, chartName, chartURL string) []byte {
	fmt.Printf("Inside TestChartDeployment...\n")
	encodedChartURL := url.QueryEscape(chartURL)
	args := fmt.Sprintf("kind=%s&namespace=%s&chartName=%s&chartURL=%s", kind, namespace, chartName, encodedChartURL)
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/kubeplus/testChartDeployment?%s", serviceHost, servicePort, args)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func CreateOverrides(platformworkflow, customresource string) []byte {
	//fmt.Printf("Inside CreateOverrides...\n")
	args := fmt.Sprintf("platformworkflow=%s&customresource=%s", platformworkflow, customresource)
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/overrides?%s", serviceHost, verificationServicePort, args)
	//fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func CheckClusterCapacity(cpuRequests, cpuLimits, memRequests, memLimits string) (bool, string) {
	//fmt.Printf("Inside CheckClusterCapacity...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/cluster_capacity", serviceHost, verificationServicePort)
	//fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)

	var clusterCapacity ClusterCapacity
        json.Unmarshal(body, &clusterCapacity)
	//fmt.Printf("Cluster Capacity:%v\n", clusterCapacity)

	bodyString := string(body)
	//jsonresp := bodyString.(map[string]string) 

	//totalAllocatableCPU := jsonresp['total_allocatable_cpu']
	//totalAllocatableMemoryGB := jsonresp['total_allocatable_memory_gb']

	//bodyString: total_allocatable_cpu:123,total_allocatable_memory_gb:2
	parts := strings.Split(bodyString,",")
	cpuParts := strings.Split(parts[0],":")
	totalAllocatableCPU, _ := strconv.Atoi(strings.TrimSpace(cpuParts[1]))

	memParts := strings.Split(parts[1],":")
	totalAllocatableMemoryGB, _ := strconv.ParseFloat(strings.TrimSpace(memParts[1]), 64)
	
	/*fmt.Printf("Total Allocatable CPU:%d\n", totalAllocatableCPU)
	fmt.Printf("Total Allocatable Memory:%f\n", totalAllocatableMemoryGB)
	*/
	cpuRequestsI, _ := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(cpuRequests, "m", "")))
	cpuLimitsI, _ := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(cpuLimits, "m", "")))

	memRequestsI, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(memRequests, "Gi", "")), 64)
	memLimitsI, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(memLimits, "Gi", "")), 64)

	result := true
	message := "Quota is within limits."
	if totalAllocatableCPU < cpuRequestsI || totalAllocatableCPU < cpuLimitsI {
		result = false
		message = fmt.Sprintf("Specified CPU quota values are more than CPU capacity: %d %d - %d", cpuRequestsI, cpuLimitsI, totalAllocatableCPU)
		return result, message
	}

	if totalAllocatableMemoryGB < memRequestsI || totalAllocatableMemoryGB < memLimitsI {
		result = false
		message = fmt.Sprintf("Specified Memory quota values are more than Memory capacity: %d %d - %d", memRequestsI, memLimitsI, totalAllocatableMemoryGB)
		return result, message
	}

	return result, message
}

func GetExistingResourceCompositions() []byte {
	fmt.Printf("Inside GetExistingResourceCompositions...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/resourcecompositions", serviceHost, verificationServicePort)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func CheckChartExists(chartURL string) []byte {
	//fmt.Printf("Inside CheckChartExists...\n")
	encodedChartURL := url.QueryEscape(chartURL)
	args := fmt.Sprintf("chartURL=%s", encodedChartURL)
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/checkchartexists?%s", serviceHost, verificationServicePort, args)
	//fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func LintChart(chartURL string) []byte {
	//fmt.Printf("Inside LintChart...\n")
	encodedChartURL := url.QueryEscape(chartURL)
	args := fmt.Sprintf("chartURL=%s", encodedChartURL)
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/dryrunchart?%s", serviceHost, verificationServicePort, args)
	//fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func AnnotateCRD(kind, plural, group, chartkinds string) []byte {
	args := fmt.Sprintf("kind=%s&plural=%s&group=%s&chartkinds=%s", kind, plural, group, chartkinds)
	fmt.Printf("Inside AnnotateCRD...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/kubeplus/annotatecrd?%s", serviceHost, servicePort, args)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func DeleteCRDInstances(kind, group, version, plural, namespace, instance string) []byte {
	fmt.Printf("Inside deleteCRDInstances...\n")
	args := fmt.Sprintf("kind=%s&group=%s&version=%s&plural=%s&namespace=%s&instance=%s", kind, group, version, plural, namespace, instance)
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/kubeplus/deletecrdinstances?%s", serviceHost, servicePort, args)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func QueryCompositionEndpoint(kind, namespace, crdKindName string) []byte {
	args := fmt.Sprintf("kind=%s&instance=%s&namespace=%s", kind, crdKindName, namespace)
	fmt.Printf("Inside QueryCompositionEndpoint...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/platform-as-code/v1/composition?%s", serviceHost, servicePort, args)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func GetValuesYaml(platformworkflow, namespace string) []byte {
	args := fmt.Sprintf("platformworkflow=%s&namespace=%s", platformworkflow, namespace)
	//fmt.Printf("Inside GetValuesYaml...\n")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis/kubeplus/getchartvalues?%s", serviceHost, servicePort, args)
	//fmt.Printf("Url:%s\n", url1)
	body := queryKubeDiscoveryService(url1)
	return body
}

func GetNamespace() string {
	filePath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	content, err := ioutil.ReadFile(filePath)
    if err != nil {
     	fmt.Printf("Namespace file reading error:%v\n", err)
    }
    ns := string(content)
    ns = strings.TrimSpace(ns)
    //fmt.Printf("CRD Hook NS:%s\n", ns)
    return ns
}

func getServiceEndpoint(servicename string) (string, string) {
	//fmt.Printf("..Inside getServiceEndpoint...\n")
	namespace := GetNamespace() // Use the namespace in which kubeplus is deployed.
	//discoveryService := "discovery-service"
	cfg, _ := rest.InClusterConfig()
	kubeClient, _ := kubernetes.NewForConfig(cfg)
	serviceClient := kubeClient.CoreV1().Services(namespace)
	discoveryServiceObj, _ := serviceClient.Get(context.Background(), servicename, metav1.GetOptions{})
	host := discoveryServiceObj.Spec.ClusterIP
	port := discoveryServiceObj.Spec.Ports[0].Port
	stringPort := strconv.Itoa(int(port))
    //fmt.Printf("Host:%s, Port:%s\n", host, stringPort)
	return host, stringPort
}

// Rename this function to a more generic name since we use it to trigger Custom Resource deployment as well.
func queryKubeDiscoveryService(url1 string) []byte {
	//fmt.Printf("..inside queryKubeDiscoveryService")
	u, err := url.Parse(url1)
	//fmt.Printf("URL:%s\n",u)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("sending request failed: %s", err.Error())
		fmt.Println(err)
		return []byte(err.Error())
	}
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)

	//fmt.Println("Response status:%s\n",resp.Status)
	//fmt.Println("Response body:%s\n",string(resp_body))
	//fmt.Println("Exiting queryKubeDiscoveryService")
	return resp_body
}

/// Functions to query Kubediscovery running as an Aggregated API Server 
/// Not used anymore
// Used to query KubeDiscovery api server
func queryResourceDetailsEndpoint(kind, name, namespace string) []byte {
	fmt.Printf("..Inside queryResourceDetailsEndpoint...")
	args := fmt.Sprintf("kind=%s&instance=%s&namespace=%s", kind, name, namespace)
	serviceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	servicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	var url1 string
	url1 = fmt.Sprintf("https://%s:%s/apis/platform-as-code/v1/resourceDetails?%s", serviceHost, servicePort, args)
	body := queryKubeAPIServer(url1)
	return body
}

func queryAPIsEndpoint(kind, namespace, crdKindName string) []byte {
	args := fmt.Sprintf("kind=%s&instance=%s&namespace=%s", kind, crdKindName, namespace)
	serviceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	servicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	var url1 string
	url1 = fmt.Sprintf("http://%s:%s/apis?%s", serviceHost, servicePort, args)
	fmt.Printf("Url:%s\n", url1)
	body := queryKubeAPIServer(url1)
	return body
}

func queryKubeAPIServer(url1 string) []byte {
	fmt.Printf("..inside queryKubeAPIServer")
	caToken := getToken()
	caCertPool := getCACert()
	u, err := url.Parse(url1)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(caToken)))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("sending request failed: %s", err.Error())
		fmt.Println(err)
	}
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)

	//fmt.Println(resp.Status)
	//fmt.Println(string(resp_body))
	//fmt.Println("Exiting QueryCompositionEndpoint")
	return resp_body
}

// Ref:https://stackoverflow.com/questions/30690186/how-do-i-access-the-kubernetes-api-from-within-a-pod-container
func getToken() []byte {
	caToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		panic(err) // cannot find token file
	}
	//fmt.Printf("Token:%s", caToken)
	return caToken
}

// Ref:https://stackoverflow.com/questions/30690186/how-do-i-access-the-kubernetes-api-from-within-a-pod-container
func getCACert() *cert.CertPool {
	caCertPool := cert.NewCertPool()
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		panic(err) // Can't find cert file
	}
	//fmt.Printf("CaCert:%s",caCert)
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool
}
