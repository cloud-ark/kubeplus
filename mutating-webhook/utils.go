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

	"k8s.io/client-go/util/retry"

	// apiv1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buger/jsonparser"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

)

//    You have two options to Update() this Deployment:
//
//    1. Modify the "deployment" variable and call: Update(deployment).
//       This works like the "kubectl replace" command and it overwrites/loses changes
//       made by other clients between you Create() and Update() the object.
//    2. Modify the "result" returned by Get() and retry Update(result) until
//       you no longer get a conflict error. This way, you can preserve changes made
//       by other clients between Create() and Update(). This is implemented below
//			 using the retry utility package included with client-go. (RECOMMENDED)
//
// More Info:
// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency

// From: https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
func AddDeploymentLabel(key, value string, name, namespace string) {
	deployClient := kubeClient.AppsV1().Deployments(namespace)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		deployment, err := deployClient.Get(name, metav1.GetOptions{})
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
		_, updateErr := deployClient.Update(deployment)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
}

func AddResourceLabel(addLabelDefinition string) (string, error) {

	start := strings.Index(addLabelDefinition, "(")
	end := strings.LastIndex(addLabelDefinition, ")")
	args := strings.Split(addLabelDefinition[start+1:end], ",")

	fmt.Printf("key, val: %s, %s\n", args[0], args[1])

	keyValLabel := strings.TrimSpace(args[0])
	keyVal := strings.Split(keyValLabel, "/")
	key := keyVal[0]
	val := keyVal[1]

	kind, namespace, resourceName, subKind, _, _ := parseImportString(args[1])

	//namespace, kind, crdKindName, subKind, err := ParseCompositionPath(args[1])
	fmt.Printf("Parsed Composition Path: %s, %s, %s, %s\n", namespace, kind, resourceName, subKind)
	jsonData := QueryCompositionEndpoint(kind, namespace, resourceName)
	fmt.Printf("Queried KubeDiscovery: %s\n", string(jsonData))
	name, err := ParseDiscoveryJSON(jsonData, subKind, "")
	if err != nil {
		return "", err
	}
	//found one resource that matches "Deployment"
	fmt.Printf("Found name: %s\n", name)
	switch subKind {
		case "Deployment":
		AddDeploymentLabel(key, val, name, namespace)
	}

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

	crdObj, err := crdClient.CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
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
	namespace = "default"
	configMapName = strings.TrimSpace(args[0])
	configMapKey = strings.TrimSpace(args[1])

	if len(args) == 3 {
		namespace = strings.TrimSpace(args[0])
		configMapName = strings.TrimSpace(args[1])
		configMapKey = strings.TrimSpace(args[2])
	}

	fmt.Printf("Namespace:%s, ConfigMapName:%s, ConfigMapKey:%s\n", namespace, configMapName, configMapKey)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	configMapObj, _ := kubeClient.CoreV1().ConfigMaps(namespace).Get(configMapName, metav1.GetOptions{})

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
		if dataType.String() == "object" {
			stringStack.Push(string(key))
			ParseRequestHelper(value, needResolving, stringStack)
			stringStack.Pop()
		} else {
			//fmt.Printf("ParseRequestHelper Key:%s\n", key)
			stringStack.Push(string(key))
			jsonPath := stringStack.Peek()
			//fmt.Printf("ParseRequestHelper Jsonpath:%s\n", jsonPath)
			val := strings.TrimSpace(string(value))
			hasImportFunc := strings.Contains(val, "Fn::ImportValue")
			hasLabelFunc := strings.Contains(val, "Fn::AddLabel")
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

				val, err := AddResourceLabel(val)

				if err != nil {
					stringStack.Pop()
					return err
				}

				needResolve := ResolveData{
					JSONTreePath: jsonPath,
					Value:        val,
					FunctionType: AddLabel,
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

func parseImportString(importString string) (string, string, string, string, string, string) {

	parts := strings.Split(strings.TrimSpace(importString), ":")
	kind := parts[0]
	fqResourceName := parts[1]
	subKind := parts[2]
	args := strings.Split(fqResourceName, ".")
	namespace := args[0]
	resourceName := args[1]

	filterPredicate := ""
	// Fn::ImportValue(MysqlCluster:default.cluster1:Service(filter=master))
	hasFilterPredicate := strings.Contains(subKind, "filter")
	if hasFilterPredicate {
		start := strings.Index(subKind, "(")
		end := strings.LastIndex(subKind, ")")
		args := strings.Split(subKind[start+1:end], "=")
		filterPredicate = args[1]
		fmt.Printf("Filter Predicate:%s\n", filterPredicate)
		subKind = subKind[0:start] // remove the filter predicate
	} 
	// Fn::ImportValue(MysqlCluster:default.cluster1:Deployment.mountPath)
	hasSpecProperty := strings.Contains(subKind, ".")
	specProperty := ""
	if hasSpecProperty {
		args := strings.Split(subKind, ".")
		specProperty = args[1]
		subKind = args[0]
	}

	return kind, namespace, resourceName, subKind, filterPredicate, specProperty
}

func ResolveSubKind(importString string) (string, error) {

	kind, namespace, resourceName, subKind, filterPredicate, specProperty := parseImportString(importString)

	fmt.Printf("Kind:%s, Namespace:%s, resourceName:%s, SubKind:%s, FilterPredicate:%s", kind, namespace, resourceName, subKind, filterPredicate)

	jsonData := QueryCompositionEndpoint(kind, namespace, resourceName)
	fmt.Printf("Queried KubeDiscovery: %s\n", string(jsonData))
	name, err := ParseDiscoveryJSON(jsonData, subKind, filterPredicate)
	if err != nil {
		return "", err
	}
	fmt.Printf("Found name: %s\n", name)

	if specProperty != "" {
		specPropertyValue := resolveSpecProperty(namespace, subKind, name, specProperty)
		return specPropertyValue, nil
	} else {
		return name, nil
	}
}

func ResolveKind(importString string) (string, error) {
	return "Not implemented yet.", nil
}

func resolveSpecProperty(namespace, subKind, name, specProperty string) (string) {
	resourceDetails := queryResourceDetailsEndpoint(subKind, name, namespace)
	propertyValues := make([]string, 0)
	propertyValuesPtr := &propertyValues
	parsePropertyValue(resourceDetails, specProperty, propertyValuesPtr)
	fmt.Printf("PropertyValues:%v\n", propertyValues)
	for _, propertyValue := range propertyValues {
		fmt.Printf("Property Value:%s\n", propertyValue)
	}
	// There might be multiple properties in Spec of a resource. 
	// Right now returning the first property value.
	// TODO: We can support "indexing" to select a specific property value to return
	// Example: MysqlCluster:default.cluster1:StatefulSet.mountPath[0] will return first container's mounthPath
	// whereas MysqlCluster:default.cluster1:StatefulSet.mountPath[1] will return second container's mountPath, etc.
	if len(propertyValues) > 0 {
		return propertyValues[0]
	}
	return ""
}

func parsePropertyValue(resourceDetails []byte, specProperty string, propertyValues *[]string) {
	jsonparser.ObjectEach(resourceDetails, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if dataType.String() == "string" {
			//fmt.Printf("Key:%s, Value:%s\n", key, value)
			if string(key) == specProperty {
				*propertyValues = append(*propertyValues, string(value))
			}
		} 
		if dataType.String() == "array" {
			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				parsePropertyValue(value, specProperty, propertyValues)
			})
		}
		if dataType.String() == "object" {
			parsePropertyValue(value, specProperty, propertyValues)
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
	return "", fmt.Errorf("Name of resource was not found. Resource Kind : %s with composition data: %s", subKind, string(composition))
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

func QueryCompositionEndpoint(kind, namespace, crdKindName string) []byte {
	args := fmt.Sprintf("kind=%s&instance=%s&namespace=%s", kind, crdKindName, namespace)
	serviceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	servicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	var url1 string
	url1 = fmt.Sprintf("https://%s:%s/apis/platform-as-code/v1/composition?%s", serviceHost, servicePort, args)
	body := queryAPIServer(url1)
	return body
}

func queryResourceDetailsEndpoint(kind, name, namespace string) []byte {
	args := fmt.Sprintf("kind=%s&instance=%s&namespace=%s", kind, name, namespace)
	serviceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	servicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	var url1 string
	url1 = fmt.Sprintf("https://%s:%s/apis/platform-as-code/v1/resourceDetails?%s", serviceHost, servicePort, args)
	body := queryAPIServer(url1)
	return body
}

func queryAPIsEndpoint(kind, namespace, crdKindName string) []byte {
	args := fmt.Sprintf("kind=%s&instance=%s&namespace=%s", kind, crdKindName, namespace)
	serviceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	servicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	var url1 string
	url1 = fmt.Sprintf("https://%s:%s/apis?%s", serviceHost, servicePort, args)
	body := queryAPIServer(url1)
	return body
}

// Used to query KubeDiscovery api server
func queryAPIServer(url1 string) []byte {
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
