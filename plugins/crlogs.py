import subprocess
import sys
import json
import platform
import os
from crmetrics import CRBase

class CRLogs(CRBase):

	def _get_container_logs(self, pod, namespace, containers, kubeconfig):
		for c in containers:
			container = c['name']
			cmd = 'kubectl logs ' + pod + ' -n ' + namespace + ' -c ' + container + ' ' + kubeconfig
			#print(cmd)

			print("======== Pod::" + pod + "/container::" + container + " ===========")
			try:
				out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										stderr=subprocess.PIPE, shell=True).communicate()[0]
				if out:
					print(out)
					print("================================================\n\n")
			except Exception as e:
				print(e)

	def get_logs(self, pod, namespace, kubeconfig):
		cmd = 'kubectl get pods ' + pod + ' -n ' + namespace + ' -o json ' + kubeconfig
		#print(cmd)
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]

			if out:
				json_output = json.loads(out)
				containers = json_output['spec']['containers']
				self._get_container_logs(pod, namespace, containers, kubeconfig)
			
				if 'initContainers' in json_output['spec']:
					init_containers = json_output['spec']['initContainers']
					self._get_container_logs(pod, namespace, init_containers, kubeconfig)

		except Exception as e:
			print(e)

	def get_resources_composition(self, kind, instance, namespace, kubeconfig):
		platf = platform.system()
		kubeplus_home = os.getenv('KUBEPLUS_HOME')
		cmd = ''
		json_output = {}
		if platf == "Darwin":
			cmd = kubeplus_home + '/plugins/kubediscovery-macos composition ' 
		elif platf == "Linux":
			cmd = kubeplus_home + '/plugins/kubediscovery-linux composition '
		else:
			print("OS not supported:" + platf)
			return json_output
		cmd = cmd + kind + ' ' + instance + ' ' + namespace + ' ' + kubeconfig
		#print(cmd)
		out = ''
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
								   stderr=subprocess.PIPE, shell=True).communicate()[0]
			out = out.decode('utf-8')
		except Exception as e:
			print(e)
		if out:
			print(out)
			try:
				json_output = json.loads(out)
			except Exception as e:
				print(e)
		return json_output

	def get_pods1(self, resources):
		pod_list = []
		for resource in resources:
			#print(resource)
			if resource['Kind'] == 'Pod':
				present = False
				for p in pod_list:
					if p['Name'] == resource['Name']:
						present = True
						break
				if not present:
					pod_list.append(resource)
		#print(pod_list)
		return pod_list

if __name__ == '__main__':
	crLogs = CRLogs()
	#crLogs.get_logs(sys.argv[1], sys.argv[2])
	#resources = sys.argv[1]
	relation = sys.argv[1]
	kind = sys.argv[2]
	instance = sys.argv[3]
	namespace = sys.argv[4]
	kubeconfig = sys.argv[5]
	#print(kind + " " + instance + " " + namespace + " " + kubeconfig)
	resources = {}
	#if relation == 'connections':
	#	resources = crLogs.get_resources_connections(kind, instance, namespace, kubeconfig)
	#	#print(resources)
	#if relation == 'composition':
	#	resources = crLogs.get_resources_composition(kind, instance, namespace, kubeconfig)
	#	#print(resources)
	#resource_json = json.loads(resources)
	pods = crLogs.get_pods_in_ns(kind, instance, kubeconfig)
	for pod in pods:
		pod_name = pod['Name']
		pod_namespace = pod['Namespace']
		#print(pod_name)
		crLogs.get_logs(pod_name, pod_namespace, kubeconfig)
		print("---------------------------------------")
