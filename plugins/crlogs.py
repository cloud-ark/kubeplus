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
				print(out)
			except Exception as e:
				print(e)

	def get_logs(self, pod, namespace):
		cmd = 'kubectl get pods ' + pod + ' -n ' + namespace + ' -o json'
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]

			json_output = json.loads(out)
			containers = json_output['spec']['containers']
			self._get_container_logs(pod, namespace, containers)
			
			if 'initContainers' in json_output['spec']:
				init_containers = json_output['spec']['initContainers']
				self._get_container_logs(pod, namespace, init_containers)

		except Exception as e:
			print(e)

if __name__ == '__main__':
	crLogs = CRLogs()
	crLogs.get_logs(sys.argv[1], sys.argv[2])
