import subprocess
import sys
import json
import platform
import requests
import os
from crmetrics import CRBase

class CRLogs(CRBase):

	def _get_container_logs(self, pod, namespace, containers, kubeconfig):
		container_logs = []
		for c in containers:
			container = c['name']
			cmd = 'kubectl logs ' + pod + ' -n ' + namespace + ' -c ' + container + ' ' + kubeconfig
			container_logs.append("======== Pod::" + pod + "/container::" + container + " ===========")
			try:
				out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										stderr=subprocess.PIPE, shell=True).communicate()[0]
				if out:
					container_logs.append(str(out))
					container_logs.append("================================================\n\n")
			except Exception as e:
				container_logs.append(str(e))
				
		return "\n".join(container_logs)

	def get_logs(self, pod, namespace, kubeconfig):
		cmd = 'kubectl get pods ' + pod + ' -n ' + namespace + ' -o json ' + kubeconfig
		joined_logs = []
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]

			if out:
				json_output = json.loads(out)
				containers = json_output['spec']['containers']
				joined_logs.append(self._get_container_logs(pod, namespace, containers, kubeconfig))
			
				if 'initContainers' in json_output['spec']:
					init_containers = json_output['spec']['initContainers']
					joined_logs.append(self._get_container_logs(pod, namespace, init_containers, kubeconfig))

		except Exception as e:
			joined_logs.append(str(e))
			
		return "\n".join(joined_logs)

if __name__ == '__main__':
	crLogs = CRLogs()
	kind = sys.argv[1]
	instance = sys.argv[2]
	kubeconfig = sys.argv[3]
	resources = {}
	
	joined_logs = []
	pods = crLogs.get_pods_in_ns(kind, instance, kubeconfig)
	for pod in pods:
		pod_name = pod['Name']
		pod_namespace = pod['Namespace']
		joined_logs.append(crLogs.get_logs(pod_name, pod_namespace, kubeconfig))
		joined_logs.append("---------------------------------------")
		
	all_logs = "\n".join(joined_logs)
	url = "http://localhost:8080/crailogs"
	payload = {"logs": all_logs}

	try:
		response = requests.post(url, json=payload)
		response.raise_for_status()
		result = response.json()
		if 'output' in result:
			print(json.dumps(result['output'], indent=2))
	except requests.exceptions.RequestException as e:
		print(f"Error communicating with model service: {e}")
	except ValueError:
		print(f"Response was not valid JSON: {response.text}")
