import sys
import json
import subprocess
import sys
import os
import yaml
import time

from logging.config import dictConfig

dictConfig({
    'version': 1,
    'formatters': {'default': {
        'format': '[%(asctime)s] %(levelname)s in %(module)s: %(message)s',
    }},
    'handlers': {'wsgi': {
        'class': 'logging.StreamHandler',
        'stream': 'ext://flask.logging.wsgi_errors_stream',
        'formatter': 'default'
    },
     'file.handler': {
            'class': 'logging.handlers.RotatingFileHandler',
            'filename': 'provider-kubeconfig.log',
            'maxBytes': 10000000,
            'backupCount': 5,
            'level': 'DEBUG',
        },
    },
    'root': {
        'level': 'INFO',
        'handlers': ['file.handler']
    }
})



def create_role_rolebinding(contents, name):
    filePath = os.getcwd() + "/" + name
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
    run_command(cmd)


def run_command(cmd):
    #print("Inside run_command")
    print(cmd)
    cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
    out = cmdOut[0].decode('utf-8')
    err = cmdOut[1].decode('utf-8')
    print(out)
    print("---")
    print(err)
    return out, err


class KubeconfigGenerator(object):

	def run_command(self, cmd):
		#print("Inside run_command")
		print(cmd)
		cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
		out = cmdOut[0]
		err = cmdOut[1]
		print(out)
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
		context_details["namespace"] = namespace
		contextInfo = {}
		contextInfo["context"] = context_details
		contextInfo["name"] = contextName
		contextList = []
		contextList.append(contextInfo)
		top_level_dict["contexts"] = contextList

		top_level_dict["current-context"] = contextName

		json_file = json.dumps(top_level_dict)
		fileName =  sa + ".json"

		fp = open(os.getcwd() + "/" + fileName, "w")
		fp.write(json_file)
		fp.close()

		configmapName = sa
		created = False 
		while not created:        
			cmd = "kubectl create configmap " + configmapName + " -n " + namespace + " --from-file=" + os.getcwd() + "/" + fileName
			self.run_command(cmd)
			get_cmd = "kubectl get configmap " + configmapName + " -n "  + namespace
			output = self.run_command(cmd)
			output = output.decode('utf-8')    
			if 'Error from server (NotFound)' in output:
				time.sleep(2)
				print("Trying again..")
			else:
				created = True


	def _apply_consumer_rbac(self, sa, namespace):
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

		# Impersonate users, groups, serviceaccounts
		ruleGroup9 = {}
		apiGroup9 = [""]
		resourceGroup9 = ["users","groups","serviceaccounts"]
		verbsGroup9 = ["impersonate"]
		ruleGroup9["apiGroups"] = apiGroup9
		ruleGroup9["resources"] = resourceGroup9
		ruleGroup9["verbs"] = verbsGroup9

		# Pod/portforward to open consumerui
		ruleGroup10 = {}
		apiGroup10 = [""]
		resourceGroup10 = ["pods/portforward"]
		verbsGroup10 = ["create","get"]
		ruleGroup10["apiGroups"] = apiGroup10
		ruleGroup10["resources"] = resourceGroup10
		ruleGroup10["verbs"] = verbsGroup10

		ruleList = []
		ruleList.append(ruleGroup1)
		ruleList.append(ruleGroup9)
		ruleList.append(ruleGroup10)
		role["rules"] = ruleList

		roleName = sa + "-role-impersonate.yaml"
		create_role_rolebinding(role, roleName)

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

		roleBindingName = sa + "-rolebinding-impersonate.yaml"
		create_role_rolebinding(roleBinding, roleBindingName)

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
		verbsGroup2 = ["get","watch","list","create","delete","update","patch"]
		ruleGroup2["apiGroups"] = apiGroup2
		ruleGroup2["resources"] = resourceGroup2
		ruleGroup2["verbs"] = verbsGroup2

		# CRUD on clusterroles and clusterrolebindings
		ruleGroup3 = {}
		apiGroup3 = ["rbac.authorization.k8s.io"]
		resourceGroup3 = ["clusterroles","clusterrolebindings","roles","rolebindings"]
		verbsGroup3 = ["get","watch","list","create","delete","update","patch","deletecollection"]
		ruleGroup3["apiGroups"] = apiGroup3
		ruleGroup3["resources"] = resourceGroup3
		ruleGroup3["verbs"] = verbsGroup3

		# CRUD on Port forward
		ruleGroup4 = {}
		apiGroup4 = [""]
		resourceGroup4 = ["pods/portforward"]
		verbsGroup4 = ["get","watch","list","create","delete","update","patch"]
		ruleGroup4["apiGroups"] = apiGroup4
		ruleGroup4["resources"] = resourceGroup4
		ruleGroup4["verbs"] = verbsGroup4

		# CRUD on platformapi.kubeplus
		ruleGroup5 = {}
		apiGroup5 = ["platformapi.kubeplus"]
		resourceGroup5 = ["*"]
		verbsGroup5 = ["get","watch","list","create","delete","update","patch"]
		ruleGroup5["apiGroups"] = apiGroup5
		ruleGroup5["resources"] = resourceGroup5
		ruleGroup5["verbs"] = verbsGroup5

		# CRUD on secrets, serviceaccounts, configmaps
		ruleGroup6 = {}
		apiGroup6 = [""]
		resourceGroup6 = ["secrets", "serviceaccounts", "configmaps","events","persistentvolumeclaims","serviceaccounts/token","services","services/proxy","endpoints"]
		verbsGroup6 = ["get","watch","list","create","delete","update","patch", "deletecollection"]
		ruleGroup6["apiGroups"] = apiGroup6
		ruleGroup6["resources"] = resourceGroup6
		ruleGroup6["verbs"] = verbsGroup6

		# CRUD on namespaces
		ruleGroup7 = {}
		apiGroup7 = [""]
		resourceGroup7 = ["namespaces"]
		verbsGroup7 = ["get","watch","list","create","delete","update","patch"]
		ruleGroup7["apiGroups"] = apiGroup7
		ruleGroup7["resources"] = resourceGroup7
		ruleGroup7["verbs"] = verbsGroup7

		# CRUD on Deployments
		ruleGroup8 = {}
		apiGroup8 = ["apps"]
		resourceGroup8 = ["deployments","daemonsets","deployments/rollback","deployments/scale","replicasets","replicasets/scale","statefulsets","statefulsets/scale"]
		verbsGroup8 = ["get","watch","list","create","delete","update","patch","deletecollection"]
		ruleGroup8["apiGroups"] = apiGroup8
		ruleGroup8["resources"] = resourceGroup8
		ruleGroup8["verbs"] = verbsGroup8

		# Impersonate users, groups, serviceaccounts
		ruleGroup9 = {}
		apiGroup9 = [""]
		resourceGroup9 = ["users","groups","serviceaccounts"]
		verbsGroup9 = ["impersonate"]
		ruleGroup9["apiGroups"] = apiGroup9
		ruleGroup9["resources"] = resourceGroup9
		ruleGroup9["verbs"] = verbsGroup9

		# Exec into the Pods and others in the "" apiGroup
		ruleGroup10 = {}
		apiGroup10 = [""]
		resourceGroup10 = ["pods","pods/attach","pods/exec","pods/portforward","pods/proxy","pods/eviction","replicationcontrollers","replicationcontrollers/scale"]
		verbsGroup10 = ["get","list","create","update","delete","watch","patch","deletecollection"]
		ruleGroup10["apiGroups"] = apiGroup10
		ruleGroup10["resources"] = resourceGroup10
		ruleGroup10["verbs"] = verbsGroup10

		# AdmissionRegistration
		ruleGroup11 = {}
		apiGroup11 = ["admissionregistration.k8s.io"]
		resourceGroup11 = ["mutatingwebhookconfigurations"]
		verbsGroup11 = ["get","create","delete","update"]
		ruleGroup11["apiGroups"] = apiGroup11
		ruleGroup11["resources"] = resourceGroup11
		ruleGroup11["verbs"] = verbsGroup11

		# APIExtension
		ruleGroup12 = {}
		apiGroup12 = ["apiextensions.k8s.io"]
		resourceGroup12 = ["customresourcedefinitions"]
		verbsGroup12 = ["get","create","delete","update", "patch"]
		ruleGroup12["apiGroups"] = apiGroup12
		ruleGroup12["resources"] = resourceGroup12
		ruleGroup12["verbs"] = verbsGroup12

		# Certificates
		ruleGroup13 = {}
		apiGroup13 = ["certificates.k8s.io"]
		resourceGroup13 = ["signers"]
		resourceNames13 = ["kubernetes.io/legacy-unknown","kubernetes.io/kubelet-serving","kubernetes.io/kube-apiserver-client","cloudark.io/kubeplus"]
		verbsGroup13 = ["get","create","delete","update", "patch", "approve"]
		ruleGroup13["apiGroups"] = apiGroup13
		ruleGroup13["resources"] = resourceGroup13
		ruleGroup13["resourceNames"] = resourceNames13
		ruleGroup13["verbs"] = verbsGroup13

		# Read all
		ruleGroup14 = {}
		apiGroup14 = ["*"]
		resourceGroup14 = ["*"]
		verbsGroup14 = ["get"]
		ruleGroup14["apiGroups"] = apiGroup14
		ruleGroup14["resources"] = resourceGroup14
		ruleGroup14["verbs"] = verbsGroup14

		ruleGroup15 = {}
		apiGroup15 = ["certificates.k8s.io"]
		resourceGroup15 = ["certificatesigningrequests", "certificatesigningrequests/approval"]
		verbsGroup15 = ["create","delete","update", "patch"]
		ruleGroup15["apiGroups"] = apiGroup15
		ruleGroup15["resources"] = resourceGroup15
		ruleGroup15["verbs"] = verbsGroup15

		ruleGroup16 = {}
		apiGroup16 = ["extensions"]
		resourceGroup16 = ["deployments","daemonsets","deployments/rollback","deployments/scale","replicasets","replicasets/scale","replicationcontrollers/scale","ingresses","networkpolicies"]
		verbsGroup16 = ["get","watch","list","create","delete","update","patch","deletecollection"]
		ruleGroup16["apiGroups"] = apiGroup16
		ruleGroup16["resources"] = resourceGroup16
		ruleGroup16["verbs"] = verbsGroup16

		ruleGroup17 = {}
		apiGroup17 = ["networking.k8s.io"]
		resourceGroup17 = ["ingresses","networkpolicies"]
		verbsGroup17 = ["get","watch","list","create","delete","update","patch","deletecollection"]
		ruleGroup17["apiGroups"] = apiGroup17
		ruleGroup17["resources"] = resourceGroup17
		ruleGroup17["verbs"] = verbsGroup17

		ruleGroup18 = {}
		apiGroup18 = ["authorization.k8s.io"]
		resourceGroup18 = ["localsubjectaccessreviews"]
		verbsGroup18 = ["create"]
		ruleGroup18["apiGroups"] = apiGroup18
		ruleGroup18["resources"] = resourceGroup18
		ruleGroup18["verbs"] = verbsGroup18

		ruleGroup19 = {}
		apiGroup19 = ["autoscaling"]
		resourceGroup19 = ["horizontalpodautoscalers"]
		verbsGroup19 = ["create", "delete", "deletecollection", "patch", "update"]
		ruleGroup19["apiGroups"] = apiGroup19
		ruleGroup19["resources"] = resourceGroup19
		ruleGroup19["verbs"] = verbsGroup19

		ruleGroup20 = {}
		apiGroup20 = ["batch"]
		resourceGroup20 = ["cronjobs","jobs"]
		verbsGroup20 = ["create", "delete", "deletecollection", "patch", "update"]
		ruleGroup20["apiGroups"] = apiGroup20
		ruleGroup20["resources"] = resourceGroup20
		ruleGroup20["verbs"] = verbsGroup20

		ruleGroup21 = {}
		apiGroup21 = ["policy"]
		resourceGroup21 = ["poddisruptionbudgets"]
		verbsGroup21 = ["create", "delete", "deletecollection", "patch", "update"]
		ruleGroup21["apiGroups"] = apiGroup21
		ruleGroup21["resources"] = resourceGroup21
		ruleGroup21["verbs"] = verbsGroup21

		ruleGroup22 = {}
		apiGroup22 = [""]
		resourceGroup22 = ["resourcequotas"]
		verbsGroup22 = ["create", "delete", "deletecollection", "patch", "update"]
		ruleGroup22["apiGroups"] = apiGroup22
		ruleGroup22["resources"] = resourceGroup22
		ruleGroup22["verbs"] = verbsGroup22

		ruleList = []
		ruleList.append(ruleGroup1)
		ruleList.append(ruleGroup2)
		ruleList.append(ruleGroup3)
		ruleList.append(ruleGroup4)
		ruleList.append(ruleGroup5)
		ruleList.append(ruleGroup6)
		ruleList.append(ruleGroup7)
		ruleList.append(ruleGroup8)
		ruleList.append(ruleGroup9)
		ruleList.append(ruleGroup10)
		ruleList.append(ruleGroup11)
		ruleList.append(ruleGroup12)
		ruleList.append(ruleGroup13)
		ruleList.append(ruleGroup14)
		ruleList.append(ruleGroup15)
		ruleList.append(ruleGroup16)
		ruleList.append(ruleGroup17)
		ruleList.append(ruleGroup18)
		ruleList.append(ruleGroup19)
		ruleList.append(ruleGroup20)
		ruleList.append(ruleGroup21)
		ruleList.append(ruleGroup22)

		role["rules"] = ruleList

		roleName = sa + "-role.yaml"
		create_role_rolebinding(role, roleName)

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
		create_role_rolebinding(roleBinding, roleBindingName)

	def _apply_rbac(self, sa, namespace, entity=''):
		if entity == 'provider':
			self._apply_provider_rbac(sa, namespace)
		if entity == 'consumer':
			self._apply_consumer_rbac(sa, namespace)

	def _create_secret(self, sa, namespace):

		annotations = {}
		annotations['kubernetes.io/service-account.name'] = sa

		metadata = {}
		metadata['name'] = sa
		metadata['namespace'] = namespace
		metadata['annotations'] = annotations

		secret = {}
		secret['apiVersion'] = "v1"
		secret['kind'] = "Secret"
		secret['metadata'] = metadata
		secret['type'] = 'kubernetes.io/service-account-token'

		secretName = sa + "-secret.yaml"

		filePath = os.getcwd() + "/" + secretName
		fp = open(filePath, "w")
		yaml_content = yaml.dump(secret)
		fp.write(yaml_content)
		fp.close()
		print("---")
		print(yaml_content)
		print("---")
		created = False
		while not created:
			cmd = " kubectl create -f " + filePath
			out = self.run_command(cmd)
			if out != '':
				out = out.decode('utf-8').strip()
				print(out)
				if 'created' in out:
					created = True
			else:
				time.sleep(2)
		print("Create secret:" + out)
		return out

	def _generate_kubeconfig(self, sa, namespace):
		cmdprefix = ""
		cmd = " kubectl create sa " + sa + " -n " + namespace
		cmdToRun = cmdprefix + " " + cmd
		self.run_command(cmdToRun)

		#cmd = " kubectl get sa " + sa + " -n " + namespace + " -o json "
		#cmdToRun = cmdprefix + " " + cmd
		#out = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]

		secretName = sa
		out = self._create_secret(secretName, namespace)
		print("Create secret:" + out)
		if 'secret/' + sa + ' created' in out:
			#json_output = json.loads(out)
			#secretName = json_output["secrets"][0]["name"]
			#print("Secret Name:" + secretName)

			tokenFound = False
			while not tokenFound:
				cmd1 = " kubectl describe secret " + secretName + " -n " + namespace
				cmdToRun = cmdprefix + " " + cmd1
				out1 = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
				out1 = out1.decode('utf-8')
				print(out1)
				token = ''
				for line in out1.split("\n"):
					if 'token' in line:
						parts = line.split(":")
						token = parts[1].strip()
				if token != '':
					tokenFound = True
				else:
					time.sleep(2)

			print("Got secret token")
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

	if len(sys.argv) < 3:
		print("python provider-kubeconfig.py <create|delete> <namespace>")
		exit()

	action = sys.argv[1]
	namespace = sys.argv[2]

	kubeconfigGenerator = KubeconfigGenerator()
	sa = 'kubeplus-saas-provider'
	if action == "create":
		out, err = run_command("kubectl get ns namespace")
		if 'not found' in out or 'not found' in err:
			run_command("kubectl create ns " + namespace)

		cmd = "kubectl label --overwrite=true ns " + namespace + " managedby=kubeplus"
		run_command(cmd)

		# 1. Generate Provider kubeconfig
		kubeconfigGenerator._generate_kubeconfig(sa, namespace)
		kubeconfigGenerator._apply_rbac(sa, namespace, entity='provider')

	if action == "delete":
		run_command("kubectl delete sa " + sa + " -n " + namespace)
		run_command("kubectl delete configmap " + sa + " -n " + namespace)
		run_command("kubectl delete clusterrole " + sa + " -n " + namespace)
		run_command("kubectl delete clusterrolebinding " + sa + " -n " + namespace)
		cwd = os.getcwd()
		run_command("rm " + cwd + "/kubeplus-saas-provider-secret.yaml")
		run_command("rm " + cwd + "/kubeplus-saas-provider.json")
		run_command("rm " + cwd + "/kubeplus-saas-provider-role.yaml")
		run_command("rm " + cwd + "/kubeplus-saas-provider-rolebinding.yaml")	

