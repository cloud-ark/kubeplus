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

if __name__ == '__main__':
	crLogs = CRLogs()
	kind = sys.argv[1]
	instance = sys.argv[2]
	namespace = sys.argv[3]
	kubeconfig = sys.argv[4]
	resources = {}

	pods = crLogs.get_pods_in_ns(kind, instance, kubeconfig)
	for pod in pods:
		pod_name = pod['Name']
		pod_namespace = pod['Namespace']
		crLogs.get_logs(pod_name, pod_namespace, kubeconfig)
		print("---------------------------------------")
