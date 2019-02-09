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
	"fmt"
	"strconv"
	"strings"
	"time"

	"database/sql"
	_ "github.com/lib/pq"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
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

	postgresv1 "github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/apis/postgrescontroller/v1"
	clientset "github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/client/clientset/versioned"
	postgresscheme "github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/client/clientset/versioned/scheme"
	informers "github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/client/informers/externalversions"
	listers "github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/client/listers/postgrescontroller/v1"
)

const controllerAgentName = "postgres-controller"

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
	PGPASSWORD  = "mysecretpassword"

	POSTGRES_PORT_BASE = 32000
	POSTGRES_PORT      int
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	foosLister        listers.PostgresLister
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
	fooInformer := sampleInformerFactory.Postgrescontroller().V1().Postgreses()

	// Create event broadcaster
	// Add postgres-controller types to the default Kubernetes Scheme so Events can be
	// logged for postgres-controller types.
	postgresscheme.AddToScheme(scheme.Scheme)
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
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Postgreses"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	fooInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFoo,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*postgresv1.Postgres)
			oldDepl := old.(*postgresv1.Postgres)
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
		/*
		DeleteFunc: func(obj interface{}) {
		        _, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
			   controller.deleteFoo(obj)
			}
		},*/
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

		foo, err := c.foosLister.Postgreses(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueFoo(foo)
		return
	}
}


func (c *Controller) deleteFoo(obj interface{}) {

	fmt.Println("Inside delete Foo")

	var err error
	if _, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
	   panic(err)
	}
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
	foo, err := c.foosLister.Postgreses(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	deploymentName := foo.Spec.DeploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	var verifyCmd string
	var actionHistory []string
	var serviceIP string
	var servicePort string
	var setupCommands []string
	var databases []string
	var users []postgresv1.UserSpec

	POSTGRES_PORT = POSTGRES_PORT_BASE
	POSTGRES_PORT_BASE = POSTGRES_PORT_BASE + 1


	// Get the deployment with the name specified in Foo.spec
	_, err = c.deploymentsLister.Deployments(foo.Namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		fmt.Printf("Received request to create CRD %s\n", deploymentName)
		serviceIP, servicePort, setupCommands, databases, users, verifyCmd = createDeployment(foo, c)
		for _, cmds := range setupCommands {
			// Don't save the connect command as we might connect later and perform more operations
			if !strings.Contains(cmds, "\\c") {
				actionHistory = append(actionHistory, cmds)
			}
		}
		fmt.Printf("Setup Commands: %v\n", setupCommands)
		fmt.Printf("Verify using: %v\n", verifyCmd)
		err = c.updateFooStatus(foo, &actionHistory, &users, &databases,
			verifyCmd, serviceIP, servicePort, "READY")
		if err != nil {
			return err
		}
	} else {
		fmt.Println("-------------------------------")
		fmt.Printf("CRD %s exists\n", deploymentName)
		fmt.Printf("Check using: kubectl describe postgres %s \n", deploymentName)

		pgresObj, err := c.sampleclientset.PostgrescontrollerV1().Postgreses(foo.Namespace).Get(deploymentName,
			metav1.GetOptions{})

		actionHistory := pgresObj.Status.ActionHistory
		serviceIP := pgresObj.Status.ServiceIP
		servicePort := pgresObj.Status.ServicePort
		verifyCmd := pgresObj.Status.VerifyCommand
		//fmt.Printf("Action History:[%s]\n", actionHistory)
		//fmt.Printf("Service IP:[%s]\n", serviceIP)
		//fmt.Printf("Service Port:[%s]\n", servicePort)
		fmt.Printf("Verify cmd: %v\n", verifyCmd)

		// 1. Find directly provided commands
		//setupCommands1 := canonicalize(foo.Spec.Commands)
		//setupCommands = getCommandsToRun(actionHistory, setupCommands1)
		//fmt.Printf("setupCommands: %v\n", setupCommands)

		var commandsToRun []string

		// 2. Reconcile databases
		desiredDatabases := foo.Spec.Databases
		currentDatabases := pgresObj.Status.Databases
		fmt.Printf("Current Databases:%v\n", currentDatabases)
		fmt.Printf("Desired Databases:%v\n", desiredDatabases)
		createDBCommands, dropDBCommands := getDatabaseCommands(desiredDatabases,
			currentDatabases)
		appendList(&commandsToRun, createDBCommands)
		appendList(&commandsToRun, dropDBCommands)

		// 3. Reconcile users
		desiredUsers := foo.Spec.Users
		currentUsers := pgresObj.Status.Users
		fmt.Printf("Current Users:%v\n", currentUsers)
		fmt.Printf("Desired Users:%v\n", desiredUsers)
		createUserCmds, dropUserCmds, alterUserCmds := getUserCommands(desiredUsers,
			currentUsers)
		appendList(&commandsToRun, createUserCmds)
		appendList(&commandsToRun, dropUserCmds)
		appendList(&commandsToRun, alterUserCmds)

		// 4. So what all commands should we run??
		fmt.Printf("commandsToRun:%v\n", commandsToRun)

		if len(commandsToRun) > 0 {
			updatePostgresInstance(pgresObj, c, commandsToRun)

			actionHistory = pgresObj.Status.ActionHistory
			//fmt.Printf("1111 Action History:%s\n", actionHistory)
			for _, cmds := range commandsToRun {
				actionHistory = append(actionHistory, cmds)
			}

			err = c.updateFooStatus(pgresObj, &actionHistory, &desiredUsers, &desiredDatabases,
				verifyCmd, serviceIP, servicePort, "READY")
			if err != nil {
				return err
			}
		}
	}
	c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateFooStatus(foo *postgresv1.Postgres,
	actionHistory *[]string, users *[]postgresv1.UserSpec, databases *[]string,
	verifyCmd string, serviceIP string, servicePort string,
	status string) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	fooCopy := foo.DeepCopy()
	fooCopy.Status.AvailableReplicas = 1

	fooCopy.Status.VerifyCommand = verifyCmd
	fooCopy.Status.ActionHistory = *actionHistory
	fooCopy.Status.Users = *users
	fooCopy.Status.Databases = *databases
	fooCopy.Status.ServiceIP = serviceIP
	fooCopy.Status.ServicePort = servicePort
	fooCopy.Status.Status = status
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the Foo resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, err := c.sampleclientset.PostgrescontrollerV1().Postgreses(foo.Namespace).Update(fooCopy)
	if err != nil {
		fmt.Println("ERROR in UpdateFooStatus %v", err)
	}

	return err
}

func updatePostgresInstance(foo *postgresv1.Postgres, c *Controller, setupCommands []string) {
	serviceIP := foo.Status.ServiceIP
	servicePort := foo.Status.ServicePort

	if len(setupCommands) > 0 {
		fmt.Println("Now setting up the database")
		var dummyList []string
		setupDatabase(serviceIP, servicePort, setupCommands, dummyList)
	}
}

func createDeployment(foo *postgresv1.Postgres, c *Controller) (string, string, []string, []string, []postgresv1.UserSpec, string) {

	deploymentsClient := c.kubeclientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deploymentName := foo.Spec.DeploymentName
	image := foo.Spec.Image
	users := foo.Spec.Users
	databases := foo.Spec.Databases
	setupCommands := canonicalize(foo.Spec.Commands)

	var userAndDBCommands []string
	var allCommands []string

	var currentDatabases []string
	var currentUsers []postgresv1.UserSpec
	createDBCmds, dropDBCmds := getDatabaseCommands(databases, currentDatabases)
	createUserCmds, dropUserCmds, alterUserCmds := getUserCommands(users, currentUsers)

	fmt.Printf("   Deployment:%v, Image:%v\n", deploymentName, image)
	fmt.Printf("   Users:%v\n", users)
	fmt.Printf("   Databases:%v\n", databases)
	fmt.Printf("   SetupCmds:%v\n", setupCommands)
	fmt.Printf("   CreateDBCmds:%v\n", createDBCmds)
	fmt.Printf("   DropDBCmds:%v\n", dropDBCmds)
	fmt.Printf("   CreateUserCmds:%v\n", createUserCmds)
	fmt.Printf("   DropUserCmds:%v\n", dropUserCmds)
	fmt.Printf("   AlterUserCmds:%v\n", alterUserCmds)

	appendList(&userAndDBCommands, createDBCmds)
	appendList(&userAndDBCommands, dropDBCmds)
	appendList(&userAndDBCommands, createUserCmds)
	appendList(&userAndDBCommands, dropUserCmds)
	appendList(&userAndDBCommands, alterUserCmds)
	fmt.Printf("   UserAndDBCmds:%v\n", userAndDBCommands)
	fmt.Printf("   SetupCmds:%v\n", setupCommands)

	appendList(&allCommands, userAndDBCommands)
	appendList(&allCommands, setupCommands)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "postgrescontroller.kubeplus/v1",
					Kind: "Postgres",
					Name: foo.Name,
					UID: foo.UID,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  deploymentName,
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 5432,
								},
							},
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									TCPSocket: &apiv1.TCPSocketAction{
										Port: apiutil.FromInt(5432),
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      60,
								PeriodSeconds:       2,
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "POSTGRES_PASSWORD",
									Value: PGPASSWORD,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	// Create Service
	fmt.Printf("Creating service...\n")
	serviceClient := c.kubeclientset.CoreV1().Services(apiv1.NamespaceDefault)
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "postgrescontroller.kubeplus/v1",
					Kind: "Postgres",
					Name: foo.Name,
					UID: foo.UID,
				},
			},
			Labels: map[string]string{
				"app": deploymentName,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "my-port",
					Port:       5432,
					TargetPort: apiutil.FromInt(5432),
					NodePort:   int32(POSTGRES_PORT),
					Protocol:   apiv1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": deploymentName,
			},
			Type: apiv1.ServiceTypeNodePort,
			//Type: apiv1.ServiceTypeClusterIP,
		},
	}

	result1, err1 := serviceClient.Create(service)
	if err1 != nil {
		panic(err1)
	}
	fmt.Printf("Created service %q.\n", result1.GetObjectMeta().GetName())
	//fmt.Printf("Service object:%v\n", result1)
	fmt.Println("---")
	fmt.Printf("ClusterIP:%s\n", result1.Spec.ClusterIP)
	serviceIP := result1.Spec.ClusterIP
	// Parse ServiceIP and Port
	fmt.Printf("Service IP:%s\n", serviceIP)

	nodePort1 := result1.Spec.Ports[0].NodePort
	nodePort := fmt.Sprint(nodePort1)
	fmt.Printf("Node Port:%s\n", nodePort)
	//servicePort := nodePort
	servicePort := "5432"
	fmt.Printf("Service Port:%s\n", servicePort)

	//time.Sleep(time.Second * 5)

	// Check if Postgres Pod is ready or not
	podReady := false
	for {
		pods := getPods(c, deploymentName)
		for _, d := range pods.Items {
			if strings.Contains(d.Name, deploymentName) {
				podConditions := d.Status.Conditions
				for _, podCond := range podConditions {
					if podCond.Type == corev1.PodReady {
						if podCond.Status == corev1.ConditionTrue {
							fmt.Println("Pod is running.")
							podReady = true
							break
						}
					}
				}
			}
			if podReady {
				break
			}
		}
		if podReady {
			break
		} else {
			fmt.Println("Waiting for Pod to get ready.")
			time.Sleep(time.Second * 4)
		}
	}

	// Wait couple of seconds more just to give the Pod some more time.
	time.Sleep(time.Second * 2)

	if len(userAndDBCommands) > 0 {
		var dummyList []string
		setupDatabase(serviceIP, servicePort, userAndDBCommands, dummyList)
	}

	if len(setupCommands) > 0 {
		setupDatabase(serviceIP, servicePort, setupCommands, databases)
	}

	verifyCmd := strings.Fields("psql -h <IP of the HOST> -p " + nodePort + " -U <user> " + " -d <db-name>")
	var verifyCmdString = strings.Join(verifyCmd, " ")
	fmt.Printf("VerifyCmd: %v\n", verifyCmd)
	return serviceIP, servicePort, allCommands, databases, users, verifyCmdString
}

func setupDatabase(serviceIP string, servicePort string, setupCommands []string, databases []string) {
	fmt.Println("Setting up database")
	fmt.Println("Commands:")
	fmt.Printf("%v", setupCommands)

	var host = serviceIP
	port := -1
	port, _ = strconv.Atoi(servicePort)
	var user = "postgres"
	var password = PGPASSWORD

	var psqlInfo string
	if len(databases) > 0 {
		dbname := databases[0]
		fmt.Println("%s\n", dbname)
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s sslmode=disable",
			host, port, user, password)
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	for _, command := range setupCommands {
		_, err = db.Exec(command)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Done setting up the database")
}

func getPods(c *Controller, deploymentName string) *apiv1.PodList {
	// TODO(devkulkarni): This is returning all Pods. We should change this
	// to only return Pods whose Label matches the Deployment Name.
	pods, err := c.kubeclientset.CoreV1().Pods("default").List(metav1.ListOptions{
		//LabelSelector: deploymentName,
		//LabelSelector: metav1.LabelSelector{
		//	MatchLabels: map[string]string{
		//	"app": deploymentName,
		//},
		//},
	})
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	if err != nil {
		fmt.Printf("%s", err)
	}
	return pods
}

func int32Ptr(i int32) *int32 { return &i }
