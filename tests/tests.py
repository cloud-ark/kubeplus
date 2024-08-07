import unittest
import sys
import json
import subprocess
import sys
import os
import yaml
import time
from kubernetes import config
from kubernetes.client import Configuration
from kubernetes.client.api import core_v1_api
from kubernetes.stream import portforward
import select
from bs4 import BeautifulSoup

class TestKubePlus(unittest.TestCase):

    @classmethod
    def run_command(self, cmd):
        cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
        out = cmdOut[0].decode('utf-8')
        err = cmdOut[1].decode('utf-8')
        show_output = os.getenv("KUBEPLUS_TEST_OUTPUT","")
        if show_output == "yes":
            print(cmd)
            print(out)
            print("---")
            print(err)
        return out, err

    @classmethod
    def _is_kubeplus_running(self):
        cmd = 'kubectl get pods -A'
        out, err = TestKubePlus.run_command(cmd)
        for line in out.split("\n"):
            if 'kubeplus' in line and 'Running' in line:
                return True
        return False

    @classmethod
    def _is_kyverno_running(self):
        cmd = 'kubectl get pods -A'
        out, err = TestKubePlus.run_command(cmd)
        for line in out.split("\n"):
            if 'kyverno' in line and 'Running' in line:
                return True
        return False

    def test_create_res_comp_for_chart_with_ns(self):
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        cmd = "kubectl create -f wordpress-service-composition-chart-withns.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        # print("Out:" + out)
        # print("Err:" + err)
        self.assertTrue('Namespace object is not allowed in the chart' in err)

    def test_create_res_comp_for_chart_with_shared_storage(self):
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        cmd = "kubectl create -f storage-class-fast.yaml"
        TestKubePlus.run_command(cmd)

        cmd = "kubectl create -f storage-isolation/wordpress-service-composition.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        # print("Out:" + out)
        # print("Err:" + err)
        self.assertTrue('Storage class with reclaim policy Retain not allowed' in err)

        cmd = "kubectl delete -f storage-class-fast.yaml"
        TestKubePlus.run_command(cmd)

    def test_create_res_comp_with_incomplete_resource_quota(self):
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        cmd = "kubectl create -f resource-quota/wordpress-service-composition-1.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        # print("Out:" + out)
        # print("Err:" + err)
        self.assertTrue(
            'If quota is specified, specify all four values: requests.cpu, requests.memory, limits.cpu, limits.memory' in err)

    def test_license_plugin(self):
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        start_clean = "kubectl delete ns hs1"
        TestKubePlus.run_command(start_clean)

        kubeplus_home = os.getenv("KUBEPLUS_HOME")
        path = os.getenv("PATH")
        if kubeplus_home == '':
            print("Skipping the test as KUBEPLUS_HOME is not set.")
            return

        cmd = "kubectl upload chart ../examples/multitenancy/hello-world/hello-world-chart-0.0.3.tgz ../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        cmd = "kubectl create -f ../examples/multitenancy/hello-world/hello-world-service-composition-localchart.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)

        crd = "helloworldservices.platformapi.kubeplus"
        crd_installed = self._check_crd_installed(crd)
        if not crd_installed:
            print("CRD " + crd + " not installed. Exiting this test.")
            return

        # Test with license that restricts number of instances (1)
        cmd = "kubectl license create HelloWorldService license.txt -n 1  -k ../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        cmd = "kubectl create -f ../examples/multitenancy/hello-world/hs1.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        # Second instance creation should be denied
        cmd = "kubectl create -f ../examples/multitenancy/hello-world/hs2.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        self.assertTrue("Allowed number of instances reached" in err)
        cmd = "kubectl license delete HelloWorldService -k ../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        # Test with expired license
        cmd = "kubectl license create HelloWorldService license.txt -e 01/01/2024 -k ../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)
        cmd = "kubectl create -f ../examples/multitenancy/hello-world/hs2.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        self.assertTrue("License expired (expiry date):01/01/2024" in err)
        cmd = "kubectl license delete HelloWorldService -k ../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        # Test with expired license and restriction on number of instances
        cmd = "kubectl license create HelloWorldService license.txt -n 1 -e 01/01/2024 -k ../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)
        cmd = "kubectl create -f ../examples/multitenancy/hello-world/hs2.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        self.assertTrue("License expired (expiry date):01/01/2024" in err)
        self.assertTrue("Allowed number of instances reached" in err)
        cmd = "kubectl license delete HelloWorldService -k ../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)
        
        # cleanup
        cmd = "kubectl delete -f ../examples/multitenancy/hello-world/hello-world-service-composition-localchart.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)


    def test_application_update(self):
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        start_clean = "kubectl delete ns hs1"
        TestKubePlus.run_command(start_clean)

        kubeplus_home = os.getenv("KUBEPLUS_HOME")
        # print("KubePlus home:" + kubeplus_home)
        path = os.getenv("PATH")
        # print("Path:" + path)
        if kubeplus_home == '':
            print("Skipping the test as KUBEPLUS_HOME is not set.")
            return

        cmd = "kubectl upload chart ../examples/multitenancy/hello-world/hello-world-chart-0.0.3.tgz ../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        cmd = "kubectl create -f ../examples/multitenancy/hello-world/hello-world-service-composition-localchart.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)

        crd = "helloworldservices.platformapi.kubeplus"
        crd_installed = self._check_crd_installed(crd)
        if not crd_installed:
            print("CRD " + crd + " not installed. Exiting this test.")
            return

        cmd = "kubectl apply -f ../examples/multitenancy/hello-world/hs1.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)
        all_running = False
        cmd = "kubectl get pods -n hs1"

        target_pod_count = 1
        pods, count, all_running = self._check_pod_status(cmd, target_pod_count)
        if count == target_pod_count:
            self.assertTrue(True)
        else:
            self.assertTrue(False)

        time.sleep(10)
        cmd = "kubectl replace -f ../examples/multitenancy/hello-world/hs1-replicas-2.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)
        all_running = False
        cmd = "kubectl get pods -n hs1"

        target_pod_count = 2
        pods, count, all_running = self._check_pod_status(cmd, target_pod_count)
        if count == target_pod_count:
            self.assertTrue(True)
        else:
            self.assertTrue(False)

        time.sleep(10)
        cmd = "kubectl delete -f ../examples/multitenancy/hello-world/hs1-replicas-2.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        cmd = "kubectl delete -f ../examples/multitenancy/hello-world/hello-world-service-composition-localchart.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)

    def test_application_upgrade(self):

        # assume appropriate plugins installation and PATH update

        # helper methods

        # from kubernetes python client example -- pod_portforward.py
        def make_http_request(port):
            http = pf.socket(port)
            http.setblocking(True)
            http.sendall(b'GET /users HTTP/1.1\r\n')
            http.sendall(b'Host: 127.0.0.1\r\n')
            http.sendall(b'Connection: close\r\n')
            http.sendall(b'Accept: */*\r\n')
            http.sendall(b'\r\n')
            response = b''
            while True:
                select.select([http], [], [])
                data = http.recv(1024)
                if not data:
                    break
                response += data
            http.close()
            
            error = pf.error(port)
            if error is None:
                print("No port forward errors on port %d." % port)
            else:
                print(f"Port {port} has the following error: {error}")
            http.close()
            return response.decode('utf-8')
        
        # get number of users from HTML response
        def count_users(response):
            num_users = 0
            if "No users" not in response:
                soup = BeautifulSoup(response, 'lxml')
                td_tags = soup.find_all('td')
                for td in td_tags:
                    users = td.get_text(strip=True)
                    for _ in users.split(" "):
                        num_users += 1
            return num_users

        def cleanup():
            cmd = 'kubectl delete -f ./application-upgrade/resource-composition-localchart.yaml --kubeconfig=./application-upgrade/provider.conf'
            TestKubePlus.run_command(cmd)

            # restore chart
            data = None
            with open('./application-upgrade/resource-composition-localchart.yaml', 'r') as f:
                data = yaml.safe_load(f)
        
            data['spec']['newResource']['chartURL'] = 'file:///resource-composition-0.0.1.tgz'

            with open('./application-upgrade/resource-composition-localchart.yaml', 'w') as f:   
                yaml.safe_dump(data, f, default_flow_style=False)

        # preliminary checks
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        if os.getenv("KUBEPLUS_HOME") == '':
            print("Skipping test as KUBEPLUS_HOME is not set.")
            return
        
        # add Kubeplus provider
        cmd = "cp ../kubeplus-saas-provider.json ./application-upgrade/provider.conf"
        TestKubePlus.run_command(cmd)

        # chart uploads
        cmd = "kubectl upload chart ./application-upgrade/resource-composition-0.0.1.tgz ./application-upgrade/provider.conf"
        TestKubePlus.run_command(cmd)
        cmd = "kubectl upload chart ./application-upgrade/resource-composition-0.0.2.tgz ./application-upgrade/provider.conf"
        TestKubePlus.run_command(cmd)

        # create API
        cmd = "kubectl create -f ./application-upgrade/resource-composition-localchart.yaml --kubeconfig=./application-upgrade/provider.conf"
        TestKubePlus.run_command(cmd)

        # CRDs check
        crd = "webappservices.platformapi.kubeplus"
        crd_installed = self._check_crd_installed(crd)
        if not crd_installed:
            print("CRD " + crd + " not installed. Exiting this test.")
            return

        # create app instance
        cmd = "kubectl create -f ./application-upgrade/tenant1.yaml --kubeconfig=./application-upgrade/provider.conf"
        out, err = TestKubePlus.run_command(cmd)

        namespace = 'bwa-tenant1'
        port = 5000

        # let the app pods come up
        wait_time = 60
        time.sleep(wait_time)

        # grab name of deployed pod
        cmd = "kubectl get pods -n %s" % namespace
        out, err = TestKubePlus.run_command(cmd)
        name = None
        for line in out.split("\n"):
            if "web-app" in line:
                parts = line.split(" ")
                for part in parts:
                    if "web-app" in part:
                        name = part.strip()
                break

        if name == None:
            print("Pod did not come up even after waiting " + str(wait_time) + " seconds.")
            print("Skipping rest of the test.")
            cleanup()

        # port forwarding
        # CLI: kubectl port-forward pod-name -n bwa-tenant1 5000:5000
        config.load_kube_config()
        c = Configuration.get_default_copy()
        c.assert_hostname = False
        Configuration.set_default(c)
        api_instance = core_v1_api.CoreV1Api()

        
        # https://github.com/kubernetes-client/python/blob/master/examples/pod_portforward.py
        pf = portforward(api_instance.connect_get_namespaced_pod_portforward,
                         name,
                         namespace,
                         ports=str(port))

        response = make_http_request(port).strip(" ")
        num_users_first = count_users(response)
        
        # upgrade to version 2
        data = None
        with open('./application-upgrade/resource-composition-localchart.yaml', 'r') as f:
            data = yaml.safe_load(f)
        
        data['spec']['newResource']['chartURL'] = 'file:///resource-composition-0.0.2.tgz'

        with open('./application-upgrade/resource-composition-localchart.yaml', 'w') as f:   
            yaml.safe_dump(data, f, default_flow_style=False)

        cmd = 'kubectl apply -f ./application-upgrade/resource-composition-localchart.yaml --kubeconfig=./application-upgrade/provider.conf'
        out, err = TestKubePlus.run_command(cmd)

        # sleep to let the pods run
        time.sleep(60)

        # grab name of deployed pod
        cmd = "kubectl get pods -n %s" % namespace
        out, err = TestKubePlus.run_command(cmd)
        name = None
        for line in out.split("\n"):
            if "web-app" in line:
                parts = line.split(" ")
                for part in parts:
                    if "web-app" in part:
                        name = part.strip()

        # check data at port after upgrade -- should be more users 
        pf = portforward(api_instance.connect_get_namespaced_pod_portforward,
                         name,
                         namespace,
                         ports=str(port))
        
        response = make_http_request(port).strip(" ")
        num_users_second = count_users(response)

        cleanup()
        
        # check if upgrade worked
        self.assertTrue(num_users_second > num_users_first)
        

    def _check_pod_status(self, cmd, num_of_pods):
        all_running = False
        pods = []
        timer = 0
        count = 0
        while not all_running and timer < 120:
            timer = timer + 1
            out, err = TestKubePlus.run_command(cmd)
            for line in out.split("\n"):
                if 'Running' in line or 'Pending' in line or 'ContainerCreating' in line:
                    count = count + 1
                if 'NAME' not in line:
                    parts = line.split(" ")
                    pod = parts[0].strip()
                    if pod != '' and pod not in pods:
                        pods.append(pod)
            if count == num_of_pods:
                all_running = True
                break

        return pods, count, all_running

    def _check_crd_installed(self, crd):
        installed = False
        cmd = "kubectl get crds"
        timer = 0
        while not installed and timer < 120:
            out, err = TestKubePlus.run_command(cmd)
            if crd in out:
                installed = True
            else:
                time.sleep(1)
                timer = timer + 1
        return installed

    def test_force_delete_application(self):
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        start_clean = "kubectl delete ns tenant1"
        TestKubePlus.run_command(start_clean)

        create_ns = "kubectl create ns tenant1"
        TestKubePlus.run_command(create_ns)

        cmd1 = "kubectl create -f wordpress-service-composition-chart-nopodpolicies.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd1)

        crd = "wordpressservices.platformapi.kubeplus"
        crd_installed = self._check_crd_installed(crd)
        if not crd_installed:
            print("CRD " + crd + " not installed. Exiting this test.")
            return

        cmd = "kubectl create -f tenant1.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        cmd = "kubectl label WordpressService tenant1 delete=true"
        TestKubePlus.run_command(cmd)

        cmd = "kubectl delete -f tenant1.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        self.assertTrue(err == "")

        cmd = "kubectl delete -f wordpress-service-composition-chart-nopodpolicies.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        self.assertTrue(err == "")

        clean_up = "kubectl delete ns tenant1"
        TestKubePlus.run_command(clean_up)

    def test_res_comp_with_no_podpolicies(self):
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        start_clean = "kubectl delete ns tenant1"
        TestKubePlus.run_command(start_clean)

        cmd = "kubectl create -f wordpress-service-composition-chart-nopodpolicies.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)
        crd = "wordpressservices.platformapi.kubeplus"
        crd_installed = self._check_crd_installed(crd)
        if not crd_installed:
            print("CRD " + crd + " not installed. Exiting this test.")
            return

        cmd = "kubectl create -f tenant1.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        all_running = False
        cmd = "kubectl get pods -n tenant1"

        target_pod_count = 2
        pods, count, all_running = self._check_pod_status(cmd, target_pod_count)

        if count < target_pod_count:
            print("Application Pod not started..")
        else:
            # print(pods)
            # Check container configs
            for pod in pods:
                cmd = "kubectl get pod " + pod + " -n tenant1 -o json "
                out, err = TestKubePlus.run_command(cmd)
                json_obj = json.loads(out)
                # print(json_obj)
                # print(json_obj['spec']['containers'][0])
                resources = json_obj['spec']['containers'][0]['resources']
                if not resources:
                    self.assertTrue(True)
                else:
                    self.assertTrue(False)

        # clean up
        # wait and then clean up
        time.sleep(30)
        cmd = "kubectl delete -f tenant1.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        cmd = "kubectl delete -f wordpress-service-composition-chart-nopodpolicies.yaml --kubeconfig=../kubeplus-saas-provider.json"
        TestKubePlus.run_command(cmd)

        removed = False
        cmd = "kubectl get crds"
        timer = 0
        while not removed and timer < 60:
            timer = timer + 1
            out, err = TestKubePlus.run_command(cmd)
            if 'wordpressservices.platformapi.kubeplus' not in out:
                removed = True
            else:
                time.sleep(1)
    
    def test_appstatus_plugin(self):
        kubeplus_home = os.getenv("KUBEPLUS_HOME")
        provider = kubeplus_home + '/kubeplus-saas-provider.json'
        
        def cleanup():
            cmd = "kubectl delete -f ../examples/multitenancy/hello-world/hs1.yaml --kubeconfig=%s" % provider
            TestKubePlus.run_command(cmd)
            cmd = "kubectl delete -f ../examples/multitenancy/hello-world/hello-world-service-composition-localchart.yaml --kubeconfig=%s" % provider
            TestKubePlus.run_command(cmd)

        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        if os.getenv("KUBEPLUS_HOME") == '':
            print("Skipping test as KUBEPLUS_HOME is not set.")
            return

        # register HelloWorldService API
        cmd = "kubectl create -f ../examples/multitenancy/hello-world/hello-world-service-composition-localchart.yaml --kubeconfig=%s" % provider
        TestKubePlus.run_command(cmd)

        # check CRD installation
        crd = "helloworldservices.platformapi.kubeplus"
        crd_installed = self._check_crd_installed(crd)
        if not crd_installed:
            print("CRD " + crd + " not installed. Exiting this test.")
            return
        
        # create app instance
        cmd = "kubectl create -f ../examples/multitenancy/hello-world/hs1.yaml --kubeconfig=%s" % provider
        out, err = TestKubePlus.run_command(cmd)

        time.sleep(10)
        # test plugin
        cmd = "kubectl appstatus HelloWorldService hs1 -k %s" % provider
        out, err = TestKubePlus.run_command(cmd)
        
        if err != '':
            print("Something went wrong with the plugin.")
            print(err)
            cleanup()
            sys.exit(1)
        
        # asserts
        lines = out.split('\n')
        self.assertTrue('Deployed' in lines[1])
        self.assertTrue('Running' in lines[2] or 'Pending' in lines[2] or 'ContainerCreating' in lines[2])

        cleanup()
    # TODO: Add tests for
    # kubectl connections
    # kubectl appresources
    # kubectl appurl
    # kubectl applogs
    # kubectl metrics
    @unittest.skip("Skipping CLI test")
    def test_kubeplus_cli(self):
        kubeplus_home = os.getenv("KUBEPLUS_HOME")
        print("KubePlus home:" + kubeplus_home)
        path = os.getenv("PATH")
        print("Path:" + path)

        instance = ""
        kind = "wp"
        ns = "default"
        kubeplus_saas_provider = kubeplus_home + "/kubeplus-saas-provider.json"
        cmdsuffix = kind + " " + instance + " " + ns + " -k " + kubeplus_saas_provider
        cmd = "kubectl connections " + cmdsuffix

    @unittest.skip("Skipping Kyverno integration test")
    def test_kyverno_policies(self):
        if not TestKubePlus._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        if not TestKubePlus._is_kyverno_running():
            print("Kyverno is not running. Deploy Kyverno and then run this test.")
            sys.exit(0)

        cmd = "kubectl create -f block-stale-images.yaml"
        TestKubePlus.run_command(cmd)

        cmd = "kubectl create -f resource-quota/wordpress-service-composition.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = TestKubePlus.run_command(cmd)
        # print("Out:" + out)
        # print("Err:" + err)

        for line in err.split("\n"):
            if 'block-stale-images' in line.strip():
                self.assertTrue(True)
                cmd = "kubectl delete -f block-stale-images.yaml"
                TestKubePlus.run_command(cmd)
                return
        self.assertTrue(False)


if __name__ == '__main__':
    unittest.main()
