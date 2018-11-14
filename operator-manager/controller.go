/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/client"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"

	"io"
	"net/http"

	"bytes"
	"compress/gzip"
	"github.com/mholt/archiver"

	operatorv1 "github.com/cloud-ark/kubeplus/operator-manager/pkg/apis/operatorcontroller/v1"
	clientset "github.com/cloud-ark/kubeplus/operator-manager/pkg/client/clientset/versioned"
	operatorscheme "github.com/cloud-ark/kubeplus/operator-manager/pkg/client/clientset/versioned/scheme"
	informers "github.com/cloud-ark/kubeplus/operator-manager/pkg/client/informers/externalversions"
	listers "github.com/cloud-ark/kubeplus/operator-manager/pkg/client/listers/operatorcontroller/v1"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

const controllerAgentName = "operator-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

var (
	etcdServiceURL        string
	openAPISpecKey        = "/openAPISpecRegistered"
	operatorsToDeleteKey  = "/operatorsToDelete"
	operatorsToInstallKey = "/operatorsToInstall"
	chartValuesKey        = "/chartvalues"
)

func init() {
	etcdServiceURL = "http://localhost:2379"
}

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	foosLister        listers.OperatorLister
	foosSynced        cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	sampleInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Deployment and Foo
	// types.
	deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()
	fooInformer := sampleInformerFactory.Operatorcontroller().V1().Operators()

	// Create event broadcaster
	// Add operator-controller types to the default Kubernetes Scheme so Events can be
	// logged for operator-controller types.
	operatorscheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		foosLister:        fooInformer.Lister(),
		foosSynced:        fooInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Operators"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	fooInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFoo,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*operatorv1.Operator)
			oldDepl := old.(*operatorv1.Operator)
			//fmt.Println("New Version:%s", newDepl.ResourceVersion)
			//fmt.Println("Old Version:%s", oldDepl.ResourceVersion)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			} else {
				controller.enqueueFoo(new)
			}
		},
		DeleteFunc: func(obj interface{}) {
			//_, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			//if err == nil {
			controller.deleteOperator(obj)
			//}
		},
	})
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting Foo controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.foosSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueFoo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Foo" {
			return
		}

		foo, err := c.foosLister.Operators(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueFoo(foo)
		return
	}
}

func (c *Controller) deleteOperator(obj interface{}) {
	var err error
	if _, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		panic(err)
	}

	foo := obj.(*operatorv1.Operator)

	fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

	operatorName := foo.Spec.Name
	fmt.Printf("Operator to delete:%s\n", operatorName)

	operatorChartURL := foo.Spec.ChartURL

	storeChartURL(operatorsToDeleteKey, operatorChartURL)

	deleteConfigMap(operatorChartURL, c.kubeclientset)

	deleteChartURL(openAPISpecKey, operatorChartURL)

	c.workqueue.Forget(obj)
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Foo resource with this namespace/name
	foo, err := c.foosLister.Operators(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	fmt.Println("**************************************")

	operatorName := foo.Spec.Name
	operatorChartURL := foo.Spec.ChartURL
	operatorChartValues := foo.Spec.Values

	fmt.Printf("Operator Name:%s\n", operatorName)
	fmt.Printf("Chart URL:%s\n", operatorChartURL)
	fmt.Printf("Values:%v\n", operatorChartValues)

	storeChartURL(operatorsToInstallKey, operatorChartURL)

	storeEtcd(chartValuesKey+"/"+operatorChartURL, operatorChartValues)

	var operatorCRDString string
	for {
		operatorCRDString = getOperatorCRDs(operatorChartURL)
		if operatorCRDString != "" {
			break
		}
		time.Sleep(time.Second * 5)
	}

	//fmt.Printf("OperatorCRDString:%s\n", operatorCRDString)

	crds := make([]string, 0)
	if err := json.Unmarshal([]byte(operatorCRDString), &crds); err != nil {
		panic(err)
	}
	fmt.Printf("OperatorCRD:%v\n", crds)

	fmt.Println("Checking if OpenAPI Spec for the Operator is registered or not")
	openAPISpecRegistered := isOpenAPISpecRegistered(operatorChartURL)

	if !openAPISpecRegistered {
		fmt.Println("OpenAPI Spec for the Operator not registered.")
		operatorName, _ := parseChartNameVersion(operatorChartURL)
		operatorOpenAPIConfigMapName := uploadOperatorOpenAPISpec(operatorChartURL, c.kubeclientset)
		saveOperatorData(operatorName, crds, operatorOpenAPIConfigMapName)
		recordOpenAPISpecRegistration(operatorChartURL)

		status := "READY"
		c.updateFooStatus(foo, &crds, status)
	} else {
		fmt.Println("OpenAPI Spec for the Operator already registered.")
	}

	c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func isOpenAPISpecRegistered(operatorChartURL string) bool {
	operatorPresent := checkIfOperatorURLPresentInETCD(openAPISpecKey, operatorChartURL)
	return operatorPresent
}

func recordOpenAPISpecRegistration(operatorChartURL string) {
	storeChartURL(openAPISpecKey, operatorChartURL)
}

func saveOperatorData(operatorName string, crds []string, operatorOpenAPIConfigMapName string) {

	resourceKey := "/operators"
	var operatorDataMap = make(map[string]interface{})
	var operatorMap = make(map[string]map[string]interface{})
	var operatorMapList = make([]map[string]map[string]interface{}, 0)

	operatorDataMap["Name"] = operatorName
	operatorDataMap["CustomResources"] = crds
	operatorDataMap["ConfigMapName"] = operatorOpenAPIConfigMapName

	operatorMap["Operator"] = operatorDataMap

	operatorMapList = append(operatorMapList, operatorMap)

	storeEtcd(resourceKey, operatorMapList)

	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		panic(err)
	}
	crdclient, err := apiextensionsclientset.NewForConfig(cfg)

	for _, crdName := range crds {
		crdObj, err := crdclient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
		if err != nil {
			fmt.Errorf("Error:%s\n", err)
		}
		group := crdObj.Spec.Group
		version := crdObj.Spec.Version
		endpoint := "apis/" + group + "/" + version
		kind := crdObj.Spec.Names.Kind
		plural := crdObj.Spec.Names.Plural
		fmt.Printf("Group:%s, Version:%s, Kind:%s, Plural:%s, Endpoint:%s\n", group, version, kind, plural, endpoint)

		objectMeta := crdObj.ObjectMeta
		fmt.Printf("Object Meta:%v\n", objectMeta)
		//name := objectMeta.GetName()
		//namespace := objectMeta.GetNamespace()
		annotations := objectMeta.GetAnnotations()
		composition := annotations["composition"]

		var crdDetailsMap = make(map[string]interface{})
		crdDetailsMap["kind"] = kind
		crdDetailsMap["endpoint"] = endpoint
		crdDetailsMap["plural"] = plural
		crdDetailsMap["composition"] = composition

		//crdName := "postgreses.postgrescontroller.kubeplus"
		storeEtcd("/"+crdName, crdDetailsMap)

		storeEtcd("/"+kind+"-OpenAPISpecConfigMap", operatorOpenAPIConfigMapName)
	}
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

func uploadOperatorOpenAPISpec(chartURL string, kubeclientset kubernetes.Interface) string {
	extractOperatorChart(chartURL)
	chartConfigMapName := createConfigMap(chartURL, kubeclientset)
	return chartConfigMapName
}

func extractOperatorChart(chartURL string) {
	chartName, _ := parseChartNameVersion(chartURL)

	chartTarFile := chartName + ".tar"
	fmt.Printf("Chart tgz file name:%s\n", chartTarFile)
	out, err := os.Create(chartTarFile)
	defer out.Close()
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get(chartURL)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var buf bytes.Buffer
	buf1, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		log.Fatal(err1)
	}

	buf.Write(buf1)

	zr, err := gzip.NewReader(&buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Read tgz file in buffer")

	fmt.Printf("Name: %s\nComment: %s\nModTime: %s\n\n", zr.Name, zr.Comment, zr.ModTime.UTC())

	if _, err := io.Copy(out, zr); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Saving tgz buffer to file")

	if err := zr.Close(); err != nil {
		log.Fatal(err)
	}

	currentDir, err := os.Getwd()

	dirName := currentDir + "/" + chartName
	fmt.Printf("Chart tar file downloaded to:%s\n", dirName)
	_, err = os.Stat(dirName)
	if os.IsNotExist(err) {
		fmt.Printf("%s does not exist\n", dirName)
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Untaring the Chart")

	err = archiver.Tar.Open(chartTarFile, dirName)

	if err != nil {
		log.Fatal(err)
	}

	os.Chdir(dirName)
	os.Chdir(chartName)
	cwd, _ := os.Getwd()
	//fmt.Println("Listing %s", cwd)

	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
	}

	os.Chdir(currentDir)
}

func deleteConfigMap(chartURL string, kubeclientset kubernetes.Interface) {
	chartName, _ := parseChartNameVersion(chartURL)

	err1 := kubeclientset.CoreV1().ConfigMaps("default").Delete(chartName, &metav1.DeleteOptions{})
	if err1 != nil {
		fmt.Printf("Error:%s\n", err1.Error())
	}
}

func createConfigMap(chartURL string, kubeclientset kubernetes.Interface) string {
	chartName, _ := parseChartNameVersion(chartURL)

	currentDir, err := os.Getwd()
	//dirName := currentDir + "tmp/charts/" + chartName
	dirName := currentDir + "/" + chartName

	os.Chdir(dirName)
	os.Chdir(chartName)
	cwd, _ := os.Getwd()
	fmt.Printf("dirName:%s\n", cwd)

	//openapispecFile := dirName + "/openapispec.json"
	openapispecFile := "openapispec.json"

	//fmt.Printf("OpenAPI Spec file:%s\n", openapispecFile)

	jsonContents, err := ioutil.ReadFile(openapispecFile)
	jsonContents1 := string(jsonContents)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("OpenAPISpec Contents:%s\n", jsonContents1)

	os.Chdir(currentDir)

	configMapToCreate := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: chartName,
		},
		Data: map[string]string{
			"openapispec": jsonContents1,
		},
	}

	_, err1 := kubeclientset.CoreV1().ConfigMaps("default").Create(configMapToCreate)

	if err1 != nil {
		fmt.Printf("Error:%s\n", err1.Error())
	}

	return chartName
}

func parseChartNameVersion(chartURL string) (string, string) {

	//"https://s3-us-west-2.amazonaws.com/cloudark-helm-charts/postgres-crd-v2-chart-0.0.2.tgz"

	// 1. Split on '/'
	// 2. Split on 'tgz'
	// 3. Find last '/' -- Everything before is chartName everything after is version

	splitOnSlash := strings.Split(chartURL, "/")
	lastItem := splitOnSlash[len(splitOnSlash)-1]
	fmt.Printf("Last item:%s\n", lastItem)

	splitOnTgz := strings.Split(lastItem, ".tgz")
	candidate := splitOnTgz[0]
	fmt.Printf("Candidate:%s\n", candidate)

	nameVersionSplitIndex := strings.LastIndex(candidate, "-")
	versionStartIndex := nameVersionSplitIndex + 1
	version := candidate[versionStartIndex:]
	fmt.Printf("Version:%s\n", version)

	nameEndIndex := nameVersionSplitIndex
	name := candidate[0:nameEndIndex]
	fmt.Printf("Name:%s\n", name)

	return name, version
}

func deleteChartURL(resourceKey, chartURL string) {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}
	kapi := client.NewKeysAPI(c)

	operatorList := getList(resourceKey)
	var newList []string
	for _, operatorURL := range operatorList {
		if operatorURL != chartURL {
			newList = append(newList, operatorURL)
		}
	}

	jsonOperatorList, err2 := json.Marshal(&newList)
	if err2 != nil {
		panic(err2)
	}
	resourceValue := string(jsonOperatorList)

	//fmt.Printf("Setting %s->%s\n",resourceKey, resourceValue)
	_, err1 := kapi.Set(context.Background(), resourceKey, resourceValue, nil)
	if err1 != nil {
		log.Fatal(err)
	} else {
		//log.Printf("Set is done. Metadata is %q\n", resp.Node.Value)
	}
}

func storeChartURL(resourceKey, chartURL string) {
	addToList(resourceKey, chartURL)
}

func checkIfOperatorURLPresentInETCD(resourceKey, chartURL string) bool {
	operatorList := getList(resourceKey)

	operatorPresent := false
	for _, ops := range operatorList {
		if ops == chartURL {
			operatorPresent = true
		}
	}

	return operatorPresent
}

func getOperatorCRDs(chartURL string) string {
	operatorCRDString := getSingleValue(chartURL)
	return operatorCRDString
}

func addToList(resourceKey, chartURL string) {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}
	kapi := client.NewKeysAPI(c)

	var operatorList []string
	operatorList = getList(resourceKey)
	operatorPresent := checkIfOperatorURLPresentInETCD(resourceKey, chartURL)
	if !operatorPresent {
		operatorList = append(operatorList, chartURL)
	}

	jsonOperatorList, err2 := json.Marshal(&operatorList)
	if err2 != nil {
		panic(err2)
	}
	resourceValue := string(jsonOperatorList)

	//fmt.Printf("Setting %s->%s\n",resourceKey, resourceValue)
	_, err1 := kapi.Set(context.Background(), resourceKey, resourceValue, nil)
	if err1 != nil {
		log.Fatal(err)
	} else {
		//log.Printf("Set is done. Metadata is %q\n", resp.Node.Value)
	}
}

func getList(resourceKey string) []string {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}
	kapi := client.NewKeysAPI(c)

	var currentListString string
	var operatorList []string

	resp, err1 := kapi.Get(context.Background(), resourceKey, nil)
	if err1 != nil {
		fmt.Errorf("Error: %v", err1)
	} else {
		currentListString = resp.Node.Value

		if err = json.Unmarshal([]byte(currentListString), &operatorList); err != nil {
			panic(err)
		}
		//fmt.Printf("OperatorList:%v\n", operatorList)
	}
	return operatorList
}

func getSingleValue(resourceKey string) string {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	resp, err1 := kapi.Get(context.Background(), resourceKey, nil)
	if err1 != nil {
		fmt.Errorf("Error: %v", err1)
		return ""
	} else {
		//log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	}
	return resp.Node.Value
}

func (c *Controller) updateFooStatus(foo *operatorv1.Operator,
	crds *[]string, status string) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	fooCopy := foo.DeepCopy()

	fooCopy.Status.CustomResourceDefinitions = *crds
	fooCopy.Status.Status = status
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the Foo resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, err := c.sampleclientset.OperatorcontrollerV1().Operators(foo.Namespace).Update(fooCopy)
	if err != nil {
		fmt.Println("ERROR in UpdateFooStatus %v", err)
	}
	return err
}
