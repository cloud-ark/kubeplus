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
// https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
func AddDeploymentLabel(key, value string, name, namespace string) {
	deployClient := kubeClient.AppsV1().Deployments(namespace)

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		deployment, err := deployClient.Get(name, metav1.GetOptions{})
		if err != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", err))
		}
		deployment.ObjectMeta.Labels[key] = value
		_, updateErr := deployClient.Update(deployment)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
}

func ParseJson(data []byte) []ResolveData {
	needResolving := make([]ResolveData, 0)
	stringStack := StringStack{Data: "", Mutex: sync.Mutex{}}
	ParseJsonHelper(data, &needResolving, &stringStack)
	return needResolving
}

func ParseJsonHelper(data []byte, needResolving *[]ResolveData, stringStack *StringStack) {
	jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if dataType.String() == "object" {
			stringStack.Push(string(key))
			ParseJsonHelper(value, needResolving, stringStack)
			stringStack.Pop()
		} else {
			stringStack.Push(string(key))
			jsonPath := stringStack.Peek()
			val := string(value)
			hasImportFunc := strings.Contains(val, "Fn::ImportValue")
			hasLabelFunc := strings.Contains(val, "Fn::AddLabel")
			if hasImportFunc {
				start := len("Fn::ImportValue(")
				end := strings.LastIndex(val, ")")
				annotationPath := val[start:end]
				needResolve := ResolveData{
					JSONTreePath:   jsonPath,
					AnnotationPath: annotationPath,
					FunctionType:   ImportValue,
				}
				*needResolving = append(*needResolving, needResolve)
				stringStack.Pop()
			} else if hasLabelFunc {
				start := len("Fn::AddLabel(")
				end := strings.LastIndex(val, ")")
				args := strings.Split(val[start:end], ",")
				fmt.Printf("args: %s %s\n", args[0], args[1])

				keyValLabel := strings.TrimSpace(args[0])
				keyVal := strings.Split(keyValLabel, "/")
				key := keyVal[0]
				val := keyVal[1]
				fmt.Printf("LABELLLL%s\n", keyValLabel)
				namespace, kind, crdKindName, subKind, err := ParseCompositionPath(args[1])
				fmt.Printf("parsed: %s %s %s %s\n", namespace, kind, crdKindName, subKind)
				jsonData := QueryAPIServer(kind, namespace, crdKindName)
				fmt.Printf("Results: %s\n", string(jsonData))
				if err != nil {
					return err
				}
				name, err := ParseDiscoveryJSON(jsonData, subKind)
				fmt.Println("****************")
				fmt.Println(name)
				fmt.Println(err)
				resourceName := name[0]
				fmt.Println("****************")
				switch subKind {
				case "Deployment":

					AddDeploymentLabel(key, val, name[0], namespace)

				}
				//found one resource that matches "Deployment"
				needResolve := ResolveData{
					JSONTreePath:   jsonPath,
					AnnotationPath: resourceName,
					FunctionType:   AddLabel,
				}
				*needResolving = append(*needResolving, needResolve)
				stringStack.Pop()
			} else {
				stringStack.Pop()
			}
		}
		return nil
	})
}

// [Namespace]?.[Kind].[InstanceName].[outputVariable]
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

// [namespace]?.[Kind].[CrdKindName].[SubKind]
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

func ParseDiscoveryJSON(composition []byte, subKind string) ([]string, error) {
	names := make([]string, 0)
	// note that for the recursive style I do, I must pass a ptr value to the function
	// the logic is that are that inside of ObjectEach or ArrayEach I cannot return from
	// ParseDiscoveryJSON, since I would actually be returning from a lamba func that is
	// defined to return err by the docs. Note you cannot have a string pointer in Go, and
	// an array actually makes more sense when you consider the kubediscovery code
	// (it can return multiple resources I believe)

	// So the solution is to do a recursive style where I append the data and
	// pass it along. This can be seen in the efficient fibonacci tail recursion example
	// for example.
	ptr := &names
	var err error

	// output from kubediscovery could be multiple resources...
	// even if it is just one, have to loop through array to get to the {}byte part
	jsonparser.ArrayEach(composition, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		ParseDiscoveryJSONHelper(value, subKind, ptr)
	})
	fmt.Printf("subname: %v \n", *ptr)
	return *ptr, err
}

// Finds the name of the subresource of a CRD resource
func ParseDiscoveryJSONHelper(composition []byte, subKind string, subName *[]string) error {
	jsonparser.ObjectEach(composition, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if dataType.String() == "array" {
			// go through each children object.
			jsonparser.ArrayEach(value, func(value1 []byte, dataType jsonparser.ValueType, offset int, err error) {
				// Debug prints
				// kind1, _ := jsonparser.GetUnsafeString(value1, "Kind")
				// name1, _ := jsonparser.GetUnsafeString(value1, "Name")
				// fmt.Printf("kind: %s\n", kind1)
				// fmt.Printf("name: %s\n", name1)

				// each object in Children array is a json object
				// looking like {Kind: deployment, name: moodle1, level: 2, namespace: default}

				// if Kind key exists and it is equal to subKind, we found a resource name
				// that matches the type. It may be possible to have multiple Ingresses or Deployments?
				// so I return a list of strings, usually it will return one Name only. Handle by func caller.
				if kind, err := jsonparser.GetUnsafeString(value1, "Kind"); err == nil && kind == subKind {
					name, _ := jsonparser.GetUnsafeString(value1, "Name")
					fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!! %s\n", name)
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

func QueryAPIServer(kind, namespace, crdKindName string) []byte {
	serviceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	servicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	args := fmt.Sprintf("kind=%s&instance=%s&namespace=%s", kind, crdKindName, namespace)
	var url1 string
	url1 = fmt.Sprintf("https://%s:%s/apis/platform-as-code/v1/composition?%s", serviceHost, servicePort, args)
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
	//fmt.Println("Exiting queryAPIServer")
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
