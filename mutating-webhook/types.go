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
)

// ResolveData ...
// Type used to store the data needed to resolve each Fn::
// Creates a list of these in ParseJson
// JSONTreePath: this is the path into the json object sent by kubernetes api
//  this is found recursively while we search the json object for Fn:: declarations.
//  and then used later in Jsonpatch object. (it needs this to know where to replace)
// AnnotationPath: This is the path to the annotation stored in the data structure
//  [Namespace]?.[Kind].[InstanceName].[outputVariable]
//  used for the ImportValue functions
// Value:
//  this is used for the addlabel functions since we don't need to use annotationpath
//  as the value we will replace it with are provided in the val of key/val function
// FunctionType
//  this is an enum of the functions we support for Fn:: in the yamls

type ResolveData struct {
	JSONTreePath   string
	AnnotationPath string
	FunctionType   Function
	Value          string
}

// StringStack...
// This is a stack I implemented that we use when recursively parsing the json
// to track where we are in the json tree.
// We use this to populate the JSONTreePath variable.
type StringStack struct {
	Data  string
	Mutex sync.Mutex
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
