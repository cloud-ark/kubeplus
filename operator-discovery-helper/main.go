package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/client"
	"log"
	"time"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

var (
	etcdServiceURL string
)

func getCRDDetailsFromAPIServer() error {

	etcdServiceURL = "http://localhost:2379"

	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	crdClient, _ := apiextensionsclientset.NewForConfig(cfg)

	crdList, err := crdClient.CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Error:%s\n", err)
		return err
	}
	for _, crd := range crdList.Items {
		crdName := crd.ObjectMeta.Name
		crdObj, err := crdClient.CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
		if err != nil {
			fmt.Errorf("Error:%s\n", err)
			return err
		}
		group := crdObj.Spec.Group
		version := crdObj.Spec.Version
		endpoint := "apis/" + group + "/" + version
		kind := crdObj.Spec.Names.Kind
		plural := crdObj.Spec.Names.Plural

		objectMeta := crdObj.ObjectMeta
		//name := objectMeta.GetName()
		//namespace := objectMeta.GetNamespace()
		annotations := objectMeta.GetAnnotations()

		var crdDetailsMap = make(map[string]interface{})
		crdDetailsMap["kind"] = kind
		crdDetailsMap["endpoint"] = endpoint
		crdDetailsMap["plural"] = plural
		crdDetailsMap["composition"] = annotations["platform-as-code/composition"]
		crdDetailsMap["constants"] = annotations["platform-as-code/constants"]
		crdDetailsMap["usage"] = annotations["platform-as-code/usage"]
		crdDetailsMap["openapispec"] = annotations["platform-as-code/openapispec"]

		//crdName := "postgreses.postgrescontroller.kubeplus"
		storeEtcd("/crds/"+crdName, crdDetailsMap)
	}
	return nil
}

func storeEtcd(resourceKey string, resourceData interface{}) {
	jsonData, err := json.Marshal(&resourceData)
	if err != nil {
		panic(err)
	}
	jsonDataString := string(jsonData)
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	_, err1 := kapi.Set(context.Background(), resourceKey, jsonDataString, nil)
	if err1 != nil {
		log.Fatal(err1)
	}
}

func main() {
	flag.Parse()
	for {
		getCRDDetailsFromAPIServer()
		time.Sleep(time.Second * 2)
	}
}
