import subprocess
import sys
import json
import yaml
import platform
import os
from crmetrics import CRBase


class AppResourcesFinder(CRBase):

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
        releaseName = ''
        out = out.strip()
        if out != '' and out != None:
            json_obj = json.loads(out)
            if 'status' in json_obj and 'helmrelease' in json_obj['status']:
                helmrelease = json_obj['status']['helmrelease'].strip()
                parts = helmrelease.split(":")
                targetNS = parts[0].strip()
                releaseName = parts[1].strip().split("\n")[0]
                return targetNS, releaseName
        return targetNS, releaseName

    def get_helm_resources(self, targetNS, helmrelease, kubeconfig):
        # print("Inside helm_resources")
        cmd = "helm get all " + helmrelease + " -n " + targetNS + ' ' + kubeconfig
        out, err = self._run_command(cmd)

        resources = []
        kind = ''
        res_name = ''
        new_resource = False
        for line in out.split("\n"):
            if new_resource:
                if 'name' in line:
                    res_name = (line.split(":")[1]).strip()
                    res_details = {}
                    res_details['name'] = res_name
                    res_details['namespace'] = targetNS
                    res_details['kind'] = kind
                    resources.append(res_details)

                    new_resource = False
            if 'kind' in line:
                kind = (line.split(":")[1]).strip()
                new_resource = True

        return resources

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

    # print("kind:" + kind + " instance:" + instance + " kubeconfig:" + kubeconfig)

    valid_consumer_api = appResourcesFinder.verify_kind_is_consumerapi(kind, kubeconfig)
    if not valid_consumer_api:
        print(("{} is not a valid Consumer API.").format(kind))
        exit(0)

    res_exists, ns, err = appResourcesFinder.check_res_exists(kind, instance, kubeconfig)
    if not res_exists:
        print(err)
        exit(0)

    kubeplus_ns = appResourcesFinder.get_kubeplus_ns(kubeconfig)
    res_ns = kubeplus_ns
    if ns != res_ns and ns != '':
        res_ns = ns
    targetNS, helmrelease = appResourcesFinder.get_target_ns(res_ns, kind, instance, kubeconfig)
    if targetNS == '' and helmrelease == '':
        print("No Helm release found for {} resource {}".format(kind, instance))
    # print(targetNS + " " + helmrelease)
    pods = appResourcesFinder.get_pods(targetNS, kind, instance, kubeconfig)
    networkpolicies = appResourcesFinder.get_networkpolicies(targetNS, kind, instance, kubeconfig)
    resourcequotas = appResourcesFinder.get_resourcequotas(targetNS, kind, instance, kubeconfig)
    helmresources = appResourcesFinder.get_helm_resources(targetNS, helmrelease, kubeconfig)

    allresources = []
    allresources.extend(helmresources)
    allresources.extend(pods)
    allresources.extend(networkpolicies)
    allresources.extend(resourcequotas)

    # Ref: https://www.educba.com/python-print-table/
    # https://stackoverflow.com/questions/20309255/how-to-pad-a-string-to-a-fixed-length-with-spaces-in-python
    print("{:<25} {:<25} {:<25} ".format("NAMESPACE", "KIND", "NAME"))
    print("{:<25} {:<25} {:<25} ".format(kubeplus_ns, kind, instance))
    for res in allresources:
        ns = res['namespace']
        kind = res['kind']
        name = res['name']
        print("{:<25} {:<25} {:<25} ".format(ns, kind, name))
