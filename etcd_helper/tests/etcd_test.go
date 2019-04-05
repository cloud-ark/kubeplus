package tests

import (
	"fmt"
	"testing"

	"github.com/cloud-ark/kubeplus/etcd_helper"
	"github.com/stretchr/testify/assert"
)

var (
	rdr  etcd_helper.Etcdreader
	wrtr etcd_helper.Etcdwriter
)

func init() {
	rdr.EtcdServiceURL = "http://localhost:2379"
	wrtr.EtcdServiceURL = "http://localhost:2379"
}

func TestGet(t *testing.T) {
	err := wrtr.Store("postgres", "THIS IS SOME DATA")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	str, err := rdr.Get("postgres")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	assert.Equal(t, str, "THIS IS SOME DATA", "A simple Post and Get of one item")
}
func TestGetList(t *testing.T) {
	dataList := []string{"database1", "database2", "database3"}
	err := wrtr.StoreList("postgres1", dataList)
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	strSlice, err := rdr.GetList("postgres1")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	assert.Equal(t, strSlice, dataList, "A simple Post and Get of an array of dataItems")
}

func TestDeleteFromList(t *testing.T) {
	dataList := []string{"database1", "database2", "database3"}
	err := wrtr.StoreList("postgres2", dataList)
	if err != nil {
		fmt.Printf("Err %s\n", err)
		t.Fail()
	}
	err = wrtr.Delete("postgres2", "database3")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	strSlice, err := rdr.GetList("postgres2")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	expectedDataList := []string{"database1", "database2"}
	assert.Equal(t, expectedDataList, strSlice, "A simple Delete operation")
}

func TestDeleteKey(t *testing.T) {
	err := wrtr.Store("postgres3", "somedata")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	str, err := rdr.Get("postgres3")

	assert.Equal(t, str, "somedata", "Ensure it was stored")

	err = wrtr.Delete("postgres3", "somedata")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	str, err = rdr.Get("postgres3")

	assert.Equal(t, err.Error()[0:18], "100: Key not found", "The key should be deleted")
	assert.Equal(t, "", str, "A Delete of one key/val pair stored using Store")
}

func TestDeleteSingleValWithKey(t *testing.T) {
	dataList := []string{"database1"}
	wrtr.StoreList("postgres4", dataList)

	_, err := rdr.Get("postgres4")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	err = wrtr.Delete("postgres4", "database1")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	str, err := rdr.Get("postgres4")
	assert.Equal(t, err.Error()[0:18], "100: Key not found", "The key should be deleted")
	assert.Equal(t, "", str, "A Delete of a key/val pair stored as a list with StoreList")
}

func TestStoreMap(t *testing.T) {
	dataMap := make(map[string]interface{}, 0)
	dataMap["someKey"] = "someValue"

	wrtr.StoreMap("chartValues", dataMap)

	myMap, err := rdr.GetMap("chartValues")

	if err != nil {
		fmt.Printf("Err: %s\n", err)
		t.Fail()
	}
	v, ok := myMap["someKey"]
	if !ok {
		fmt.Printf("Err: %s\n", "Did not properly store a map")
		t.Fail()
	}
	myVal, ok := v.(string)
	if !ok {
		fmt.Printf("Err: %s\n", "Did not properly store a map")
		t.Fail()
	}
	assert.Equal(t, "someValue", myVal, "The retrieved data from etcd must match the Map that is stored")
}
