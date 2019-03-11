package etcd_helper

import (
	"context"
	"encoding/json"
	"log"

	"github.com/coreos/etcd/client"
)

// Etcdreader ...
// Performs read operations on an Etcd database
type Etcdreader struct {
	EtcdServiceURL string
}

// Etcdwriter ...
// Performs write operations on an Etcd database
type Etcdwriter struct {
	EtcdServiceURL string
}

// getClient
// Internal helper method creating a connection to an etcd database
// at a specific etcd URL.
func getClient(etcdServiceURL string) (client.Client, error) {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	return c, err
}

// Get ...
// Method to Get a single value for a key
func (o Etcdreader) Get(resourceKey string) (string, error) {
	c, err := getClient(o.EtcdServiceURL)
	if err != nil {
		return "", err
	}
	kapi := client.NewKeysAPI(c)

	resp, err := kapi.Get(context.Background(), resourceKey, nil)
	if err != nil {
		return "", err
	}
	// (danielpygo): TODO
	// ADD LOGLEVEL ...
	return resp.Node.Value, nil
}

// GetList ...
// Method to Get a list of values for a key
func (o Etcdreader) GetList(resourceKey string) ([]string, error) {
	c, err := getClient(o.EtcdServiceURL)
	kapi := client.NewKeysAPI(c)

	var currentListString string
	var currentList []string

	resp, err1 := kapi.Get(context.Background(), resourceKey, nil)
	if err1 != nil {
		return nil, err1
	} else {
		currentListString = resp.Node.Value
		if err = json.Unmarshal([]byte(currentListString), &currentList); err != nil {
			return nil, err
		}
	}
	return currentList, nil
}

// GetMap ...
// Method to Get a map for a key
func (o Etcdreader) GetMap(resourceKey string) (map[string]interface{}, error) {
	c, err := getClient(o.EtcdServiceURL)
	kapi := client.NewKeysAPI(c)

	valuesMap := make(map[string]interface{}, 0)
	var currentMapString string

	resp, err1 := kapi.Get(context.Background(), resourceKey, nil)
	if err1 != nil {
		return nil, err1
	} else {
		currentMapString = resp.Node.Value
		if err = json.Unmarshal([]byte(currentMapString), &valuesMap); err != nil {
			return nil, err
		}
	}
	return valuesMap, nil
}

// Store ...
// Method to Store a value for a key
func (o Etcdwriter) Store(resourceKey, resourceData string) error {
	c, err := getClient(o.EtcdServiceURL)
	if err != nil {
		log.Fatal(err)
		return err

	}
	kapi := client.NewKeysAPI(c)

	_, err = kapi.Set(context.Background(), resourceKey, resourceData, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

// StoreList ...
// Method to Store a list into a key
func (o Etcdwriter) StoreList(resourceKey string, dataList []string) error {
	c, err := getClient(o.EtcdServiceURL)
	if err != nil {
		log.Fatal(err)
		return err
	}
	kapi := client.NewKeysAPI(c)

	jsonList, err := json.Marshal(&dataList)
	if err != nil {
		log.Fatal(err)
		return err
	}
	resourceVal := string(jsonList)

	_, err = kapi.Set(context.Background(), resourceKey, resourceVal, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

// StoreMap ...
// Method to Store a Map for a key in json format
func (o Etcdwriter) StoreMap(resourceKey string, resourceData map[string]interface{}) error {
	c, err := getClient(o.EtcdServiceURL)
	if err != nil {
		log.Fatal(err)
		return err
	}
	kapi := client.NewKeysAPI(c)

	jsonData, err := json.Marshal(&resourceData)
	if err != nil {
		log.Fatal(err)
		return err
	}
	jsonDataString := string(jsonData)
	_, err = kapi.Set(context.Background(), resourceKey, jsonDataString, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return err
}

// Delete ...
// Method to Delete a key/val pair from a list of strings or a single string
func (o Etcdwriter) Delete(resourceKey, resourceValue string) error {
	c, err := getClient(o.EtcdServiceURL)
	if err != nil {
		return err
	}
	kapi := client.NewKeysAPI(c)

	var r Etcdreader
	r.EtcdServiceURL = o.EtcdServiceURL
	valueList, err := r.GetList(resourceKey)
	if err == nil {
		if len(valueList) == 1 {
			//When using StoreList for one elem, or if a List gets reduced to one elem
			//We need to try and delete both the key and val with kapi.Delete Or else
			// resourceKey gets set to "null"
			return o.deleteSingleElem(resourceKey, "[\""+resourceValue+"\"]", kapi)
		}
		var newList []string
		for _, val := range valueList {
			if val != resourceValue {
				newList = append(newList, val)
			}
		}
		jsonValueList, err := json.Marshal(&newList)

		if err != nil {
			log.Fatal(err)
			return err
		}
		newResourceValue := string(jsonValueList)

		_, err = kapi.Set(context.Background(), resourceKey, newResourceValue, nil)
		if err != nil {
			log.Fatal(err)
			return err
		}
	} else {
		_, err := r.Get(resourceKey)
		if err != nil {
			log.Fatal(err)
			return err
		}
		return o.deleteSingleElem(resourceKey, resourceValue, kapi)
	}
	return nil
}

// Delete ...
// Method to Delete a single key/value pair
// If it is successful key gets deleted as well.
func (o Etcdwriter) deleteSingleElem(resourceKey, resourceValue string, kapi client.KeysAPI) error {
	//Deletes the key and val, iff value is val
	_, err := kapi.Delete(context.Background(), resourceKey, &client.DeleteOptions{PrevValue: resourceValue})
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
