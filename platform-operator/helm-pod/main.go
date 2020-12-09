package main

import (
	restful "github.com/emicklei/go-restful"
	"time"
	"fmt"
	"bytes"
	"net/http"
	"k8s.io/client-go/rest"
	//"k8s.io/client-go/kubernetes"
	//"os/exec"
	"strings"
	"strconv"
	//"sync"

	"os"
	"k8s.io/client-go/dynamic"
	"k8s.io/apimachinery/pkg/runtime/schema"

	//"k8s.io/cli-runtime/pkg/genericclioptions"
	//restclient "k8s.io/client-go/rest"
	//"k8s.io/kubectl/pkg/util/templates"
	//"k8s.io/kubernetes/pkg/kubectl/cmd/exec"
	
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	platformworkflowclientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/client/clientset/versioned"
	//platformworkflowv1alpha1 "github.com/cloud-ark/kubeplus/platform-operator/pkg/apis/workflowcontroller/v1alpha1"
)

type fileSpec struct {
	PodNamespace string
	PodName      string
	File         string
}

type kindDetails struct {
	Kind string
	Version string
	Group string
	Plural string
}

var (
	kubeClient kubernetes.Interface
	dynamicClient dynamic.Interface
	cfg *rest.Config
	err error
	kindDetailsMap map[string]kindDetails
	CMD_RUNNER_POD string
	CMD_RUNNER_CONTAINER string
)

func init() {
	cfg, err = rest.InClusterConfig()
	//	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		panic(err.Error())
	}
	kubeClient = kubernetes.NewForConfigOrDie(cfg)
	dynamicClient, err = dynamic.NewForConfig(cfg)
	kindDetailsMap = make(map[string]kindDetails, 0)
	CMD_RUNNER_POD = "kubeplus"
	CMD_RUNNER_CONTAINER = "helmer"
}

func main() {
	fmt.Printf("Before registering...\n")
	register()
	fmt.Printf("After registering...\n")
	for {
		time.Sleep(60 * time.Second)
	}
}

func register() {
	ws := new(restful.WebService)
	ws.
		Path("/apis/kubeplus").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/deploy").To(deployChart).
		Doc("Deploy chart").
		Param(ws.PathParameter("customresource", "name of the customresource").DataType("string")))

	ws.Route(ws.GET("/metrics").To(getMetrics).
		Doc("Get Metrics"))

	restful.Add(ws)
	http.ListenAndServe(":8090", nil)
	fmt.Printf("Listening on port 8090...")
	fmt.Printf("Done installing helmer paths...")
}

func getMetrics(request *restful.Request, response *restful.Response) {
	fmt.Printf("Inside getMetrics...\n")
	customresource := request.QueryParameter("instance")
	kind := request.QueryParameter("kind")
	namespace := request.QueryParameter("namespace")
	fmt.Printf("Custom Resource:%s\n", customresource)
	fmt.Printf("Kind:%s\n", kind)
	fmt.Printf("Namespace:%s\n", namespace)

	config, _ := rest.InClusterConfig()
	var sampleclientset platformworkflowclientset.Interface
	sampleclientset = platformworkflowclientset.NewForConfigOrDie(config)

	resourceMonitors, err := sampleclientset.WorkflowsV1alpha1().ResourceMonitors(namespace).List(metav1.ListOptions{})
	for _, resMonitor := range resourceMonitors.Items {
		fmt.Printf("ResourceMonitor:%v\n", resMonitor)
		if err != nil {
			fmt.Errorf("Error:%s\n", err)
		}
		kind := resMonitor.Spec.Resource.Kind
		group := resMonitor.Spec.Resource.Group
		version := resMonitor.Spec.Resource.Version
		plural := resMonitor.Spec.Resource.Plural
		relationshipToMonitor := resMonitor.Spec.MonitorRelationships
		fmt.Printf("Kind:%s\n", kind)
		fmt.Printf("Group:%s\n", group)    		
		fmt.Printf("Version:%s\n", version)
		fmt.Printf("Plural:%s\n", plural)
		fmt.Printf("RelationshipToMonitor:%s\n", relationshipToMonitor)
	}

	helmrelease := getReleaseName(kind, customresource, namespace)
	fmt.Printf("Helm release3:%s\n", helmrelease)

 	cmdRunnerPod := CMD_RUNNER_POD

 	/*cpPluginsCmd := "cp /plugins/* bin/"
	fmt.Printf("cp plugins cmd:%s\n", cpPluginsCmd)
	executeExecCall(cmdRunnerPod, namespace, cpPluginsCmd)
	*/

	metricsCmd := "./root/kubectl metrics helmrelease " + helmrelease + " -o prometheus " 
	fmt.Printf("metrics cmd:%s\n", metricsCmd)
	ok, helmreleaseMetrics := executeExecCall(cmdRunnerPod, namespace, metricsCmd)
	if ok {
		fmt.Printf("%v\n", helmreleaseMetrics)
		// WIP - convert metrics from helmrelease to custom resource
		//prometheusMetrics := getPrometheusMetrics(kind, customresource, namespace, helmreleaseMetrics)
		response.Write([]byte(helmreleaseMetrics))
	} else {
		response.Write([]byte{})
	}
}

/*
func getPrometheusMetrics(kind, customresource, namespace, helmreleaseMetrics string) string {
	var metrics_helm_release map[string]interface{}
	err := json.Unmarshal([]byte(helmreleaseMetrics), &metrics_helm_release)
	if err != nil {
    	panic(err)
	}

	cpu = metrics_helm_release["cpu"]
	memory = metrics_helm_release["memory"]
	storage = metrics_helm_release["storage"]
	num_of_pods = metrics_helm_release["num_of_pods"]
	num_of_containers = metrics_helm_release["num_of_containers"]

	millis := time.Now()
    cpuMetrics = "cpu{kind="+kind+",customresoure="+customresource+",namespace="+namespace+"} " + str(cpu) + " "  + str(millis)
    memoryMetrics = "memory{kind="+kind+",customresoure="+customresource+",namespace="+namespace+"} " + str(memory) + " " + str(millis)
    storageMetrics = "storage{kind="+kind+",customresoure="+customresource+",namespace="+namespace+"} " + str(storage) + " " + str(millis)

	//numOfPods = 'pods{helmrelease="'+release_name+'"} ' + str(num_of_pods) + ' ' + str(millis)
	//numOfContainers = 'containers{helmrelease="'+release_name+'"} ' + str(num_of_containers) + ' ' + str(millis)
	metricsToReturn = cpuMetrics + "\n" + memoryMetrics + "\n" + storageMetrics + "\n" + numOfPods + "\n" + numOfContainers
	return metricsToReturn
}
*/

func getReleaseName(kind, customresource, namespace string) string {
	helmrelease := ""
	fmt.Printf("Kind:%s, Instance:%s, Namespace:%s\n", kind, customresource, namespace)
	derefstring := kind + ":" + customresource
	fmt.Printf("De-ref string:%s\n", derefstring)
	kinddetails, ok := kindDetailsMap[derefstring]
	if ok {
		fmt.Printf("Kinddetails:%v\n", kinddetails)
		group := kinddetails.Group
		version := kinddetails.Version
		plural := kinddetails.Plural

		res := schema.GroupVersionResource{Group: group,
										   Version: version,
										   Resource: plural}

		obj, _ := dynamicClient.Resource(res).Namespace(namespace).Get(customresource, metav1.GetOptions{})
		objData := obj.UnstructuredContent()
		fmt.Printf("objData:%v\n", objData)
		status := objData["status"]
		fmt.Printf("Status:%v\n", status)
		for key, element := range status.(map[string]interface{}) {
			fmt.Printf("Key:%s\n",key)
			key = strings.TrimSpace(key)
			if key == "helmrelease" {
				helmrelease = element.(string)
				fmt.Printf("Helm release1:%s\n", helmrelease)
				break
			}
		}
		fmt.Printf("Helm release2:%s\n", helmrelease)
	}
	return helmrelease
}

// ./kubectl exec helmer-677f87c67f-xvzz6 -- ./root/helm install moodle-operator-chart
// curl -v "10.0.9.208/kubeplus/deploy?platformworkflow=moodle1-workflow&customresource=mystack1&namespace=default"
// curl -v "10.0.2.202/kubeplus/deploy?platformworkflow=mysqlcluster&customresource=stack1&namespace=default"
// ./kubectl exec kubeplus -c helmer -- sh -c "export KUBEPLUS_HOME=/; export PATH=$KUBEPLUS_HOME/plugins/:/root:$PATH; kubectl metrics helmrelease falling-aardvark"
// curl -v "10.0.3.244:90/apis/platform-as-code/metrics?kind=MysqlClusterStack&customresource=stack1&namespace=default"
func deployChart(request *restful.Request, response *restful.Response) {
	fmt.Printf("Inside deployChart...\n")

    platformWorkflowName := request.QueryParameter("platformworkflow")
	customresource := request.QueryParameter("customresource")
	namespace := request.QueryParameter("namespace")
	overrides := request.QueryParameter("overrides")
	fmt.Printf("Custom Resource:%s\n", customresource)
	fmt.Printf("Resource Composition:%s\n", platformWorkflowName)
	fmt.Printf("Namespace:%s\n", namespace)
	fmt.Printf("Overrides:%s\n", overrides)

	if platformWorkflowName != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		var sampleclientset platformworkflowclientset.Interface
		sampleclientset = platformworkflowclientset.NewForConfigOrDie(config)

		platformWorkflow1, err := sampleclientset.WorkflowsV1alpha1().ResourceCompositions(namespace).Get(platformWorkflowName, metav1.GetOptions{})
		fmt.Printf("PlatformWorkflow:%v\n", platformWorkflow1)
		if err != nil {
			fmt.Errorf("Error:%s\n", err)
		}

	    customAPI := platformWorkflow1.Spec.NewResource
    	//for _, customAPI := range customAPIs {
    		kind := customAPI.Resource.Kind
    		group := customAPI.Resource.Group
    		version := customAPI.Resource.Version
    		plural := customAPI.Resource.Plural
    		chartURL := customAPI.ChartURL
    		chartName := customAPI.ChartName
 			fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s, ChartURL:%s\n ChartName:%s", kind, group, version, plural, chartURL, chartName)
 			
 			kinddetails := kindDetails{
 				Group: group,
 				Version: version,
 				Kind: kind,
 				Plural: plural,
 			}
 			kindDetailsMap[kind + ":" + customresource] = kinddetails

 			cmdRunnerPod := CMD_RUNNER_POD
 			if chartURL != "" {
	 			// 1. Extract Chart Name
	 			lastIndexOfSlash := strings.LastIndex(chartURL, "/")
	 			chartName1 := chartURL[lastIndexOfSlash+1:]
	 			fmt.Printf("ChartName1:%s\n", chartName1)
	 			parts := strings.Split(chartName1, "?")
	 			chartName2 := parts[0]
	 			fmt.Printf("ChartName2:%s\n", chartName2)

	 			lsCmd := "ls -l "

	 			// 2. Download the Chart
	 			wgetCmd := "wget " + chartURL
	 			fmt.Printf("wget cmd:%s\n", wgetCmd)
	 			executeExecCall(cmdRunnerPod, namespace, wgetCmd)
	 			executeExecCall(cmdRunnerPod, namespace, lsCmd)

	 			// 3. Rename the Chart to a friendlier name
	 			mvCmd := "mv /" + chartName1 + " /" + chartName2
	 			fmt.Printf("mv cmd:%s\n", mvCmd)
	 			executeExecCall(cmdRunnerPod, namespace, mvCmd)
	 			executeExecCall(cmdRunnerPod, namespace, lsCmd)

	 			// 4. Untar the Chart file
	 			untarCmd := "tar -xvzf " + chartName2
	  			fmt.Printf("untar cmd:%s\n", untarCmd)
	 			executeExecCall(cmdRunnerPod, namespace, untarCmd)
	 			executeExecCall(cmdRunnerPod, namespace, lsCmd)

	 			// 5. Create overrides.yaml
	 			//overrides := getOverrides(kind, group, version, plural, customresource, namespace)
	 			f, errf := os.Create("/chart/overrides.yaml")
	 			if errf != nil {
	 				fmt.Errorf("Error:%s\n", errf)
	 			}
	 			f.WriteString(overrides)

	 			// 6. Install the Chart
	 			fmt.Printf("ChartName to install:%s\n", chartName)
	 			helmInstallCmd := "./root/helm install -f " + "/chart/overrides.yaml " + " ./" + chartName 
	  			fmt.Printf("helm install cmd:%s\n", helmInstallCmd)
	 			ok, helmReleaseOP := executeExecCall(cmdRunnerPod, namespace, helmInstallCmd)
	 			if ok {
	 				lines := strings.Split(helmReleaseOP, "\n")
	 				for _, line := range lines {
	 					present := strings.Contains(line, "NAME:")
	 					if present {
	 						parts := strings.Split(line, ":")
	 						for _, part := range parts {
	 							if part != "" && part != " " && part != "NAME" {
			 						releaseName := part
			 						fmt.Printf("ReleaseName:%s\n", releaseName)
	 								releaseName = strings.TrimSpace(releaseName)
	 								fmt.Printf("RN:%s\n", releaseName)
	 								go updateStatus(kind, group, version, plural, customresource, namespace, releaseName)
	 							}
	 						}
	 					}
	 				}
	 			}
	    	}
 		//}
	}
}

func updateStatus(kind, group, version, plural, instance, namespace, releaseName string) {

	res := schema.GroupVersionResource{Group: group,
									   Version: version,
									   Resource: plural}
	for {
		obj, err := dynamicClient.Resource(res).Namespace(namespace).Get(instance, metav1.GetOptions{})
		if err == nil {
			objData := obj.UnstructuredContent()
			helmrelease := make(map[string]interface{},0)
			helmrelease["helmrelease"] = releaseName
			objData["status"] = helmrelease
			obj.SetUnstructuredContent(objData)
			dynamicClient.Resource(res).Namespace(namespace).Update(obj, metav1.UpdateOptions{})
			// break out of the for loop
			break 
		} else {
			time.Sleep(2 * time.Second)
		}
	}
	fmt.Printf("Done updating status of the CR instance:%s\n", instance)
}

// We cannot reach into the object to get the overrides as the object will not be
// created when deployChart will be called.
func getOverrides(kind, group, version, plural, instance, namespace string) string {
	var overrides string
	overrides = ""

	res := schema.GroupVersionResource{Group: group,
									   Version: version,
									   Resource: plural}

	obj, _ := dynamicClient.Resource(res).Namespace(namespace).Get(instance, metav1.GetOptions{})
	objData := obj.UnstructuredContent()
	fmt.Printf("objData:%v\n", objData)
	spec := objData["spec"]
	fmt.Printf("Spec:%v\n", spec)
	for key, element := range spec.(map[string]interface{}) {
		fmt.Printf("Key:%s\n",key)
		elem, ok := element.(int64)
		if ok {
			strelem := strconv.FormatInt(elem, 10)
			overrides = overrides + " " + key + ": " + strelem + "\n"
			fmt.Printf("overrides:%s\n", overrides)
		}
		elems, ok1 := element.(string)
		if ok1 {
			overrides = overrides + " " + key + ": " + elems + "\n"
		}
	}
	fmt.Printf("Overrides:%s\n", overrides)
	return overrides
}

func executeExecCall(runner, namespace, command string) (bool, string) {
	fmt.Println("Inside ExecuteExecCall")
	req := kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(runner).
		Namespace(namespace).
		SubResource("exec")

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		fmt.Printf("Error found trying to Exec command on pod: %s \n", err.Error())
		return false, ""
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Command:   strings.Fields(command),
		Container: CMD_RUNNER_CONTAINER,
		//Stdin:     stdin != nil,
		Stdout: true,
		Stderr: true,
		TTY:    false,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		fmt.Printf("Error found trying to Exec command on pod: %s \n", err.Error())
		return false, ""
	}

	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &execOut,
		Stderr: &execErr,
		Tty:    false,
	})
	if err != nil {
		fmt.Printf("StdOutput: %s\n", execOut.String())
		fmt.Printf("StdErr: %s\n", execErr.String())
		fmt.Printf("The command %s returned False : %s \n", command, err.Error())
		return false, ""
	}

	responseString := execOut.String()
	fmt.Printf("Output! :%v\n", responseString)
	return true, responseString
}