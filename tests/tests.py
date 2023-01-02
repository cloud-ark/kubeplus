import unittest
import requests
import sys
import json
import subprocess
import sys
import os
import yaml
import time

class TestKubePlus(unittest.TestCase):

    def _run_command(self, cmd):
        print(cmd)
        cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
        out = cmdOut[0].decode('utf-8')
        err = cmdOut[1].decode('utf-8')
        print(out)
        print("---")
        print(err)
        return out, err

    def _is_kubeplus_running(self):
        cmd = 'kubectl get pods -A'
        out, err = self._run_command(cmd)
        for line in out.split("\n"):
            if 'kubeplus' in line and 'Running' in line:
                return True
        return False

    def test_create_res_comp_for_chart_with_ns(self):
        if not self._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        cmd = "kubectl create -f wordpress-service-composition-chart-withns.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = self._run_command(cmd)
        print("Out:" + out)
        print("Err:" + err)
        self.assertTrue('Namespace object is not allowed in the chart' in err)

    def test_create_res_comp_for_chart_with_shared_storage(self):
        if not self._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        cmd = "kubectl create -f storage-class-fast.yaml"
        self._run_command(cmd)

        cmd = "kubectl create -f storage-isolation/wordpress-service-composition.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = self._run_command(cmd)
        print("Out:" + out)
        print("Err:" + err)
        self.assertTrue('Storage class with reclaim policy Retain not allowed' in err)

    def test_create_res_comp_with_incomplete_resource_quota(self):
        if not self._is_kubeplus_running():
            print("KubePlus is not running. Deploy KubePlus and then run tests")
            sys.exit(0)

        cmd = "kubectl create -f resource-quota/wordpress-service-composition-1.yaml --kubeconfig=../kubeplus-saas-provider.json"
        out, err = self._run_command(cmd)
        print("Out:" + out)
        print("Err:" + err)
        self.assertTrue('If quota is specified, specify all four values: requests.cpu, requests.memory, limits.cpu, limits.memory' in err)


if __name__ == '__main__':
    #unittest.main()
    testKubePlus = TestKubePlus()
    testKubePlus.test_create_res_comp_for_chart_with_ns()
    testKubePlus.test_create_res_comp_for_chart_with_shared_storage()
    testKubePlus.test_create_res_comp_with_incomplete_resource_quota()
