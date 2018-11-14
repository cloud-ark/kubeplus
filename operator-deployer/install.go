/*
   This file is taken from Helm source and has been modified to work as part
   of operator-deployer.

   Original file: helm/cmd/helm/install.go

*/

/*
Copyright The Helm Authors.

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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/strvals"
)

const installDesc = `
This command installs a chart archive.

The install argument must be a chart reference, a path to a packaged chart,
a path to an unpacked chart directory or a URL.

To override values in a chart, use either the '--values' flag and pass in a file
or use the '--set' flag and pass configuration from the command line.  To force string
values in '--set', use '--set-string' instead. In case a value is large and therefore
you want not to use neither '--values' nor '--set', use '--set-file' to read the
single large value from file.

	$ helm install -f myvalues.yaml ./redis

or

	$ helm install --set name=prod ./redis

or

	$ helm install --set-string long_int=1234567890 ./redis

or
    $ helm install --set-file multiline_text=path/to/textfile

You can specify the '--values'/'-f' flag multiple times. The priority will be given to the
last (right-most) file specified. For example, if both myvalues.yaml and override.yaml
contained a key called 'Test', the value set in override.yaml would take precedence:

	$ helm install -f myvalues.yaml -f override.yaml ./redis

You can specify the '--set' flag multiple times. The priority will be given to the
last (right-most) set specified. For example, if both 'bar' and 'newbar' values are
set for a key called 'foo', the 'newbar' value would take precedence:

	$ helm install --set foo=bar --set foo=newbar ./redis


To check the generated manifests of a release without installing the chart,
the '--debug' and '--dry-run' flags can be combined. This will still require a
round-trip to the Tiller server.

If --verify is set, the chart MUST have a provenance file, and the provenance
file MUST pass all verification steps.

There are five different ways you can express the chart you want to install:

1. By chart reference: helm install stable/mariadb
2. By path to a packaged chart: helm install ./nginx-1.2.3.tgz
3. By path to an unpacked chart directory: helm install ./nginx
4. By absolute URL: helm install https://example.com/charts/nginx-1.2.3.tgz
5. By chart reference and repo url: helm install --repo https://example.com/charts/ nginx

CHART REFERENCES

A chart reference is a convenient way of reference a chart in a chart repository.

When you use a chart reference with a repo prefix ('stable/mariadb'), Helm will look in the local
configuration for a chart repository named 'stable', and will then look for a
chart in that repository whose name is 'mariadb'. It will install the latest
version of that chart unless you also supply a version number with the
'--version' flag.

To see the list of chart repositories, use 'helm repo list'. To search for
charts in a repository, use 'helm search'.
`

type installCmd struct {
	name           string
	namespace      string
	valueFiles     valueFiles
	chartPath      string
	dryRun         bool
	disableHooks   bool
	disableCRDHook bool
	replace        bool
	verify         bool
	keyring        string
	out            io.Writer
	client         helm.Interface
	values         []string
	stringValues   []string
	fileValues     []string
	nameTemplate   string
	version        string
	timeout        int64
	wait           bool
	repoURL        string
	username       string
	password       string
	devel          bool
	depUp          bool
	description    string

	certFile string
	keyFile  string
	caFile   string
}

type valueFiles []string

func (v *valueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *valueFiles) Type() string {
	return "valueFiles"
}

func (v *valueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}

// Original newInsallCmd changed and reduced to work with operator-deployer
func installChart(c helm.Interface, out io.Writer, chartName string, chartValues []byte) (error, []string, string) {
	inst := &installCmd{
		out:    out,
		client: c,
	}

	cp, err := locateChartPath(inst.repoURL, inst.username, inst.password, chartName, inst.version, inst.verify, inst.keyring,
		inst.certFile, inst.keyFile, inst.caFile)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	inst.chartPath = cp
	fmt.Printf("Chart PATH:%s\n", cp)

	err1, crds, releaseName := inst.run(chartValues)
	if err1 != nil {
		fmt.Println("***********************************")
		fmt.Printf("Error: %v", err1)
		fmt.Println("***********************************")
		return err1, []string{}, releaseName
	}
	return nil, crds, releaseName
}

func (i *installCmd) run(rawVals []byte) (error, []string, string) {

	var releaseName string
	fmt.Println("CHART PATH: %s\n", i.chartPath)
	str1 := fmt.Sprintf("%s", rawVals)
	fmt.Println("Chart Values:%s\n", str1)
	crds := make([]string, 0)

	if i.namespace == "" {
		i.namespace = defaultNamespace()
	}

	if i.nameTemplate != "" {
		var err error
		i.name, err = generateName(i.nameTemplate)
		if err != nil {
			return err, crds, releaseName
		}
		// Print the final name so the user knows what the final name of the release is.
		fmt.Printf("FINAL NAME: %s\n", i.name)
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	chartRequested, err := chartutil.Load(i.chartPath)
	if err != nil {
		return err, crds, releaseName
	}

	if req, err := chartutil.LoadRequirements(chartRequested); err == nil {
		// If checkDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/kubernetes/helm/issues/2209
		if err := checkDependencies(chartRequested, req); err != nil {
			if i.depUp {
				man := &downloader.Manager{
					Out:       i.out,
					ChartPath: i.chartPath,
					HelmHome:  settings.Home,
					//Keyring:    defaultKeyring(),
					SkipUpdate: false,
					Getters:    getter.All(settings),
				}
				if err := man.Update(); err != nil {
					return err, crds, releaseName
				}

				// Update all dependencies which are present in /charts.
				chartRequested, err = chartutil.Load(i.chartPath)
				if err != nil {
					return err, crds, releaseName
				}
			} else {
				return err, crds, releaseName
			}
		}
	} else if err != chartutil.ErrRequirementsNotFound {
		return fmt.Errorf("cannot load requirements: %v\n", err), crds, releaseName
	}

	res, err := i.client.InstallReleaseFromChart(
		chartRequested,
		i.namespace,
		helm.ValueOverrides(rawVals),
		helm.ReleaseName(i.name),
		helm.InstallDryRun(i.dryRun),
		helm.InstallReuseName(i.replace),
		helm.InstallDisableHooks(i.disableHooks),
		helm.InstallDisableCRDHook(i.disableCRDHook),
		helm.InstallTimeout(i.timeout),
		helm.InstallWait(i.wait),
		helm.InstallDescription(i.description))
	if err != nil {
		fmt.Printf("Error1:%v\n", err)
		return err, crds, releaseName
	}

	rel := res.GetRelease()

	if rel == nil {
		return nil, crds, releaseName
	}

	i.printRelease(rel)

	releaseName = rel.Name

	// Print the status like status command does
	status, err := i.client.ReleaseStatus(rel.Name)
	if err != nil {
		return err, crds, releaseName
	}
	//PrintStatus(i.out, status)
	fmt.Println("Status:%v", status)
	fmt.Println("==========")
	fmt.Println("==========")
	fmt.Println("==========")
	fmt.Println("**********")
	fmt.Println("Printing Resources")

	resources := status.GetInfo().GetStatus().GetResources()
	lines := strings.Split(resources, "\n")
	//num := len(lines)
	//fmt.Println("Num of Lines:%d", num)
	fmt.Printf("Lines:%v\n", lines)

	startCustomResourceLines := false
	for _, line := range lines {
		if startCustomResourceLines {
			if !strings.HasPrefix(line, "NAME") && !strings.HasPrefix(line, "==>") && line != " " && line != "" {
				parts := strings.Split(line, " ")
				crds = append(crds, string(parts[0]))
			}
			if strings.HasPrefix(line, "==>") {
				startCustomResourceLines = false
			}
		}
		if strings.Contains(line, "v1beta1/CustomResourceDefinition") {
			fmt.Println("LINE:%s", line)
			startCustomResourceLines = true
		}
	}

	fmt.Println("Resources:%s", crds)

	return nil, crds, releaseName
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

// vals merges values from files specified via -f/--values and
// directly via --set or --set-string or --set-file, marshaling them to YAML
func vals(valueFiles valueFiles, values []string, stringValues []string, fileValues []string, CertFile, KeyFile, CAFile string) ([]byte, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		var bytes []byte
		var err error
		if strings.TrimSpace(filePath) == "-" {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			bytes, err = readFile(filePath, CertFile, KeyFile, CAFile)
		}

		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s\n", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s\n", err)
		}
	}

	// User specified a value via --set-string
	for _, value := range stringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-string data: %s\n", err)
		}
	}

	// User specified a value via --set-file
	for _, value := range fileValues {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := readFile(string(rs), CertFile, KeyFile, CAFile)
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, base, reader); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-file data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

// printRelease prints info about a release if the Debug is true.
func (i *installCmd) printRelease(rel *release.Release) {
	if rel == nil {
		return
	}
	// TODO: Switch to text/template like everything else.
	fmt.Fprintf(i.out, "NAME:   %s\n", rel.Name)
	if settings.Debug {
		fmt.Println("---::--- %v", rel)
		//printRelease(i.out, rel)
	}
}

// locateChartPath looks for a chart directory in known places, and returns either the full path or an error.
//
// This does not ensure that the chart is well-formed; only that the requested filename exists.
//
// Order of resolution:
// - current working directory
// - if path is absolute or begins with '.', error out here
// - chart repos in $HELM_HOME
// - URL
//
// If 'verify' is true, this will attempt to also verify the chart.
func locateChartPath(repoURL, username, password, name, version string, verify bool, keyring,
	certFile, keyFile, caFile string) (string, error) {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)

	if fi, err := os.Stat(name); err == nil {
		abs, err := filepath.Abs(name)

		if err != nil {
			return abs, err
		}

		if verify {
			if fi.IsDir() {
				return "", errors.New("cannot verify a directory")
			}
			if _, err := downloader.VerifyChart(abs, keyring); err != nil {
				return "", err
			}
		}

		return abs, nil
	}

	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	crepo := filepath.Join(settings.Home.Repository(), name)
	if _, err := os.Stat(crepo); err == nil {
		return filepath.Abs(crepo)
	}

	dl := downloader.ChartDownloader{
		HelmHome: settings.Home,
		Out:      os.Stdout,
		Keyring:  keyring,
		Getters:  getter.All(settings),
		Username: username,
		Password: password,
	}

	if verify {
		dl.Verify = downloader.VerifyAlways
	}

	if repoURL != "" {
		chartURL, err := repo.FindChartInAuthRepoURL(repoURL, username, password, name, version,
			certFile, keyFile, caFile, getter.All(settings))
		if err != nil {
			return "", err
		}
		name = chartURL
	}

	if _, err := os.Stat(settings.Home.Archive()); os.IsNotExist(err) {
		os.MkdirAll(settings.Home.Archive(), 0744)
	}

	fmt.Printf("Name:%s, Version:%s\n", name, version)
	filename, _, err := dl.DownloadTo(name, version, settings.Home.Archive())

	fmt.Printf("Filename:%s\n", filename)
	fmt.Printf("Name:%s\n", name)
	if err != nil {
		fmt.Errorf("Error: %v", err)
		panic(err)
	}

	if err == nil {
		lname, err := filepath.Abs(filename)
		if err != nil {
			fmt.Println("22")
			return filename, err
		}
		fmt.Println("Fetched %s to %s\n", name, filename)
		return lname, nil
	} else if settings.Debug {
		fmt.Println("23")
		return filename, err
	}

	return filename, fmt.Errorf("failed to download %q (hint: running `helm repo update` may help)", name)
}

func generateName(nameTemplate string) (string, error) {
	t, err := template.New("name-template").Funcs(sprig.TxtFuncMap()).Parse(nameTemplate)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = t.Execute(&b, nil)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func defaultNamespace() string {
	if ns, _, err := kube.GetConfig(settings.KubeContext, settings.KubeConfig).Namespace(); err == nil {
		return ns
	}
	return "default"
}

func checkDependencies(ch *chart.Chart, reqs *chartutil.Requirements) error {
	missing := []string{}

	deps := ch.GetDependencies()
	for _, r := range reqs.Dependencies {
		found := false
		for _, d := range deps {
			if d.Metadata.Name == r.Name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, r.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}

//readFile load a file from the local directory or a remote file with a url.
func readFile(filePath, CertFile, KeyFile, CAFile string) ([]byte, error) {
	u, _ := url.Parse(filePath)
	p := getter.All(settings)

	// FIXME: maybe someone handle other protocols like ftp.
	getterConstructor, err := p.ByScheme(u.Scheme)

	if err != nil {
		return ioutil.ReadFile(filePath)
	}

	getter, err := getterConstructor(filePath, CertFile, KeyFile, CAFile)
	if err != nil {
		return []byte{}, err
	}
	data, err := getter.Get(filePath)
	return data.Bytes(), err
}
