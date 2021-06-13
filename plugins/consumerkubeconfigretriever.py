import sys
import json
import subprocess
import sys
import os
#import yaml


class ConsumerKubeconfigRetriever(object):

	def _get_kubeplus_ns(self):
		cmd = " kubectl get deployments -A "
		#print(cmd)
		out = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
		#print(out)
		out = out.decode('utf-8')
		kubeplusNamespace = ''
		for line in out.split("\n"):
			line1 = re.sub(' +', ' ', line)
			parts = line1.split()
			if parts[1] == 'kubeplus-deployment':
				kubeplusNamespace = parts[0]
				break
		return kubeplusNamespace

	def _apply_consumer_rbac(self, kubeplusNS, kindplural, providerkubeconfig):

		group = "platformapi.kubeplus"
		version = "v1alpha1"
		sa = 'kubeplus-saas-consumer'
		#ksnamespace = self._get_kubeplus_ns()

		role = {}
		role["apiVersion"] = "rbac.authorization.k8s.io/v1"
		role["kind"] = "ClusterRole"
		metadata = {}
		metadata["name"] = sa
		role["metadata"] = metadata

		ruleGroup1 = {}
		apiGroup1 = ["*"]
		resourceGroup1 = ["*"]
		verbsGroup1 = ["get","watch","list"]
		ruleGroup1["apiGroups"] = apiGroup1
		ruleGroup1["resources"] = resourceGroup1
		ruleGroup1["verbs"] = verbsGroup1

		ruleGroup2 = {}
		apiGroup2 = [group]
		resourceGroup2 = [kindplural]
		verbsGroup2 = ["get","watch","list","create","delete","update"]
		ruleGroup2["apiGroups"] = apiGroup2
		ruleGroup2["resources"] = resourceGroup2
		ruleGroup2["verbs"] = verbsGroup2

		ruleList = []
		ruleList.append(ruleGroup1)
		ruleList.append(ruleGroup2)
		role["rules"] = ruleList

		fp = open("./kubeplus-consumer-role.yaml", "w")
		role_json = json.dumps(role)
		#print(role_json)
		#print("-----\n")
		fp.write(role_json)
		fp.close()

		cmd = " kubectl create -f ./kubeplus-consumer-role.yaml --kubeconfig=" + providerkubeconfig

		cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
		#print(cmdOut)

		roleBinding = {}
		roleBinding["apiVersion"] = "rbac.authorization.k8s.io/v1"
		roleBinding["kind"] = "ClusterRoleBinding"
		metadata = {}
		metadata["name"] = sa
		roleBinding["metadata"] = metadata

		subject = {}
		subject["kind"] = "ServiceAccount"
		subject["name"] = sa
		subject["apiGroup"] = ""
		subject["namespace"] = kubeplusNS
		subjectList = []
		subjectList.append(subject)
		roleBinding["subjects"] = subjectList

		roleRef = {}
		roleRef["kind"] = "ClusterRole"
		roleRef["name"] = sa
		roleRef["apiGroup"] = "rbac.authorization.k8s.io"
		roleBinding["roleRef"] = roleRef

		fp = open("./kubeplus-consumer-rolebinding.yaml", "w")
		rolebinding_json = json.dumps(roleBinding)
		#print(rolebinding_json)
		fp.write(rolebinding_json)
		fp.close()

		cmd = " kubectl create -f ./kubeplus-consumer-rolebinding.yaml --kubeconfig=" + providerkubeconfig
		cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]	


if __name__ == '__main__':

	kubeplusNS = sys.argv[1]
	kindplural = sys.argv[2]
	providerkubeconfig = sys.argv[3]
	consumerKfgRetriever = ConsumerKubeconfigRetriever()
	consumerKfgRetriever._apply_consumer_rbac(kubeplusNS, kindplural, providerkubeconfig)

