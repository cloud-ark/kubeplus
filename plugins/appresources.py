import subprocess
import sys
import json
import yaml
import platform
import os
from crmetrics import CRBase

class AppResourcesFinder(CRBase):

    def _run_command(self, cmd):
        cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
        out = cmdOut[0].decode('utf-8')
        err = cmdOut[1].decode('utf-8')
        return out, err
   
    def _get_resources(self, kind, plural, targetNS, kubeconfig):
        cmd = "kubectl get " + plural + " -n " + targetNS + " " + kubeconfig
        out, err = self._run_command(cmd)
        resources = []
        for line in out.split("\n"):
            res_details = {}
            line = line.strip()
            if 'NAME' not in line and line != '' and line != '\n':
                line1 = ' '.join(line.split())
                parts = line1.split(" ")
                res_name = parts[0].strip()
                res_details['name'] = res_name
                res_details['namespace'] = targetNS
                res_details['kind'] = kind 
                resources.append(res_details)
        return resources

    def get_kubeplus_ns(self, kubeconfig):
        cmd = 'kubectl get deployments -A ' + kubeconfig
        out, err = self._run_command(cmd)
        for line in out.split("\n"):
            if 'NAME' not in line:
                if 'kubeplus-deployment' in line:
                    line1 = ' '.join(line.split())
                    parts = line1.split(" ")
                    kubeplus_ns = parts[0].strip()
                    return kubeplus_ns

    def get_target_ns(self, kubeplus_ns, kind, instance, kubeconfig):
        cmd = 'kubectl get ' + kind + ' ' + instance + " -n " + kubeplus_ns + ' -o json ' + kubeconfig 
        out, err = self._run_command(cmd)
        targetNS = ''
        if out != '':
            json_obj = json.loads(out)
            helmrelease = json_obj['status']['helmrelease'].strip()
            parts = helmrelease.split(":")
            targetNS = parts[0].strip()
            return targetNS

    def get_networkpolicies(self, targetNS, kind, instance, kubeconfig):
        resources = self._get_resources('NetworkPolicy', 'networkpolicies', targetNS, kubeconfig)
        return resources

    def get_resourcequotas(self, targetNS, kind, instance, kubeconfig):
        resources = self._get_resources('ResourceQuota', 'resourcequotas', targetNS, kubeconfig)
        return resources

    def get_pods(self, targetNS, kind, instance, kubeconfig):
        resources = self._get_resources('Pod', 'pods', targetNS, kubeconfig)
        return resources

if __name__ == '__main__':
    appResourcesFinder = AppResourcesFinder()
    kind = sys.argv[1]
    instance = sys.argv[2]
    kubeconfig = sys.argv[3]

    #print("kind:" + kind + " instance:" + instance + " kubeconfig:" + kubeconfig)

    kubeplus_ns = appResourcesFinder.get_kubeplus_ns(kubeconfig)    
    targetNS = appResourcesFinder.get_target_ns(kubeplus_ns, kind, instance, kubeconfig)
    pods = appResourcesFinder.get_pods(targetNS, kind, instance, kubeconfig)
    networkpolicies = appResourcesFinder.get_networkpolicies(targetNS, kind, instance, kubeconfig)
    resourcequotas = appResourcesFinder.get_resourcequotas(targetNS, kind, instance, kubeconfig)

    allresources = []
    allresources.extend(pods)
    allresources.extend(networkpolicies)
    allresources.extend(resourcequotas)

    print("NAMESPACE\tKIND\t\t\tNAME")

    line = kubeplus_ns + '\t\t' + kind + '\t' + instance
    print(line)
    for res in allresources:
        tabs = '\t\t'
        if res['kind'] == 'Pod':
            tabs = '\t\t\t'
        line = res['namespace'] + '\t' + res['kind'] + tabs + res['name']
        print(line)
