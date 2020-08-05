package main

// For the random() label function argument
import (
	"fmt"
	"strings"
	"sync"
)

type labelfunction func() string

// Function ...
// Enum declaration
type Function int

// An enum for supported functions
const (
	ImportValue Function = 0
	AddLabel    Function = 1
	AddAnnotation Function = 2
)

// ResolveData ...
// Type used to store the data needed to resolve each Fn::
// Creates a list of these in ParseJson
// JSONTreePath: this is the path into the json object sent by kubernetes api
//  this is found recursively while we search the json object for Fn:: declarations.
//  and then used later in Jsonpatch object. (it needs this to know where to replace)
// ImportString: This is the string that represents values to be imported using ImportValue stored in the data structure
//  [Namespace]?.[Kind].[InstanceName].[outputVariable]
//  used for the ImportValue functions
// Value:
//  this is used for the addlabel functions since we don't need to use annotationpath
//  as the value we will replace it with are provided in the val of key/val function
// FunctionType
//  this is an enum of the functions we support for Fn:: in the yamls

type ResolveData struct {
	JSONTreePath   string
	ImportString string
	FunctionType   Function
	Value          string
}

type PlatformStackData struct {
	Name string
	Namespace string
}

type StackElementData struct {
	Name string
	Kind string
	Namespace string
}

// StringStack...
// This is a stack I implemented that we use when recursively parsing the json
// to track where we are in the json tree.
// We use this to populate the JSONTreePath variable.
type StringStack struct {
	Data  string
	Mutex sync.Mutex
}


var (

	KindPluralMap  map[string]string
	kindVersionMap map[string]string
	kindGroupMap map[string]string

	REPLICA_SET  string
	DEPLOYMENT   string
	POD          string
	CONFIG_MAP   string
	SERVICE      string
	SECRET       string
	PVCLAIM      string
	PV           string
	ETCD_CLUSTER string
	INGRESS      string
	STATEFULSET  string
	DAEMONSET    string
	RC           string
	PDB 		 string
)

func init() {

	DEPLOYMENT = "Deployment"
	REPLICA_SET = "ReplicaSet"
	POD = "Pod"
	CONFIG_MAP = "ConfigMap"
	SERVICE = "Service"
	SECRET = "Secret"
	PVCLAIM = "PersistentVolumeClaim"
	PV = "PersistentVolume"
	ETCD_CLUSTER = "EtcdCluster"
	INGRESS = "Ingress"
	STATEFULSET = "StatefulSet"
	DAEMONSET = "DaemonSet"
	RC = "ReplicationController"
	PDB = "PodDisruptionBudget"

	KindPluralMap = make(map[string]string)
	kindVersionMap = make(map[string]string) 
	kindGroupMap = make(map[string]string)
	
	KindPluralMap[DEPLOYMENT] = "deployments"
	kindVersionMap[DEPLOYMENT] = "apis/apps/v1"
	kindGroupMap[DEPLOYMENT] = "apps"

	KindPluralMap[REPLICA_SET] = "replicasets"
	kindVersionMap[REPLICA_SET] = "apis/apps/v1"
	kindGroupMap[REPLICA_SET] = "apps"

	KindPluralMap[DAEMONSET] = "daemonsets"
	kindVersionMap[DAEMONSET] = "apis/apps/v1"
	kindGroupMap[DAEMONSET] = "apps"

	KindPluralMap[RC] = "replicationcontrollers"
	kindVersionMap[RC] = "api/v1"
	kindGroupMap[RC] = ""

	KindPluralMap[PDB] = "poddisruptionbudgets"
	kindVersionMap[PDB] = "apis/policy/v1beta1"
	kindGroupMap[PDB] = "policy"

	KindPluralMap[POD] = "pods"
	kindVersionMap[POD] = "api/v1"
	kindGroupMap[POD] = ""

	KindPluralMap[SERVICE] = "services"
	kindVersionMap[SERVICE] = "api/v1"
	kindGroupMap[SERVICE] = ""

	KindPluralMap[INGRESS] = "ingresses"
	kindVersionMap[INGRESS] = "networking.k8s.io/v1beta1"//"extensions/v1beta1"
	kindGroupMap[INGRESS] = "networking.k8s.io"

	KindPluralMap[SECRET] = "secrets"
	kindVersionMap[SECRET] = "api/v1"
	kindGroupMap[SECRET] = ""

	KindPluralMap[PVCLAIM] = "persistentvolumeclaims"
	kindVersionMap[PVCLAIM] = "api/v1"
	kindGroupMap[PVCLAIM] = ""

	KindPluralMap[PV] = "persistentvolumes"
	kindVersionMap[PV] = "api/v1"
	kindGroupMap[PV] = ""

	KindPluralMap[STATEFULSET] = "statefulsets"
	kindVersionMap[STATEFULSET] = "apis/apps/v1"
	kindGroupMap[STATEFULSET] = "apps"

	KindPluralMap[CONFIG_MAP] = "configmaps"
	kindVersionMap[CONFIG_MAP] = "api/v1"
	kindGroupMap[CONFIG_MAP] = ""
}

func getKindAPIDetails(kind string) (string, string, string, string) {
	kindplural := KindPluralMap[kind]
	kindResourceApiVersion := kindVersionMap[kind]
	kindResourceGroup := kindGroupMap[kind]

	parts := strings.Split(kindResourceApiVersion, "/")
	kindAPI := parts[len(parts)-1]

	return kindplural, kindResourceApiVersion, kindAPI, kindResourceGroup
}


func (s *StringStack) Len() int {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	return len(s.Data)
}
func (s *StringStack) Push(key string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Data = fmt.Sprintf("%s%s%s", s.Data, "/", key)
}
func (s *StringStack) Pop() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if len(s.Data) == 0 {
		return
	}
	ind := strings.LastIndex(s.Data, "/")
	s.Data = s.Data[:ind]
}
func (s *StringStack) Peek() string {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	return s.Data
}

// Entry ...
// This is a single Entry in our data structure: map[<CrdKind>]-> Entries
// we use this to store annotation data. Then we search each of these when
// doing Fn::ImportValue
type Entry struct {
	InstanceName string
	Namespace    string
	Key          string
	Value        string
}

// StoredAnnotations ...
// data structure: map[<CrdKind>]-> Entries
// maps kind to list of entries
type StoredAnnotations struct {
	KindToEntry map[string][]Entry
}

func (a *StoredAnnotations) Exists(e Entry, kind string) bool {
	var entryList []Entry
	var kindExists bool
	if entryList, kindExists = annotations.KindToEntry[kind]; !kindExists {
		return false
	}
	for i := 0; i < len(entryList); i++ {
		entry := entryList[i]
		if strings.EqualFold(entry.InstanceName, e.InstanceName) &&
			strings.EqualFold(entry.Key, e.Key) &&
			strings.EqualFold(entry.Value, e.Value) &&
			strings.EqualFold(entry.Namespace, e.Namespace) {
			return true
		}
	}
	return false
}

func (a *StoredAnnotations) Delete(e Entry, kind string) bool {
	var entryList []Entry
	var kindExists bool
	if entryList, kindExists = annotations.KindToEntry[kind]; !kindExists {
		fmt.Println("Could not delete bc kind does not exist.")
		return false
	}
	var indexToDelete int
	for i := 0; i < len(entryList); i++ {
		entry := entryList[i]
		if strings.EqualFold(entry.InstanceName, e.InstanceName) &&
			strings.EqualFold(entry.Value, e.Value) &&
			strings.EqualFold(entry.Namespace, e.Namespace) &&
			strings.EqualFold(entry.Key, e.Key) {
			indexToDelete = i
			break
		}
	}
	le := len(annotations.KindToEntry[kind])
	annotations.KindToEntry[kind][indexToDelete] = annotations.KindToEntry[kind][le-1] //swap to last
	annotations.KindToEntry[kind][le-1] = Entry{}                                      //write zero value
	annotations.KindToEntry[kind] = annotations.KindToEntry[kind][:le-1]               //truncate
	return true
}
