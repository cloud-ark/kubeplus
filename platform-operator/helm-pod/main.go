package main

import (
	restful "github.com/emicklei/go-restful"
	"time"
	"fmt"
	"bytes"
	"net/http"
	"net/url"
	"io/ioutil"
	"k8s.io/client-go/rest"
	//"k8s.io/client-go/kubernetes"
	//"os/exec"
	"strings"
	"strconv"
	//"sync"
	"context"
	"io/fs"
	"path/filepath"
	"encoding/json"

	"os"
	"k8s.io/client-go/dynamic"
	"k8s.io/apimachinery/pkg/runtime/schema"

	//"k8s.io/cli-runtime/pkg/genericclioptions"
	//restclient "k8s.io/client-go/rest"
	//"k8s.io/kubectl/pkg/util/templates"
	//"k8s.io/kubernetes/pkg/kubectl/cmd/exec"

	//"github.com/golang/glog"
	
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	platformworkflowclientset "github.com/cloud-ark/kubeplus/platform-operator/pkg/generated/clientset/versioned"
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
	KUBEPLUS_DEPLOYMENT string
	CMD_RUNNER_CONTAINER string
	KUBEPLUS_NAMESPACE string
	HELMER_PORT string
)

func init() {
	cfg, err = rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	kubeClient = kubernetes.NewForConfigOrDie(cfg)
	dynamicClient, err = dynamic.NewForConfig(cfg)
	kindDetailsMap = make(map[string]kindDetails, 0)
	KUBEPLUS_DEPLOYMENT = "kubeplus-deployment"
	CMD_RUNNER_CONTAINER = "helmer"
	KUBEPLUS_NAMESPACE = getNamespace()
	HELMER_PORT = "8090"
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

	ws.Route(ws.GET("/testChartDeployment").To(testChartDeployment).
		Doc("Test Chart Deployment"))

	ws.Route(ws.GET("/getPlural").To(getPlural).
		Doc("Get Plural"))

	ws.Route(ws.GET("/checkResource").To(checkResource).
		Doc("Check Resource"))

	ws.Route(ws.GET("/annotatecrd").To(annotateCRD).
		Doc("Annotate CRD"))

	ws.Route(ws.GET("/deletecrdinstances").To(deleteCRDInstances).
		Doc("Delete CRD Instances"))

	ws.Route(ws.GET("/updatecrdinstances").To(updateCRDInstances).
		Doc("Update CRD Instances"))

	ws.Route(ws.GET("/deletechartcrds").To(deleteChartCRDs).
		Doc("Delete Chart CRDs"))

	ws.Route(ws.GET("/getchartvalues").To(getChartValues).
		Doc("Get Chart Values"))

	restful.Add(ws)
	http.ListenAndServe(":" + HELMER_PORT, nil)
	fmt.Printf("Listening on port " + HELMER_PORT + " ...")
	fmt.Printf("Done installing helmer paths...")
}

func getNamespace() string {
	filePath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	content, err := ioutil.ReadFile(filePath)
    if err != nil {
     	fmt.Printf("Namespace file reading error:%v\n", err)
    }
    ns := string(content)
    ns = strings.TrimSpace(ns)
    fmt.Printf("Helmer NS:%s\n", ns)
    return ns
}

func getChartValues(request *restful.Request, response *restful.Response) {

    platformWorkflowName := request.QueryParameter("platformworkflow")
	namespace := request.QueryParameter("namespace")
	fmt.Printf("PlatformWorkflowName:%s\n", platformWorkflowName)
	fmt.Printf("Namespace:%s\n", namespace)

 	var valuesToReturn string

	if platformWorkflowName != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		var sampleclientset platformworkflowclientset.Interface
		sampleclientset = platformworkflowclientset.NewForConfigOrDie(config)

		platformWorkflow1, err := sampleclientset.WorkflowsV1alpha1().ResourceCompositions(namespace).Get(context.Background(), platformWorkflowName, metav1.GetOptions{})
		//fmt.Printf("PlatformWorkflow:%v\n", platformWorkflow1)
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
 			fmt.Printf("Kind:%s, Group:%s, Version:%s, Plural:%s, ChartURL:%s ChartName:%s\n", kind, group, version, plural, chartURL, chartName)

 			cmdRunnerPod := getKubePlusPod()
 			if cmdRunnerPod == "" {
 				fmt.Printf("Command runner Pod name could not be determined.. cannot continue.")
 				valuesToReturn = ""
 			}
 			parsedChartName := downloadUntarChartandGetName(chartURL, chartName, cmdRunnerPod, namespace)
 			chartValuesPath := "/" + parsedChartName + "/values.yaml"
 			fmt.Printf("Chart Values Path:%s\n",chartValuesPath)
 			readCmd := "cat " + chartValuesPath
 			fmt.Printf("cat cmd:%s\n", readCmd)
 			_, valuesToReturn = executeExecCall(cmdRunnerPod, readCmd)
 			//fmt.Printf("valuesToReturn:%v\n",valuesToReturn)
		}

	response.Write([]byte(valuesToReturn))
}

func getKubePlusPod() string {
	podName := ""
	/*
	deploymentObj, err1 := kubeClient.AppsV1().Deployments(KUBEPLUS_NAMESPACE).Get(KUBEPLUS_DEPLOYMENT, metav1.GetOptions{})
	if err1 != nil {
		fmt.Printf("Error:%v\n", err1)
		return podName
	}*/
	replicaSetList, err2 := kubeClient.AppsV1().ReplicaSets(KUBEPLUS_NAMESPACE).List(context.Background(), metav1.ListOptions{})
	if err2 != nil {
		fmt.Printf("Error:%v\n", err2)
		return podName
	}
	replicaSetName := ""
	for _, repSetObj := range replicaSetList.Items {
		ownerRefObj := repSetObj.ObjectMeta.OwnerReferences[0]
		depOwnerName := ownerRefObj.Name
		if depOwnerName == KUBEPLUS_DEPLOYMENT {
			replicaSetName = repSetObj.ObjectMeta.Name
			//fmt.Printf("DepOwnerName:%s, RSSetName:%s\n", depOwnerName, replicaSetName)
			break
		}
	}

	podList, err3 := kubeClient.CoreV1().Pods(KUBEPLUS_NAMESPACE).List(context.Background(), metav1.ListOptions{})
	if err3 != nil {
		fmt.Printf("Error:%v\n", err3)
		return podName
	}
	for _, podObj := range podList.Items {
		if podObj.ObjectMeta.OwnerReferences != nil {
			ownerRefObj := podObj.ObjectMeta.OwnerReferences[0]
			podOwnerName := ownerRefObj.Name
			if podOwnerName == replicaSetName {
				podName = podObj.ObjectMeta.Name
				//fmt.Printf("RSSetName:%s, PodName:%s\n", replicaSetName, podName)
				break
			}
		}
	}
	return podName
}

func deleteChartCRDs(request *restful.Request, response *restful.Response) {
	fmt.Printf("-- Inside deleteChartCRDs...\n")
	chartName := request.QueryParameter("chartName")
	fmt.Printf("Chart Name:%s\n", chartName)
	_, errF := os.Stat("/" + chartName)
	fmt.Printf("Path checking:%v\n", errF)
	if !os.IsNotExist(errF) {
		cmd := "kubectl delete -f /" + chartName + "/crds"
		cmdRunnerPod := getKubePlusPod()
		// Do in the background as deleting crds can be a time consuming action
		go executeExecCall(cmdRunnerPod, cmd)

		chartsDir := "/" + chartName + "/charts"
		_, errF = os.Stat(chartsDir)
		if !os.IsNotExist(errF) {
			err := filepath.Walk(chartsDir, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					fmt.Printf("Failure accessing path: %q: %v\n", path, err)
					return err
				}
				if info.IsDir() && info.Name() == "crds" {
					fmt.Printf("Found subchart crds: %q\n", path)
					cmd = "kubectl delete -f " + path
					go executeExecCall(cmdRunnerPod, cmd)
				}
				return nil
			})
			if err != nil {
				fmt.Printf("Error walking charts directory: %q: %v\n", chartsDir, err)
				return
			}
		}
	}
}

func deleteCRDInstances(request *restful.Request, response *restful.Response) {

	//sync.Mutex.Lock()
	fmt.Printf("-- Inside deleteCRDInstances...\n")
	kind := request.QueryParameter("kind")
	group := request.QueryParameter("group")
	version := request.QueryParameter("version")
	plural := request.QueryParameter("plural")
	namespace := request.QueryParameter("namespace")
	crName := request.QueryParameter("instance")
	fmt.Printf("Kind:%s\n", kind)
	fmt.Printf("Group:%s\n", group)
	fmt.Printf("Version:%s\n", version)
	fmt.Printf("Plural:%s\n", plural)
	fmt.Printf("Namespace:%s\n", namespace)
	fmt.Printf("CRName:%s\n", crName)

	apiVersion := group + "/" + version

	fmt.Printf("APIVersion:%s\n", apiVersion)

	ownerRes := schema.GroupVersionResource{Group: group,
									 		Version: version,
									   		Resource: plural}

	fmt.Printf("GVR:%v\n", ownerRes)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	dynamicClient, _ := dynamic.NewForConfig(config)

	crdObjList, err := dynamicClient.Resource(ownerRes).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error:%v\n...checking in non-namespace", err)
		crdObjList, err = dynamicClient.Resource(ownerRes).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error:%v\n", err)
		}
	}
	//fmt.Printf("CRDObjList:%v\n", crdObjList)

	execOutput := ""

	for _, instanceObj := range crdObjList.Items {
		objData := instanceObj.UnstructuredContent()
		//mapval, ok1 := lhsContent.(map[string]interface{})
		//if ok1 {
		objName := instanceObj.GetName()
		//fmt.Printf("Instance Name:%s\n", objName)

		if crName != "" && objName != crName {
			continue;
		}

		//fmt.Printf("objData:%v\n", objData)
		status := objData["status"]
		//fmt.Printf("Status:%v\n", status)
		labels := instanceObj.GetLabels()
                forcedDelete, _ := labels["delete"]
		if status != nil { //|| forcedDelete != "" {
			helmreleaseNS, helmrelease := getHelmReleaseName(status)
			fmt.Printf("Helm release NS and release name:%s, %s\n", helmreleaseNS, helmrelease)
			if helmreleaseNS != "" && helmrelease != "" {
				// Do in the background as 'helm delete' and deleting NS are time consuming actions
				go func() {
					deleteHelmRelease(helmreleaseNS, helmrelease)

					fmt.Printf("Helm release deleted...\n")
					fmt.Printf("Deleting the namespace...\n")

					// The namespace created to deploy the Helm chart
					// needs to be deleted only if the helmreleaseNS does not match namespace.
					// For managed app use-case, namespace and helmreleaseNS will be same.
					if helmreleaseNS != namespace {
						namespaceDeleteCmd := "./root/kubectl delete ns " + helmreleaseNS
						cmdRunnerPod := getKubePlusPod()
						_, op := executeExecCall(cmdRunnerPod, namespaceDeleteCmd)
						fmt.Printf("kubectl delete ns o/p:%v\n", op)
					}
				}()

				if crName == "" {
					fmt.Printf("Deleting the object %s\n", objName)
					dynamicClient.Resource(ownerRes).Namespace(namespace).Delete(context.Background(), objName, metav1.DeleteOptions{})
				}
			}
		} else if forcedDelete != "" {
                	fmt.Printf("Force delete the object %s\n", objName)
			response.Write([]byte(""))
		} else {
			response.Write([]byte("Error: Custom Resource instance cannot be deleted. It is not ready yet."))
			return
		}
	}

	if crName == "" { //crName == "" means that we received request to delete all objects
		lowercaseKind := strings.ToLower(kind)
		configMapName := lowercaseKind + "-usage"
		// Delete the usage configmap
		fmt.Printf("Deleting the usage configmap:%s\n", configMapName)
		kubeClient.CoreV1().ConfigMaps(namespace).Delete(context.Background(), configMapName, metav1.DeleteOptions{})
		fmt.Println("Done deleting CRD Instances..")
	}

	response.Write([]byte(execOutput))
	//sync.mutex.Unlock()
}

func deleteHelmRelease(helmreleaseNS, helmrelease string) bool {
	//fmt.Printf("Helm release:%s\n", helmrelease)
	cmd := "helm delete " + helmrelease + " -n " + helmreleaseNS
	fmt.Printf("Helm delete cmd:%s\n", cmd)
	var output string
	cmdRunnerPod := getKubePlusPod()
	ok, output := executeExecCall(cmdRunnerPod, cmd)
	fmt.Printf("Helm delete o/p:%v\n", output)
	return ok
}

func updateCRDInstances(request *restful.Request, response *restful.Response) {
	fmt.Printf("-- Inside updateCRDInstances...\n")

	kind := request.QueryParameter("kind")
	group := request.QueryParameter("group")
	version := request.QueryParameter("version")
	plural := request.QueryParameter("plural")
	namespace := request.QueryParameter("namespace")
	chartURL := request.QueryParameter("chartURL")
	chartName := request.QueryParameter("chartName")
	crName := request.QueryParameter("instance")

	fmt.Printf("Kind:%s\n", kind)
	fmt.Printf("Group:%s\n", group)
	fmt.Printf("Version:%s\n", version)
	fmt.Printf("Plural:%s\n", plural)
	fmt.Printf("Namespace:%s\n", namespace)
	fmt.Printf("ChartURL:%s\n", chartURL)
	fmt.Printf("ChartName:%s\n", chartName)
	fmt.Printf("CRName:%s\n", crName)

	if _, errF := os.Stat("/" + chartName); os.IsNotExist(errF) {
		fmt.Printf("Error: chart '%s' not installed", chartName)
		errFString := string(errF.Error())
		response.Write([]byte(errFString))
		return
	}
	
	cmdRunnerPod := getKubePlusPod()
	downloadUntarChartandGetName(chartURL, chartName, cmdRunnerPod, namespace)

	apiVersion := group + "/" + version
	fmt.Printf("APIVersion:%s\n", apiVersion)

	ownerRes := schema.GroupVersionResource{Group: group, Version: version, Resource: plural}
	fmt.Printf("GVR:%v\n", ownerRes)
	
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	dynamicClient, _ := dynamic.NewForConfig(config)

	crdObjList, err := dynamicClient.Resource(ownerRes).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error:%v\n...checking in non-namespace", err)
		crdObjList, err = dynamicClient.Resource(ownerRes).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error:%v\n", err)
			errString := string(err.Error())
			response.Write([]byte(errString))
			return
		}
	}

	execOutput := ""

	for _, instanceObj := range crdObjList.Items {
		objData := instanceObj.UnstructuredContent()
		objName := instanceObj.GetName()

		if crName != "" && objName != crName {
			continue;
		}

		fmt.Printf("objData:%v\n", objData)
		status := objData["status"]
		fmt.Printf("Status:%v\n", status)

		if status != nil {
			helmreleaseNS, helmrelease := getHelmReleaseName(status)
			fmt.Printf("Helm release:%s, %s\n", helmreleaseNS, helmrelease)
			if helmreleaseNS != "" && helmrelease != "" {
				_, output := upgradeHelmRelease(helmreleaseNS, helmrelease, chartName)
				output = strings.ReplaceAll(output, "\n", "")
				fmt.Printf("Helm release updated...%s\n", output)
				helmReleaseFQDN := status.(map[string]interface{})["helmrelease"]
				helmReleaseFQDN_str := helmReleaseFQDN.(string)
				//helmrelease["helmrelease"] = targetNS + ":" + releaseName
				//status1 := helmReleaseFQDN_str + "\n" + output + "\n" + "Update complete."
				helmreleaseUpdate := make(map[string]string)
				helmreleaseUpdate["helmrelease"] = helmReleaseFQDN_str
				helmreleaseUpdate["error"] = chartURL + " upgrade output:" + output
	                        objData["status"] = helmreleaseUpdate
        	                fmt.Printf("objData:%v\n",objData)
                		instanceObj.SetUnstructuredContent(objData)
                        	dynamicClient.Resource(ownerRes).Namespace(namespace).Update(context.Background(), &instanceObj, metav1.UpdateOptions{})
			}
		}
	}

	response.Write([]byte(execOutput))
}

func upgradeHelmRelease(helmreleaseNS, helmrelease, chartName string) (bool, string) {
	fmt.Printf("Helm release:%s\n", helmrelease)
	cmd := "helm upgrade " + helmrelease + " /" + chartName + " -n " + helmreleaseNS
	fmt.Printf("Helm upgrade cmd:%s\n", cmd)
	var output string
	cmdRunnerPod := getKubePlusPod()
	ok, output := executeExecCall(cmdRunnerPod, cmd)
	fmt.Printf("Helm upgrade o/p:%v\n", output)
	return ok, output
}

func annotateCRD(request *restful.Request, response *restful.Response) {
	fmt.Printf("Inside annotateCRD...\n")
	kind := request.QueryParameter("kind")
	group := request.QueryParameter("group")
	plural := request.QueryParameter("plural")
	chartkinds := request.QueryParameter("chartkinds")	
	//fmt.Printf("Kind:%s\n", kind)
	//fmt.Printf("Group:%s\n", group)
	//fmt.Printf("Plural:%s\n", plural)
	//fmt.Printf("Chart Kinds:%s\n", chartkinds)
	chartkinds = strings.Replace(chartkinds, "-", ";", 1)

 	cmdRunnerPod := getKubePlusPod()

	//kubectl annotate crd mysqlservices.platformapi.kubeplus resource/annotation-relationship="on:MysqlCluster; Secret, key:meta.helm.sh/release-name, value:INSTANCE.metadata.name"

 	fqcrd := plural + "." + group
 	fmt.Printf("FQCRD:%s\n", fqcrd)
 	lowercaseKind := strings.ToLower(kind)
 	annotationValue := lowercaseKind + "-INSTANCE.metadata.name"
 	fmt.Printf("Annotation value:%s\n", annotationValue)
 	annotationString := " resource/annotation-relationship=\"" + "on:" + chartkinds + ", key:meta.helm.sh/release-name, value:" + annotationValue + "\""
 	fmt.Printf("Annotation String:%s\n", annotationString)
	cmd := "./root/kubectl annotate crd " + fqcrd + annotationString
	fmt.Printf("API resources cmd:%s\n", cmd)
	var ok bool
	var output string 
	for {
		ok, output = executeExecCall(cmdRunnerPod, cmd)
		fmt.Printf("CRD annotate o/p:%v\n", output)
		if !ok {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
}

func checkResource(request *restful.Request, response *restful.Response) {
	fmt.Printf("Inside checkResource...\n")

	kind := request.QueryParameter("kind")
	plural := request.QueryParameter("plural")

 	cmdRunnerPod := getKubePlusPod()

	cmd := "./root/kubectl api-resources " //| grep " + kind + " | grep " + group + " | awk '{print $1}' " 
	fmt.Printf("API resources cmd:%s\n", cmd)
	_, output := executeExecCall(cmdRunnerPod, cmd)

	failed := ""
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.Split(line, " ")
		nonEmptySlice := make([]string,0)
		for _, p := range parts {
			if p != "" && p != " " {
				nonEmptySlice = append(nonEmptySlice, p)
			}
		}
		if len(nonEmptySlice) > 0 {
			existingKind := nonEmptySlice[len(nonEmptySlice)-1]
			existingKind = strings.TrimSuffix(existingKind, "\n")
			//fmt.Printf("ExistingKind:%s\n", existingKind)
			if kind == existingKind {
				failed = kind
				break
			}
			existingPlural := nonEmptySlice[0]
			existingPlural = strings.TrimSuffix(existingPlural, "\n")
			//fmt.Printf("ExistingKind:%s\n", existingKind)
			if plural == existingPlural {
				failed = plural
			}
		}
	}

	//fmt.Printf("Plural to return:%s\n", string(pluralToReturn))
	response.Write([]byte(failed))
}


func getPlural(request *restful.Request, response *restful.Response) {
	fmt.Printf("Inside getPlural...\n")

	kind := request.QueryParameter("kind")
	//group := request.QueryParameter("group")
	//fmt.Printf("Kind:%s\n", kind)
	//fmt.Printf("Group:%s\n", group)

 	cmdRunnerPod := getKubePlusPod()

	cmd := "./root/kubectl api-resources " //| grep " + kind + " | grep " + group + " | awk '{print $1}' " 
	fmt.Printf("API resources cmd:%s\n", cmd)
	_, output := executeExecCall(cmdRunnerPod, cmd)

	pluralToReturn := ""
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.Split(line, " ")
		nonEmptySlice := make([]string,0)
		for _, p := range parts {
			if p != "" && p != " " {
				nonEmptySlice = append(nonEmptySlice, p)
			}
		}
		if len(nonEmptySlice) > 0 {
			existingKind := nonEmptySlice[len(nonEmptySlice)-1]
			existingKind = strings.TrimSuffix(existingKind, "\n")
			//fmt.Printf("ExistingKind:%s\n", existingKind)
			if kind == existingKind {
				pluralToReturn = nonEmptySlice[0]
				break
			}
		}
	}

	//fmt.Printf("Plural to return:%s\n", string(pluralToReturn))

	response.Write([]byte(pluralToReturn))
}

func getMetrics(request *restful.Request, response *restful.Response) {
	fmt.Printf("Inside getMetrics...\n")

	metricsToReturn := ""
 	cmdRunnerPod := getKubePlusPod()

	customresource := request.QueryParameter("instance")
	kind := request.QueryParameter("kind")
	namespace := request.QueryParameter("namespace")
	fmt.Printf("Custom Resource:%s\n", customresource)
	fmt.Printf("Kind:%s\n", kind)
	fmt.Printf("Namespace:%s\n", namespace)

	/*helmrelease := getReleaseName(kind, customresource, namespace)
	fmt.Printf("Helm release3:%s\n", helmrelease)
	if helmrelease != "" {
		metricsCmd := "./root/kubectl metrics helmrelease " + helmrelease + " -o prometheus " 
		fmt.Printf("metrics cmd:%s\n", metricsCmd)
		_, metricsToReturn = executeExecCall(cmdRunnerPod, metricsCmd)
	} else {*/
		config, _ := rest.InClusterConfig()
		var sampleclientset platformworkflowclientset.Interface
		sampleclientset = platformworkflowclientset.NewForConfigOrDie(config)

		resourceMonitors, err := sampleclientset.WorkflowsV1alpha1().ResourceMonitors(KUBEPLUS_NAMESPACE).List(context.Background(), metav1.ListOptions{})
		//followConnections := ""
		for _, resMonitor := range resourceMonitors.Items {
			fmt.Printf("ResourceMonitor:%v\n", resMonitor)
			if err != nil {
				fmt.Errorf("Error:%s\n", err)
			}
			reskind := resMonitor.Spec.Resource.Kind
			resgroup := resMonitor.Spec.Resource.Group
			resversion := resMonitor.Spec.Resource.Version
			resplural := resMonitor.Spec.Resource.Plural
			relationshipToMonitor := resMonitor.Spec.MonitorRelationships
			fmt.Printf("Kind:%s\n", reskind)
			fmt.Printf("Group:%s\n", resgroup)    		
			fmt.Printf("Version:%s\n", resversion)
			fmt.Printf("Plural:%s\n", resplural)
			fmt.Printf("RelationshipToMonitor:%s\n", relationshipToMonitor)

			if relationshipToMonitor != "all" && relationshipToMonitor != "owner" {
				metricsToReturn = "Value " + relationshipToMonitor + " not supported. Valid values: all, owner"
				break
			}
			if reskind == kind {
				if relationshipToMonitor == "all" {
					//followConnections = " --follow-connections"
				}
			}
		}
		metricsCmd := "./root/kubectl metrics " + kind + " " + customresource + " " + namespace + " -o prometheus " //+ followConnections
		fmt.Printf("metrics cmd:%s\n", metricsCmd)
		_, metricsToReturn = executeExecCall(cmdRunnerPod, metricsCmd)
	//}
 	/*cpPluginsCmd := "cp /plugins/* bin/"
	fmt.Printf("cp plugins cmd:%s\n", cpPluginsCmd)
	executeExecCall(cmdRunnerPod, cpPluginsCmd)
	*/
	/*if ok {
		fmt.Printf("%v\n", helmreleaseMetrics)
		// WIP - convert metrics from helmrelease to custom resource
		//prometheusMetrics := getPrometheusMetrics(kind, customresource, namespace, helmreleaseMetrics)
		response.Write([]byte(helmreleaseMetrics))
	} else {
		response.Write([]byte{})
	}*/
	response.Write([]byte(metricsToReturn))
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

func getReleaseName(kind, customresource, namespace string) (string, string) {
	helmrelease := ""
	helmreleaseNS := ""
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

		obj, _ := dynamicClient.Resource(res).Namespace(namespace).Get(context.Background(), customresource, metav1.GetOptions{})
		objData := obj.UnstructuredContent()
		fmt.Printf("objData:%v\n", objData)
		status := objData["status"]
		//fmt.Printf("Status:%v\n", status)
		helmreleaseNS, helmrelease = getHelmReleaseName(status)
		fmt.Printf("Helm release2:%s\n", helmrelease)
	}
	return helmreleaseNS, helmrelease
}

func getHelmReleaseName(object interface{}) (string, string) {
	helmrelease := ""
	helmreleaseNS := ""
	helmreleaseName := ""
	status := object.(map[string]interface{})
	for key, element := range status {
		//fmt.Printf("Key:%s\n",key)
		key = strings.TrimSpace(key)
		if key == "helmrelease" {
			helmrelease = element.(string)
			//fmt.Printf("Helm release1:%s\n", helmrelease)
			lines := strings.Split(helmrelease, "\n")
			releaseLine := strings.TrimSpace(lines[0])
			parts := strings.Split(releaseLine,":")
			helmreleaseNS = strings.TrimSpace(parts[0])
			helmreleaseName = strings.TrimSpace(parts[1])
			break
		}
	}
	return helmreleaseNS, helmreleaseName
}

func downloadUntarChartandGetName(chartURL, chartName, cmdRunnerPod, namespace string) string {
	fmt.Printf("Inside downloadUntarChartandGetName\n")

	parsedChartName := ""
 	if !strings.Contains(chartURL, "file:///") {
 		// Download and untar the chart
 		parsedChartName = downloadChart(chartURL, cmdRunnerPod, namespace)
	} else {
		// Untar the chart
		parts := strings.Split(chartURL, "file:///")
		charttgz := strings.TrimSpace(parts[1])
		fmt.Printf("Chart tgz:%s\n",charttgz)
		removePreviousChart(chartName, cmdRunnerPod)
		parsedChartName = untarChart(charttgz, cmdRunnerPod, namespace)
	}
	return parsedChartName
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
	//encodedoverrides := request.QueryParameter("overrides")
	//overrides, err := url.QueryUnescape(encodedoverrides)
	overrideBytes, _ := os.ReadFile("/crdinstances/" + platformWorkflowName + "-" + customresource)
	overrides := string(overrideBytes)

	cpu_req := request.QueryParameter("cpu_req")
	cpu_lim := request.QueryParameter("cpu_lim")
	mem_req := request.QueryParameter("mem_req")
	mem_lim := request.QueryParameter("mem_lim")
	labels := request.QueryParameter("labels")
	if err != nil {
		fmt.Printf("Error encountered in decoding overrides:%v\n", err)
		fmt.Printf("Not continuing...")
		response.Write([]byte(""))
	}
	dryrun := request.QueryParameter("dryrun")
	fmt.Printf("PlatformWorkflowName:%s\n", platformWorkflowName)
	fmt.Printf("Custom Resource:%s\n", customresource)
	fmt.Printf("Resource Composition:%s\n", platformWorkflowName)
	fmt.Printf("Namespace:%s\n", namespace)
	//fmt.Printf("Overrides:%s\n", overrides)
	fmt.Printf("Dryrun:%s\n", dryrun)
	fmt.Printf("Labels:%s\n", labels)

	if labels != "" {

		labelsMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(labels), &labelsMap); err != nil {
			panic(err)
		}
		//fmt.Println("LabelsMap:%v\n",labelsMap)
		isCreatedByKubePlus, ok := labelsMap["created-by"]
		if ok && isCreatedByKubePlus == "kubeplus" {
			// This is the call triggered by creation of CRD instance.
			// Nothing to do.
			response.Write([]byte(string("")))
			return
		}
	}

	kinds := make([]string, 0)
	//ok := false
	execOutput := ""

	crObjNamespace := namespace

	if platformWorkflowName != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		var sampleclientset platformworkflowclientset.Interface
		sampleclientset = platformworkflowclientset.NewForConfigOrDie(config)

		resourceCompositionNS := KUBEPLUS_NAMESPACE
		platformWorkflow1, err := sampleclientset.WorkflowsV1alpha1().ResourceCompositions(resourceCompositionNS).Get(context.Background(), platformWorkflowName, metav1.GetOptions{})
		//fmt.Printf("PlatformWorkflow:%v\n", platformWorkflow1)
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

 			cmdRunnerPod := getKubePlusPod()

	 		lowercaseKind := strings.ToLower(kind)
	 		releaseName := lowercaseKind + "-" + customresource
	 		fmt.Printf("Release name:%s\n", releaseName)
 			if chartURL != "" {

 			parsedChartName := downloadUntarChartandGetName(chartURL, chartName, cmdRunnerPod, namespace)

 				// 1. Download the chart
 				//parsedChartName := downloadChart(chartURL, cmdRunnerPod, namespace)

	 			// 5. Create overrides.yaml
	 			//overrides := getOverrides(kind, group, version, plural, customresource, namespace)
	 			chartDir := "/chart/" + parsedChartName
	 			fmt.Printf("Chart dir:%s\n", chartDir)
	 			overRidesFile := chartDir + "/overrides.yaml"
	 			fmt.Printf("Overrides file:%s\n", overRidesFile)
				errm := os.Mkdir(chartDir, 0755)
				if errm != nil {
					fmt.Errorf("Error:%s\n", errm)
				}
	 			f, errf := os.Create(overRidesFile)
	 			if errf != nil {
	 				fmt.Errorf("Error:%s\n", errf)
	 			}
	 			f.WriteString(overrides)

	 			// 6. Install the Chart
	 			//fmt.Printf("ChartName to install:%s\n", chartName)

				targetNSName := namespace
				if dryrun == "" {
					createNSCmd := ""
					getNSCmd := ""
					doHelmInstall := true
					doHelmUpgrade := false
					getNSSuccess := false 
					if namespace == KUBEPLUS_NAMESPACE {
						getNSCmd = "./root/kubectl get ns " + customresource
						getNSSuccess, execOutput = executeExecCall(cmdRunnerPod, getNSCmd)
						fmt.Printf("Output of kubectl get ns:%v\n", execOutput)
						if !getNSSuccess {
							createNSCmd = "./root/kubectl create ns " + customresource
							_, execOutput = executeExecCall(cmdRunnerPod, createNSCmd)
							fmt.Printf("Output of Create NS Cmd:%v\n", execOutput)
						} else {
							fmt.Printf("NS " + customresource + " exists. Performing Helm upgrade.")
							doHelmUpgrade = true
							doHelmInstall = false
						}
						targetNSName = customresource
					}

					if doHelmInstall {
					annotateNSCmd := "./root/kubectl annotate --overwrite=true namespace " + targetNSName + " meta.helm.sh/release-name=\"" + releaseName + "\""
					fmt.Printf("Annotation NS Cmd:%v\n", annotateNSCmd)
					_, execOutput = executeExecCall(cmdRunnerPod, annotateNSCmd)
					fmt.Printf("Output of Annotate NS Cmd:%v\n", execOutput)

					labelNSCmd := "./root/kubectl label --overwrite=true namespace " + targetNSName + " managedby=kubeplus" 
					fmt.Printf("Label NS Cmd:%v\n", labelNSCmd)
					_, execOutput = executeExecCall(cmdRunnerPod, labelNSCmd)
					fmt.Printf("Output of Label NS Cmd:%v\n", execOutput)

					// Install the Helm chart in the namespace that is created for that instance
					helmInstallCmd := "helm install " + releaseName + " ./" + parsedChartName  + " -f " + overRidesFile + " -n " + targetNSName
	  				fmt.Printf("ABC helm install cmd:%s\n", helmInstallCmd)
					go runHelmInstall(cmdRunnerPod, helmInstallCmd, releaseName, kind, group, version, plural, customresource, crObjNamespace, targetNSName, cpu_req, cpu_lim, mem_req, mem_lim)
					}

					if doHelmUpgrade {
						helmUpgradeCmd := "helm upgrade " + releaseName + " ./" + parsedChartName  + " -f " + overRidesFile + " -n " + targetNSName
	  					fmt.Printf("ABC helm upgrade cmd:%s\n", helmUpgradeCmd)

						_, helmUpgradeOutput := executeExecCall(cmdRunnerPod, helmUpgradeCmd)
						fmt.Printf("Helm upgrade o/p:%v\n", helmUpgradeOutput)
					}
				}

				if dryrun == "true" {
					fmt.Printf("DRY RUN - ABC:%s\n", dryrun)
					helmInstallCmd := "helm install " + releaseName + " ./" + parsedChartName  + " -f " + overRidesFile + " -n " + namespace + " --dry-run"

					_, execOutput := executeExecCall(cmdRunnerPod, helmInstallCmd)

					if strings.Contains(execOutput, "Error") {
						fmt.Printf("ExecOutput:%s\n", execOutput)
						response.Write([]byte(execOutput))
					} else {
	 					lines := strings.Split(execOutput, "\n")
	 					for _, line := range lines {
	 						present := strings.Contains(line, "kind:")
	 						if present {
	 							parts := strings.Split(line, ":")
	 							if len(parts) >= 2 {
	 								kindName := parts[1]
	 								kindName = strings.TrimSpace(kindName)
	 								kindName = strings.TrimSuffix(kindName, "\n")
									kinds = append(kinds, kindName)
	 							}
	 						}
	 					}
						//fmt.Printf("Kinds:%v\n", kinds)
						// Appending the Namespace to the list of Kinds since we are creating NS corresponding to
						// each Helm release.
						kinds = append(kinds, "Namespace")
						kindsString := strings.Join(kinds, "-")
						fmt.Printf("KindString:%s\n", kindsString)
						response.Write([]byte(kindsString))
					}
				}
	    	}
	}
	// Everything went fine
	response.Write([]byte(string("")))
}

func runHelmInstall(cmdRunnerPod, helmInstallCmd, releaseNameInCmd, kind, group, version, plural, customresource, crObjNamespace, targetNSName, cpu_req, cpu_lim, mem_req, mem_lim string) {

	ok, execOutput := executeExecCall(cmdRunnerPod, helmInstallCmd)
	 			if ok {
	 				helmReleaseOP := execOutput
	 				lines := strings.Split(helmReleaseOP, "\n")
					var releaseFound bool = false
					releaseName := "" //should be same as releaseNameInCmd
					notes := ""
					var notesStart bool = false
	 				for _, line := range lines {
							if notesStart {
								notes = notes + line + "\n"
							}
							if strings.Contains(line, "NOTES") {
								notesStart = true
							}
		 					present := strings.Contains(line, "NAME:")
		 					if present && !releaseFound {
	 							parts := strings.Split(line, ":")
	 							for _, part := range parts {
		 							if part != "" && part != " " && part != "NAME" {
				 						releaseName = part
				 						fmt.Printf("ReleaseName:%s\n", releaseName)
		 								releaseName = strings.TrimSpace(releaseName)
		 								fmt.Printf("RN:%s\n", releaseName)
										releaseFound = true
									}
	 							}
	 						}
	 				}
					if releaseFound {
						//statusToUpdate := releaseName + "\n" + notes
						go updateStatus(kind, group, version, plural, customresource, crObjNamespace, targetNSName, releaseName, notes)
						if (cpu_req != "" && cpu_lim != "" && mem_req != "" && mem_lim != "") {
							go createResourceQuota(targetNSName, releaseName, cpu_req, cpu_lim, mem_req, mem_lim)
						}
						go createNetworkPolicy(targetNSName, releaseName)
		 			}
	 			} else {
					//statusToUpdate := releaseNameInCmd + "\n" + execOutput
		 			go updateStatus(kind, group, version, plural, customresource, crObjNamespace, targetNSName, releaseNameInCmd, execOutput)
					errOp := string(execOutput)
					instanceExists := strings.Contains(errOp, "cannot re-use a name that is still in use")
					if !instanceExists {
						deleteNSCmd := "./root/kubectl delete ns " + customresource
						_, execOutput1 := executeExecCall(cmdRunnerPod, deleteNSCmd)
						fmt.Printf("Output of delete NS Cmd:%v\n", execOutput1)
					}
					// there was some error
					/*response.Write([]byte(string(execOutput)))*/
	 			}
}

func createResourceQuota(targetNS, helmRelease, cpu_req, cpu_lim, mem_req, mem_lim string) {
	fmt.Printf("Inside createResourceQuota...\n")

	args := fmt.Sprintf("namespace=%s&helmrelease=%s&cpu_req=%s&cpu_lim=%s&mem_req=%s&mem_lim=%s",targetNS,helmRelease,cpu_req,cpu_lim,mem_req,mem_lim)
        var url1 string
        url1 = fmt.Sprintf("http://localhost:5005/resource_quota?%s", args)
        fmt.Printf("Url:%s\n", url1)

	http.Get(url1)

	fmt.Printf("After invoking /resource_quota")
}

func createNetworkPolicy(targetNS, helmRelease string) {
	fmt.Printf("Inside createNetworkPolicy...\n")

	args := fmt.Sprintf("namespace=%s&helmrelease=%s",targetNS,helmRelease)
        var url1 string
        url1 = fmt.Sprintf("http://localhost:5005/network_policy?%s", args)
        fmt.Printf("Url:%s\n", url1)

	http.Get(url1)

	fmt.Printf("After invoking /network_policy")
}


func testChartDeployment(request *restful.Request, response *restful.Response) {
	fmt.Printf("Inside testChartDeployment...\n")

	namespace := request.QueryParameter("namespace")
	kind := request.QueryParameter("kind")
	chartName := request.QueryParameter("chartName")
	encodedChartURL := request.QueryParameter("chartURL")
	chartURL, _ := url.QueryUnescape(encodedChartURL)

 	cmdRunnerPod := getKubePlusPod()

  	//parsedChartName := downloadChart(chartURL, cmdRunnerPod, namespace)
 	parsedChartName := downloadUntarChartandGetName(chartURL, chartName, cmdRunnerPod, namespace)
 	releaseName := strings.ToLower(kind) + "-" + parsedChartName

	//helmInstallCmd := "./root/helm install " + releaseName + " ./" + parsedChartName  + " -n " + namespace + " --dry-run" 
	helmInstallCmd := "helm install " + releaseName + " ./" + parsedChartName  + " -n " + namespace
	fmt.Printf("helm install cmd:%s\n", helmInstallCmd)
	_, execOutput := executeExecCall(cmdRunnerPod, helmInstallCmd)
	fmt.Printf("Test chart deployment - DEF:%s\n", execOutput)

	helmDeleteCmd := "helm delete " + releaseName + " -n " + namespace
	fmt.Printf("helm delete cmd:%s\n", helmDeleteCmd)
	executeExecCall(cmdRunnerPod, helmDeleteCmd)

	response.Write([]byte(execOutput))
}

func removePreviousChart(chartName, cmdRunnerPod string) {
	// 1. Remove previous instance of chart
	rmCmd := "rm -rf /" + chartName
	fmt.Printf("rm cmd:%s\n", rmCmd)
	executeExecCall(cmdRunnerPod, rmCmd)
}

func downloadChart(chartURL, cmdRunnerPod, namespace string) string {
	 			// 1. Extract Chart Name
	 			lastIndexOfSlash := strings.LastIndex(chartURL, "/")
	 			chartName1 := chartURL[lastIndexOfSlash+1:]
	 			//fmt.Printf("ChartName1:%s\n", chartName1)
	 			parts := strings.Split(chartName1, "?")
	 			chartName2 := parts[0]
	 			//fmt.Printf("ChartName2:%s\n", chartName2)

				removePreviousChart(chartName2, cmdRunnerPod)
	 			lsCmd := "ls -l "
	 			executeExecCall(cmdRunnerPod, lsCmd)

	 			// 2. Download the Chart
	 			wgetCmd := "wget --no-check-certificate " + chartURL
	 			fmt.Printf("wget cmd:%s\n", wgetCmd)
	 			executeExecCall(cmdRunnerPod, wgetCmd)
	 			executeExecCall(cmdRunnerPod, lsCmd)

	 			// 3. Rename the Chart to a friendlier name
	 			mvCmd := "mv /" + chartName1 + " /" + chartName2
	 			fmt.Printf("mv cmd:%s\n", mvCmd)
	 			executeExecCall(cmdRunnerPod, mvCmd)
	 			executeExecCall(cmdRunnerPod, lsCmd)

	 			chartName := untarChart(chartName2, cmdRunnerPod, namespace)

	 			return chartName
}

func untarChart(chartName2, cmdRunnerPod, namespace string) string {
				fmt.Printf("Inside untarChart.")

	 			// 4. Untar the Chart file
	 			untarCmd := "tar -xvzf " + chartName2
	  			fmt.Printf("untar cmd:%s\n", untarCmd)
	 			_, op := executeExecCall(cmdRunnerPod, untarCmd)
	 			//fmt.Printf("Untar output:%s",op)
	 			lines := strings.Split(op, "\n")
	 			chartName := ""
	 			parts := strings.Split(lines[0],"/")
	 			//fmt.Printf("ABC:%v",parts)
	 			chartName = strings.TrimSpace(parts[0])
	 			fmt.Printf("Chart Name:%s\n", chartName)

	 			lsCmd := "ls -l "
	 			executeExecCall(cmdRunnerPod, lsCmd)
	 			return chartName
}

func updateStatus(kind, group, version, plural, instance, crdObjNS, targetNS, releaseName, notes string) {
	fmt.Println("Inside updateStatus")

	res := schema.GroupVersionResource{Group: group,
									   Version: version,
									   Resource: plural}
        fmt.Printf("Res:%v\n",res)
	fmt.Printf("kind:%s, group: %s, version:%s, plural:%s, instance:%s, crdObjNS:%s, targetNS:%s, releaseName:%s",
		   kind, group, version, plural, instance, crdObjNS, targetNS, releaseName)
	
	timeout := 300 // try for 5 minutes;
	count := 0
	for {
		obj, err := dynamicClient.Resource(res).Namespace(crdObjNS).Get(context.Background(), instance, metav1.GetOptions{})
		//fmt.Printf("Error:%v\n", err)
		//fmt.Printf("Obj:%v\n",obj)
		if err == nil {
			objData := obj.UnstructuredContent()
			helmrelease := make(map[string]interface{},0)
			// Helm release will be done in the target namespace where the customresource instance
			// is deployed.
			releaseName = strings.ReplaceAll(releaseName, "\n", "")
			helmrelease["helmrelease"] = targetNS + ":" + releaseName + "\n" + notes
			objData["status"] = helmrelease
			//fmt.Printf("objData:%v\n",objData)
			obj.SetUnstructuredContent(objData)
			currentLabels := obj.GetLabels()
			if currentLabels == nil {
				currentLabels = make(map[string]string,1)
			}
			currentLabels["created-by"] = "kubeplus"
			obj.SetLabels(currentLabels)
			dynamicClient.Resource(res).Namespace(crdObjNS).Update(context.Background(), obj, metav1.UpdateOptions{})
			//fmt.Printf("UpdatedObj:%v, err1:%v\n",updatedObj, err1) //add the respective variables if want to print.
			// break out of the for loop
			break
		} else {
			time.Sleep(1 * time.Second)
			count = count + 1
		}
	    if count >= timeout {
		fmt.Printf("CR instance %s not ready till timeout\n.", instance)
		break
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

	obj, _ := dynamicClient.Resource(res).Namespace(namespace).Get(context.Background(), instance, metav1.GetOptions{})
	objData := obj.UnstructuredContent()
	fmt.Printf("objData:%v\n", objData)
	spec := objData["spec"]
	fmt.Printf("Spec:%v\n", spec)
	for key, element := range spec.(map[string]interface{}) {
		//fmt.Printf("Key:%s\n",key)
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

func executeExecCall(runner, command string) (bool, string) {
	//fmt.Println("Inside ExecuteExecCall")
	req := kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(runner).
		Namespace(KUBEPLUS_NAMESPACE).
		SubResource("exec")

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		responseString := "Error: " + err.Error()
		fmt.Printf("Error found trying to Exec command on pod: %s \n", responseString)
		return false, responseString
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
		responseString := "Error: " + err.Error()
		fmt.Printf("Error found trying to Exec command on pod: %s \n", responseString)
		return false, responseString
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
		responseString := "Error: " + execErr.String()
		fmt.Printf("StdErr: %s\n", responseString)
		fmt.Printf("The command %s returned False : %s \n", command, err.Error())
		return false, responseString
	}

	responseString := execOut.String()
	//fmt.Printf("Output! :%v\n", responseString)
	return true, responseString
}
