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

	"github.com/buger/jsonparser"
)

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
				fmt.Println("****************")

				// listOptions := metav1.ListOptions{
				// 	FieldSelector: fmt.Sprintf("metadata.name=%s", name),
				// }
				// switch subKind {
				// case "Deployment":
				// 	deployment := kubeClient.AppsV1().Deployments(namespace)
				// }
				needResolve := ResolveData{
					JSONTreePath:   jsonPath,
					AnnotationPath: name,
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

func ParseDiscoveryJSON(composition []byte, subKind string) (string, error) {
	var subName string
	var err error
	jsonparser.ArrayEach(composition, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		subName, err = ParseDiscoveryJSONHelper(composition, subKind, "")
	})
	return subName, err
}

// Finds the name of the subresource of a CRD resource
func ParseDiscoveryJSONHelper(composition []byte, subKind string, subName string) (string, error) {
	if len(subName) > 0 {
		return subName, nil
	}
	jsonparser.ObjectEach(composition, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		var name string
		if dataType.String() == "object" {
			fmt.Printf("key: %s\n", string(key))
			fmt.Printf("value: %s\n", string(value))
			kind1, _ := jsonparser.GetUnsafeString(value, "Kind")
			name1, _ := jsonparser.GetUnsafeString(value, "Name")
			fmt.Printf("kind: %s\n", kind1)
			fmt.Printf("name: %s\n\n", name1)

			if kind, err := jsonparser.GetUnsafeString(value, "Kind"); err != nil && kind == subKind {
				name, _ = jsonparser.GetUnsafeString(value, "Name")
				fmt.Printf("name %s\n", name)
				fmt.Printf("sname %s\n", subName)
				return nil
			}
			ParseDiscoveryJSONHelper(value, subKind, name)
		} else {
			return nil
		}
		return nil
	})
	return "", fmt.Errorf("Could not find a name for kind")
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
