package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/buger/jsonparser"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kubernetes/pkg/apis/core/v1"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

const (
	admissionWebhookAnnotationValidateKey = "admission-webhook-example.banzaicloud.com/validate"
	admissionWebhookAnnotationMutateKey   = "admission-webhook-example.banzaicloud.com/mutate"
	admissionWebhookAnnotationStatusKey   = "admission-webhook-example.banzaicloud.com/status"
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
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
	annotations.KindToEntry = make(map[string][]Entry, 0)

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
			hasFunc := strings.Contains(val, "Fn::ImportValue")
			if hasFunc {
				start := len("Fn::ImportValue(")
				end := strings.LastIndex(val, ")")
				annotationPath := val[start:end]
				fmt.Println(annotationPath)
				needResolve := ResolveData{
					JSONTreePath:   jsonPath,
					AnnotationPath: annotationPath,
					FunctionType:   ImportValue,
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

// main mutation process
func (whsvr *WebhookServer) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request

	fmt.Println(req.Kind.Kind)
	allAnnotations, _, _, err := jsonparser.Get(req.Object.Raw, "metadata", "annotations")
	fmt.Printf("Adding annotations: %s\n", string(allAnnotations))
	var patchOperations []patchOperation
	patchOperations = make([]patchOperation, 0)
	if err == nil { //annotations exist, parse them into our data structure
		name, err := jsonparser.GetUnsafeString(req.Object.Raw, "metadata", "name")
		if err != nil {
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		kind, err := jsonparser.GetUnsafeString(req.Object.Raw, "kind")
		if err != nil {
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		jsonparser.ObjectEach(allAnnotations, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			entry := Entry{InstanceName: name, Key: string(key), Value: string(value)}

			var entryList []Entry
			var kindExists bool
			if entryList, kindExists = annotations.KindToEntry[kind]; !kindExists {
				entryList = make([]Entry, 0)
			}
			if annotations.Exists(entry, kind) {
				return nil
			}
			entryList = append(entryList, entry)
			annotations.KindToEntry[kind] = entryList
			return nil
		})
		fmt.Println("----- Stored data: -----")
		fmt.Println(annotations.KindToEntry)

		forResolve := ParseJson(req.Object.Raw)

		for i := 0; i < len(forResolve); i++ {
			var resolveObj ResolveData
			resolveObj = forResolve[i]
			annotationPath := resolveObj.AnnotationPath
			kind, instanceName, key, err := splitAnnotationPatch(annotationPath)
			if err != nil {
				return &v1beta1.AdmissionResponse{
					Result: &metav1.Status{
						Message: err.Error(),
					},
				}
			}
			value, err := searchAnnotation(annotations.KindToEntry[kind], instanceName, key)
			if err != nil {
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
		}
	}
	patchBytes, _ := json.Marshal(patchOperations)

	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func searchAnnotation(entries []Entry, instanceName, key string) (string, error) {
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if entry.InstanceName == instanceName &&
			entry.Key == key {
			return entry.Value, nil
		}
	}
	// Could not find the data
	fmt.Printf("instance name: %s key: %s \n", instanceName, key)
	return "", fmt.Errorf("Error annotation data was not stored, nothing to replace with. Check path.")
}
func splitAnnotationPatch(annotationPath string) (string, string, string, error) {
	path := strings.Split(annotationPath, ".")
	if len(path) != 3 {
		return "", "", "", fmt.Errorf("AnnotationPath inside function is not of len 3")
	}
	return path[0], path[1], path[2], nil
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
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("Can't decode body: %v", err)
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
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
