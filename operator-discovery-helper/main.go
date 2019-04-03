package main

import (
	"flag"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/client"
	"log"
	"time"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
    "k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
	etcdServiceURL string
)

func getCRDDetailsFromAPIServer() error {

		etcdServiceURL = "http://localhost:2379"

		cfg, _ := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
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
			//fmt.Printf("Object Meta:%v\n", objectMeta)
			//name := objectMeta.GetName()
			//namespace := objectMeta.GetNamespace()
			annotations := objectMeta.GetAnnotations()

			//composition := make([]string, 0)
			compositionString := annotations["composition"]
			/*composition1 := strings.Split(compositionString, ",")
			for _, elem := range composition1 {
				elem = strings.TrimSpace(elem)
				composition = append(composition, elem)
			}
			*/

			//fmt.Printf("Group:%s, Version:%s, Kind:%s, Plural:%s, Endpoint:%s, Composition:%s\n", 
			//	group, version, kind, plural, endpoint, composition)

			var crdDetailsMap = make(map[string]interface{})
			crdDetailsMap["kind"] = kind
			crdDetailsMap["endpoint"] = endpoint
			crdDetailsMap["plural"] = plural
			crdDetailsMap["composition"] = compositionString

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

	//fmt.Printf("Setting %s->%s\n",resourceKey, jsonDataString)
	_, err1 := kapi.Set(context.Background(), resourceKey, jsonDataString, nil)
	if err1 != nil {
		log.Fatal(err1)
	} else {
		//log.Printf("Set is done. Metadata is %q\n", resp.Node.Value)
	}
}


func main() {
	flag.Parse()
	for {
		getCRDDetailsFromAPIServer()
		time.Sleep(time.Second * 2)
	}
}
