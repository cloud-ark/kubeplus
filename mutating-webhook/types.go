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
type ResolveData struct {
	JSONTreePath   string
	AnnotationPath string
	FunctionType   Function
	Argument       labelfunction
}

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

type Entry struct {
	InstanceName string
	Key          string
	Value        string
}

//maps kind to list of entries
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
		if entry.InstanceName == e.InstanceName &&
			entry.Value == e.Value {
			return true
		}
	}
	return false
}
