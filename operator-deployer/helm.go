/*
   Parts of the code here for opening tunnel are based on Helm client source code.
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/client"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	_ "k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/tlsutil"
	"log"
	"os"
	"strings"
	"time"
)

var (
	settings      helm_env.EnvSettings
	tlsServerName string // overrides the server name used to verify the hostname on the returned certificates from the server.
	tlsCaCertFile string // path to TLS CA certificate file
	tlsCertFile   string // path to TLS certificate file
	tlsKeyFile    string // path to TLS key file
	tlsVerify     bool   // enable TLS and verify remote certificates
	tlsEnable     bool   // enable TLS
)

var (
	masterURL      string
	kubeconfig     string
	etcdServiceURL string
	deployedCharts []string
        operatorsToDeleteKey  = "/operatorsToDelete"
        operatorsToInstallKey = "/operatorsToInstall"
)

func init() {
	etcdServiceURL = "http://localhost:2379"
	deployedCharts = make([]string, 0)
}

func configForContext(context string, kubeconfig string) (*rest.Config, error) {
	config, err := kube.GetConfig(context, kubeconfig).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %s", context, err)
	}
	return config, nil
}

func getKubeClient() (*rest.Config, kubernetes.Interface, error) {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		fmt.Println("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Println("Error building kubernetes clientset: %s", err.Error())
	}
	return cfg, kubeClient, nil
}

func setupConnection() error {
	var errToReturn error
	//if settings.TillerHost == "" {
	config, client, err := getKubeClient()
	if err != nil {
		fmt.Printf("Error: %v", err)
		errToReturn = err
	}

	tunnel, err := portforwarder.New(settings.TillerNamespace, client, config)
	if err != nil {
		panic(err)
		errToReturn = err
	}

	settings.TillerHost = fmt.Sprintf("localhost:%d", tunnel.Local)
	fmt.Println("Created tunnel using local port:", tunnel.Local)
	//}
	return errToReturn
}

func main() {
	settings.TillerNamespace = "kube-system"
	settings.Home = "/.helm"
	settings.TillerConnectionTimeout = 300

	// Run forever
	for {

		operatorsToInstall := getOperatorChartList(operatorsToInstallKey)
		if len(operatorsToInstall) > 0 {
			//fmt.Printf("Operators to install:%v\n", operatorsToInstall)
		}
		operatorsToDelete := getOperatorChartList(operatorsToDeleteKey)
		if len(operatorsToDelete) > 0 {
			//fmt.Printf("Operators to delete:%v\n", operatorsToDelete)
		}

		newOperators := subtract(operatorsToInstall, deployedCharts)

		if len(newOperators) > 0 || len(operatorsToDelete) > 0 {

			connectionError := setupConnection()
			if connectionError == nil {

				options := []helm.Option{helm.Host(settings.TillerHost), helm.ConnectTimeout(settings.TillerConnectionTimeout)}

				if tlsVerify || tlsEnable {
					if tlsCaCertFile == "" {
						tlsCaCertFile = settings.Home.TLSCaCert()
					}
					if tlsCertFile == "" {
						tlsCertFile = settings.Home.TLSCert()
					}
					if tlsKeyFile == "" {
						tlsKeyFile = settings.Home.TLSKey()
					}
					fmt.Println("Host=%q, Key=%q, Cert=%q, CA=%q\n", tlsKeyFile, tlsCertFile, tlsCaCertFile)
					tlsopts := tlsutil.Options{
						ServerName:         tlsServerName,
						KeyFile:            tlsKeyFile,
						CertFile:           tlsCertFile,
						InsecureSkipVerify: true,
					}
					if tlsVerify {
						tlsopts.CaCertFile = tlsCaCertFile
						tlsopts.InsecureSkipVerify = false
					}
					tlscfg, err := tlsutil.ClientConfig(tlsopts)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					options = append(options, helm.WithTLS(tlscfg))
				}

				helmClient := helm.NewClient(options...)

				// Delete Operators
				for _, chartURL := range operatorsToDelete {
					releaseName := getReleaseName("release-" + chartURL)
					fmt.Printf("Release Name to delete:%s\n", releaseName)
					if releaseName != "" {
						err1 := deleteChart(helmClient, releaseName)
						if err1 != nil {
						   fmt.Printf("Error deleting chart %s %s", releaseName, err1)
						}
						// Delete Helm Release ConfigMap
						deleteHelmReleaseConfigMap(releaseName)

						// Delete Chart URL from deployedCharts
						newChartList := make([]string, 0)
						for _, depChart := range deployedCharts {
							if depChart != chartURL {
								newChartList = append(newChartList, depChart)
							}
						}
						deployedCharts = newChartList
					} else {
						fmt.Printf("Did not find any release for %s\n", chartURL)
					}
				}

				// Install Operators
				operatorsToInstall = subtract(operatorsToInstall, operatorsToDelete)
				if len(operatorsToInstall) > 0 {
					fmt.Printf("Effective Operators to install:%v\n", operatorsToInstall)
				}
				for _, chartURL := range operatorsToInstall {
					releases, err := helmClient.ListReleases(
						helm.ReleaseListLimit(10),
						helm.ReleaseListNamespace("default"),
					)
					if err != nil {
						fmt.Printf("Error: %v", err)
					}

					alreadyDeployed, operatorName, operatorVersion := checkIfDeployed(chartURL, releases.GetReleases())

					// releases.GetReleases() may return empty if connection to Tiller breaks
					// If that happens alreadyDeployed will be false (its default value.)
					// But the chart may actually be deployed from previous run. So check in our deployedCharts list
					// to avoid re-triggering chart deployment.
					// TODO(devdattakulkarni): Revisit when supporting Chart upgrades.
					if !alreadyDeployed {
						for _, deployedChart := range deployedCharts {
							if deployedChart == chartURL {
								alreadyDeployed = true
							}
						}
					}

					if !alreadyDeployed {
						fmt.Println("Installing chart.")
						fmt.Printf("Chart URL%s\n", chartURL)

						cmd := &cobra.Command{}
						out := cmd.OutOrStdout()

						err, chartValues := getChartValues(chartURL)
						if err != nil {
							panic(err)
						}
						fmt.Printf("ChartValues:%v\n", chartValues)
						err1, crds, releaseName := installChart(helmClient, out, chartURL, chartValues)
						if err1 != nil {
							fmt.Println("%%%%%%%%%")
							fmt.Printf("Error: %v", err1)
							fmt.Println("%%%%%%%%%")
							errorString := string(err1.Error())
							// Save Error in Etcd
							saveOperatorCRDs(chartURL, []string{"error", errorString})
						} else {
							// Save CRDs in Etcd
							saveOperatorCRDs(chartURL, crds)

							// Save release name
							saveReleaseName("release-"+chartURL, releaseName)

							// Save Chart URL in deployedCharts
							deployedCharts = append(deployedCharts, chartURL)
						}
					} else {
						fmt.Println("Operator chart %s %s already deployed", operatorName, operatorVersion)
					}
				}

				// Set the operatorsToInstall key with the new List
				updateInstallList(operatorsToInstallKey, operatorsToInstall)

				// After all Operators are deleted, reset the list in etcd
				if len(operatorsToDelete) > 0 {
					emptyList := make([]string, 0)
					storeList(operatorsToDeleteKey, emptyList)
				}
			}
		} // If len() || len()
		time.Sleep(time.Second * 5)
	}
}

func deleteHelmReleaseConfigMap(releaseName string) {

     _, kubeclientset, err := getKubeClient()
     if err != nil {
     	fmt.Printf("Error: %v", err)
     }

     configMapName := releaseName + ".v1"
     err1 := kubeclientset.CoreV1().ConfigMaps("kube-system").Delete(configMapName, &metav1.DeleteOptions{})
     if err1 != nil {
        fmt.Printf("Error in deleting Helm ConfigMap:%s\n", err1.Error())
     }
}

func getReleaseName(resourceKey string) string {
	releaseName := getSingleValue(resourceKey)
	return releaseName
}

func getSingleValue(resourceKey string) string {
	var releaseName string
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	kapi := client.NewKeysAPI(c)

	resp, err1 := kapi.Get(context.Background(), resourceKey, nil)
	if err1 != nil {
		fmt.Printf("Error: %v\n", err1)
		return ""
	} else {
		//log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
		releaseName = resp.Node.Value
		return releaseName
	}
}

func saveReleaseName(resourceKey, releaseName string) {
	storeSingleValue(resourceKey, releaseName)
}

func updateInstallList(resourceKey string, operatorList []string) {
	storeList(resourceKey, operatorList)
}

func storeList(resourceKey string, dataList []string) {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}
	kapi := client.NewKeysAPI(c)

	jsonOperatorList, err2 := json.Marshal(&dataList)
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

func storeSingleValue(resourceKey, resourceData string) {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	fmt.Printf("Setting %s->%s\n", resourceKey, resourceData)
	_, err1 := kapi.Set(context.Background(), resourceKey, resourceData, nil)
	if err1 != nil {
		log.Fatal(err1)
	} else {
		// print common key info
		//log.Printf("Set is done. Metadata is %q\n", resp.Node.Value)
	}
}

func subtract(installList, deleteList []string) []string {
	operatorsToInstall := make([]string, 0)
	for _, installChart := range installList {
		found := false
		for _, deleteChart := range deleteList {
			if installChart == deleteChart {
				found = true
			}
		}
		if !found {
			operatorsToInstall = append(operatorsToInstall, installChart)
		}
	}
	return operatorsToInstall
}

func getChartValues(chartURL string) (error, []byte) {
	valuesData := make([]byte, 0)
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	kapi := client.NewKeysAPI(c)

	resourceKey := "/chartvalues/" + chartURL

	resp, err1 := kapi.Get(context.Background(), resourceKey, nil)
	if err1 != nil {
		fmt.Printf("Error: %v\n", err1)
		return err1, valuesData
	} else {
		//log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
		valuesString := resp.Node.Value
		fmt.Printf("ValuesString:%s\n", valuesString)

		if valuesString != "null" {
			valuesMap := map[string]interface{}{}
			if err = json.Unmarshal([]byte(valuesString), &valuesMap); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Printf("ValuesMap:%v\n", valuesMap)
			valuesData, err1 := yaml.Marshal(valuesMap)
			if err1 != nil {
				return err1, valuesData
			}
			return nil, valuesData
		} else {
			return nil, nil
		}
	}
}

func getOperatorChartList(resourceKey string) []string {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	kapi := client.NewKeysAPI(c)

	resp, err1 := kapi.Get(context.Background(), resourceKey, nil)
	if err1 != nil {
		return []string{}
	} else {
		//log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
		operatorListString := resp.Node.Value

		var operatorChartList []string
		if err = json.Unmarshal([]byte(operatorListString), &operatorChartList); err != nil {
			fmt.Printf("Error: %v", err)
		}
		return operatorChartList
	}
}

func saveOperatorCRDs(chartURL string, contents []string) {
	fmt.Println("Entering saveOperatorCRDs")
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	kapi := client.NewKeysAPI(c)

	resourceKey := chartURL

	jsonContents, err2 := json.Marshal(&contents)
	if err2 != nil {
		fmt.Printf("Error: %v", err2)
	}
	resourceValue := string(jsonContents)

	//fmt.Printf("Setting %s->%s\n", resourceKey, resourceValue)
	_, err3 := kapi.Set(context.Background(), resourceKey, resourceValue, nil)
	if err3 != nil {
		fmt.Printf("Error: %v", err3)
	} else {
		//log.Printf("Set is done. Metadata is %q\n", resp.Node.Value)
	}
	fmt.Println("Exiting saveOperatorCRDs")
}

func checkIfDeployed(chartURL string, rels []*release.Release) (bool, string, string) {
	alreadyDeployed := false
	var operatorName, operatorVersion string
	for _, r := range rels {
		name := strings.TrimSuffix(r.GetChart().GetMetadata().GetName(), "\n")
		version := strings.TrimSuffix(r.GetChart().GetMetadata().GetVersion(), "\n")
		fmt.Printf("Name:%s Version:%s\n", name, version)
		nameAndVersion := name + "-" + version

		if strings.Contains(chartURL, nameAndVersion) {
			alreadyDeployed = true
			operatorName = name
			operatorVersion = version
		}
	}
	return alreadyDeployed, operatorName, operatorVersion
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
