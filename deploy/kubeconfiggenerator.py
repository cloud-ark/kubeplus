import sys
import json
import subprocess
import sys
import os
import yaml


class KubeconfigGenerator(object):

	def run_command(self, cmd):
		#print("Inside run_command")
		print(cmd)
		cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
		out = cmdOut[0]
		err = cmdOut[1]
		#print(out)
		if out != '':
			return out
			#printlines(out.decode('utf-8'))
		#print("Error:")
		#print(err)
		if err != '':
			return err
			#printlines(err.decode('utf-8'))

	def _create_kubecfg_file(self, sa, namespace, token, ca, server):
		top_level_dict = {}
		top_level_dict["apiVersion"] = "v1"
		top_level_dict["kind"] = "Config"

		contextName = sa

		usersList = []
		usertoken = {}
		usertoken["token"] = token
		userInfo = {}
		userInfo["name"] = sa
		userInfo["user"] = usertoken
		usersList.append(userInfo)
		top_level_dict["users"] = usersList

		clustersList = []
		cluster_details = {}
		cluster_details["server"] = server
		
		# TODO: Use the certificate authority to perform tls 
		# cluster_details["certificate-authority-data"] = ca
		cluster_details["insecure-skip-tls-verify"] = True

		clusterInfo = {}
		clusterInfo["cluster"] = cluster_details
		clusterInfo["name"] = sa
		clustersList.append(clusterInfo)
		top_level_dict["clusters"] = clustersList

		context_details = {}
		context_details["cluster"] = sa
		context_details["user"] = sa
		contextInfo = {}
		contextInfo["context"] = context_details
		contextInfo["name"] = contextName
		contextList = []
		contextList.append(contextInfo)
		top_level_dict["contexts"] = contextList

		top_level_dict["current-context"] = contextName

		json_file = json.dumps(top_level_dict)
		fileName =  sa + ".json"

		fp = open("/root/" + fileName, "w")
		fp.write(json_file)
		fp.close()

		configmapName = sa + "-kubeconfig"
		cmd = "kubectl create configmap " + configmapName + " -n " + namespace + " --from-file=/root/" + fileName
		self.run_command(cmd)

	def _create_role_rolebinding(self, contents, name):
		filePath = "/root/" + name
		fp = open(filePath, "w")
		#json_content = json.dumps(contents)
		#fp.write(json_content)
		yaml_content = yaml.dump(contents)
		fp.write(yaml_content)
		fp.close()
		print("---")
		print(yaml_content)
		print("---")
		cmd = " kubectl create -f " + filePath
		self.run_command(cmd)

	def _apply_provider_rbac(self, sa, namespace):
		role = {}
		role["apiVersion"] = "rbac.authorization.k8s.io/v1"
		role["kind"] = "ClusterRole"
		metadata = {}
		metadata["name"] = sa
		role["metadata"] = metadata

		# Read all resources
		ruleGroup1 = {}
		apiGroup1 = ["*",""]
		resourceGroup1 = ["*"]
		verbsGroup1 = ["get","watch","list"]
		ruleGroup1["apiGroups"] = apiGroup1
		ruleGroup1["resources"] = resourceGroup1
		ruleGroup1["verbs"] = verbsGroup1

		# CRUD on resourcecompositions et. al.
		ruleGroup2 = {}
		apiGroup2 = ["workflows.kubeplus"]
		resourceGroup2 = ["resourcecompositions","resourcemonitors","resourcepolicies","resourceevents"]
		verbsGroup2 = ["get","watch","list","create","delete","update"]
		ruleGroup2["apiGroups"] = apiGroup2
		ruleGroup2["resources"] = resourceGroup2
		ruleGroup2["verbs"] = verbsGroup2

		# CRUD on clusterroles and clusterrolebindings
		ruleGroup3 = {}
		apiGroup3 = ["rbac.authorization.k8s.io"]
		resourceGroup3 = ["clusterroles","clusterrolebindings"]
		verbsGroup3 = ["get","watch","list","create","delete","update"]
		ruleGroup3["apiGroups"] = apiGroup3
		ruleGroup3["resources"] = resourceGroup3
		ruleGroup3["verbs"] = verbsGroup3

		# CRUD on platformapi.kubeplus
		ruleGroup5 = {}
		apiGroup5 = ["platformapi.kubeplus"]
		resourceGroup5 = ["*"]
		verbsGroup5 = ["get","watch","list","create","delete","update"]
		ruleGroup5["apiGroups"] = apiGroup5
		ruleGroup5["resources"] = resourceGroup5
		ruleGroup5["verbs"] = verbsGroup5

		# CRUD on networkpolicies
		ruleGroup6 = {}
		apiGroup6 = ["networking.k8s.io"]
		resourceGroup6 = ["networkpolicies"]
		verbsGroup6 = ["get","watch","list","create","delete","update"]
		ruleGroup6["apiGroups"] = apiGroup6
		ruleGroup6["resources"] = resourceGroup6
		ruleGroup6["verbs"] = verbsGroup6

		# CRUD on namespaces
		ruleGroup7 = {}
		apiGroup7 = [""]
		resourceGroup7 = ["namespaces"]
		verbsGroup7 = ["get","watch","list","create","delete","update"]
		ruleGroup7["apiGroups"] = apiGroup7
		ruleGroup7["resources"] = resourceGroup7
		ruleGroup7["verbs"] = verbsGroup7

		# CRUD on HPA
		ruleGroup8 = {}
		apiGroup8 = ["autoscaling"]
		resourceGroup8 = ["horizontalpodautoscalers"]
		verbsGroup8 = ["get","watch","list","create","delete","update"]
		ruleGroup8["apiGroups"] = apiGroup8
		ruleGroup8["resources"] = resourceGroup8
		ruleGroup8["verbs"] = verbsGroup8

		ruleList = []
		ruleList.append(ruleGroup1)
		ruleList.append(ruleGroup2)
		ruleList.append(ruleGroup3)
		ruleList.append(ruleGroup5)
		ruleList.append(ruleGroup6)
		ruleList.append(ruleGroup7)
		ruleList.append(ruleGroup8)
		role["rules"] = ruleList

		roleName = sa + "-role.yaml"
		self._create_role_rolebinding(role, roleName)

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
		subject["namespace"] = namespace
		subjectList = []
		subjectList.append(subject)
		roleBinding["subjects"] = subjectList

		roleRef = {}
		roleRef["kind"] = "ClusterRole"
		roleRef["name"] = sa
		roleRef["apiGroup"] = "rbac.authorization.k8s.io"
		roleBinding["roleRef"] = roleRef

		roleBindingName = sa + "-rolebinding.yaml"
		self._create_role_rolebinding(roleBinding, roleBindingName)

	def _apply_rbac(self, sa, namespace, entity=''):
		if entity == 'provider':
			self._apply_provider_rbac(sa, namespace)
		if entity == 'consumer':
			self._apply_consumer_rbac(sa, namespace)

	def _generate_kubeconfig(self, sa, namespace):
		cmdprefix = ""
		cmd = " kubectl create sa " + sa + " -n " + namespace
		cmdToRun = cmdprefix + " " + cmd
		self.run_command(cmdToRun)

		cmd = " kubectl get sa " + sa + " -n " + namespace + " -o json "
		cmdToRun = cmdprefix + " " + cmd
		out = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
		out = out.decode('utf-8')
		if out:
			json_output = json.loads(out)
			secretName = json_output["secrets"][0]["name"]
			#print("Secret Name:" + secretName)

			cmd1 = " kubectl describe secret " + secretName + " -n " + namespace
			cmdToRun = cmdprefix + " " + cmd1
			out1 = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
			out1 = out1.decode('utf-8')
			token = ''
			for line in out1.split("\n"):
				if 'token' in line:
					parts = line.split(":")
					token = parts[1].strip()

			cmd1 = " kubectl get secret " + secretName + " -n " + namespace + " -o json "
			cmdToRun = cmdprefix + " " + cmd1
			out1 = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
			out1 = out1.decode('utf-8')
			json_output1 = json.loads(out1)
			ca_cert = json_output1["data"]["ca.crt"].strip()
			#print("CA Cert:" + ca_cert)

			#cmd2 = " kubectl config view --minify -o json "
			cmd2 = "kubectl -n default get endpoints kubernetes | awk '{print $2}' | grep -v ENDPOINTS"
			cmdToRun = cmdprefix + " " + cmd2
			out2 = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
			#print("Config view Minify:")
			print(out2)
			out2 = out2.decode('utf-8')
			#json_output2 = json.loads(out2)
			#server = json_output2["clusters"][0]["cluster"]["server"].strip()
			server = out2.strip()
			server = "https://" + server
			print("Kube API Server:" + server)
			self._create_kubecfg_file(sa, namespace, token, ca_cert, server)


if __name__ == '__main__':
	kubeconfigGenerator = KubeconfigGenerator()
	namespace = sys.argv[1]

	# 1. Generate Provider kubeconfig
	sa = 'kubeplus-saas-provider'
	kubeconfigGenerator._generate_kubeconfig(sa, namespace)
	kubeconfigGenerator._apply_rbac(sa, namespace, entity='provider')

	# 2. Generate Consumer kubeconfig
	sa = 'kubeplus-saas-consumer'
	kubeconfigGenerator._generate_kubeconfig(sa, namespace)
#	kubeconfigGenerator._apply_rbac(sa, namespace, entity='consumer')
