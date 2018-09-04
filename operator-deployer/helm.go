/*
   Parts of the code here for opening tunnel are based on Helm client source code.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	_ "k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/tlsutil"
	"os"
	"time"
	"log"
	"context"
	"strings"
	//_ "k8s.io/apimachinery"
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
	masterURL  string
	kubeconfig string
	etcdServiceURL string
	deployedCharts []string
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

// getKubeClient creates a Kubernetes config and client for a given kubeconfig context.
func getKubeClient1(context string, kubeconfig string) (*rest.Config, kubernetes.Interface, error) {
	config, err := configForContext(context, kubeconfig)
	if err != nil {
		return nil, nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
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
	if settings.TillerHost == "" {
		//config, client, err := getKubeClient(settings.KubeContext, settings.KubeConfig)
		config, client, err := getKubeClient()
		if err != nil {
			fmt.Errorf("Error: %v", err)
			errToReturn = err
		}
		//fmt.Println("Config:%v", config)
		//fmt.Println("Client:%v", client)

		tunnel, err := portforwarder.New(settings.TillerNamespace, client, config)
		if err != nil {
			panic(err)
			errToReturn = err
		}

		settings.TillerHost = fmt.Sprintf("localhost:%d", tunnel.Local)
		fmt.Println("Created tunnel using local port:", tunnel.Local)
	}

	//fmt.Println("SERVER:", settings.TillerHost)
	return errToReturn
}

func main() {
	settings.TillerNamespace = "kube-system"
	settings.Home = "/.helm"
	settings.TillerConnectionTimeout = 300
	//fmt.Println("ConnectTimeout:%d", settings.TillerConnectionTimeout)
	//settings.KubeConfig = "/Users/devdatta/.kube/config"
	//settings.Home = "/Users/devdatta/.helm"
	
	// Run forever -- if Operator chart is deleted we will create it.
	for {

		fmt.Println("===================================")

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

			//fmt.Println("Helm client: %v", helmClient)

			operatorChartList := getOperatorChartList()
			for _, chartURL := range operatorChartList {
				releases, err := helmClient.ListReleases(
					helm.ReleaseListLimit(10),
					helm.ReleaseListNamespace("default"),
				)
				if err != nil {
					fmt.Errorf("Error: %v", err)
				}
				//fmt.Println("Releases: %v", releases)

				//operatorName := "postgres-crd-v2-chart"
				//operatorVersion := "0.0.2"
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
					//chartURL := "https://s3-us-west-2.amazonaws.com/cloudark-helm-charts/postgres-crd-v2-chart-0.0.2.tgz"

					cmd := &cobra.Command{}
					out := cmd.OutOrStdout()

					err1, crds := newInstallCmd(helmClient, out, chartURL)
					if err1 != nil {
						fmt.Println("%%%%%%%%%")
						fmt.Errorf("Error: %v", err1)
						fmt.Println("%%%%%%%%%")
						errorString := string(err1.Error())
						// Save Error in Etcd
						saveOperatorCRDs(chartURL, []string{"error", errorString})
					} else {
						// Save CRDs in Etcd
						saveOperatorCRDs(chartURL, crds)

						// Save Chart URL in deployedCharts
						deployedCharts = append(deployedCharts, chartURL)
					}
				} else {
					fmt.Println("Operator chart %s %s already deployed", operatorName, operatorVersion)
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
//	return nil
}

func getOperatorChartList() []string {
	cfg := client.Config{
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}
	kapi := client.NewKeysAPI(c)

	resourceKey := "/operatorsToInstall"

	resp, err1 := kapi.Get(context.Background(), resourceKey, nil)
	if err1 != nil {
		fmt.Errorf("Error: %v", err1)
		return []string{}
	} else {
		log.Printf("Get is done. Metadata is %q\n", resp)
		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
		operatorListString := resp.Node.Value
		//fmt.Printf("OperatorListString:%s\n", operatorListString)

		var operatorChartList []string
		if err = json.Unmarshal([]byte(operatorListString), &operatorChartList); err != nil {
			fmt.Errorf("Error: %v", err)
		}
		fmt.Printf("OperatorList:%v\n", operatorChartList)
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
		fmt.Errorf("Error: %v", err)
	}
	kapi := client.NewKeysAPI(c)

	resourceKey := chartURL

	jsonContents, err2 := json.Marshal(&contents)
	if err2 != nil {
		fmt.Errorf("Error: %v", err2)
	}
	resourceValue := string(jsonContents)
	fmt.Printf("Resource Value:%s\n", resourceValue)

	fmt.Printf("Setting %s->%s\n", resourceKey, resourceValue)
	resp, err3 := kapi.Set(context.Background(), resourceKey, resourceValue, nil)
	if err3 != nil {
		fmt.Errorf("Error: %v", err3)
	} else {
		// print common key info
		log.Printf("Set is done. Metadata is %q\n", resp.Node.Value)
	}
	fmt.Println("Exiting saveOperatorCRDs")
}

func checkIfDeployed(chartURL string, rels []*release.Release) (bool, string, string) {
	alreadyDeployed := false
	var operatorName, operatorVersion string
	for _, r := range rels {
		name := r.GetChart().GetMetadata().GetName()
		version := r.GetChart().GetMetadata().GetVersion()
		fmt.Printf("Name:%s Version:%s\n", name, version)
		nameAndVersion := name + "-" + version
		//fmt.Printf("NameVersion:%s\n", nameAndVersion)
		if strings.Contains(chartURL, nameAndVersion) {
			alreadyDeployed = true
			operatorName = name
			operatorVersion = version
		}
		//if name == operatorName && version == operatorVersion {
		//	alreadyDeployed = true
		//}
	}
	return alreadyDeployed, operatorName, operatorVersion
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
