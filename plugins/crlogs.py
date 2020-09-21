import subprocess
import sys
import json

class CRLogs(object):

	def _get_container_logs(self, pod, namespace, containers):
		for c in containers:
			container = c['name']
			cmd = 'kubectl logs ' + pod + ' -n ' + namespace + ' -c ' + container

			print("======== Pod::" + pod + "/container::" + container + " ===========")
			try:
				out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										stderr=subprocess.PIPE, shell=True).communicate()[0]
				if out:
					print(out)
			except Exception as e:
				print(e)

	def get_logs(self, pod, namespace):
		cmd = 'kubectl get pods ' + pod + ' -n ' + namespace + ' -o json'
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]

			if out:
				json_output = json.loads(out)
				containers = json_output['spec']['containers']
				self._get_container_logs(pod, namespace, containers)
			
				if 'initContainers' in json_output['spec']:
					init_containers = json_output['spec']['initContainers']
					self._get_container_logs(pod, namespace, init_containers)

		except Exception as e:
			print(e)

	def get_pods(self, resources):
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
	resources = sys.argv[1]
	resource_json = json.loads(resources)
	namespace = sys.argv[2]
	pods = crLogs.get_pods(resource_json)
	for pod in pods:
		pod_name = pod['Name']
		print(pod_name)
		crLogs.get_logs(pod_name, namespace)
