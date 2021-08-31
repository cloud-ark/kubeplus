import subprocess
import sys
import json
import platform
import os
from crmetrics import CRBase

class AppPortFinder(CRBase):

	def get_svc(self, resources):
		svc_list = []
		for resource in resources:
			#print(resource)
			if resource['Kind'] == 'Service':
				present = False
				for s in svc_list:
					if s['Name'] == resource['Name']:
						present = True
						break
				if not present:
					svc_list.append(resource)
		#print(pod_list)
		return svc_list

	def get_svc_port(self, svcs, namespace, kubeconfig):
		nodePort = -1
		for svc in svcs:
			cmd = 'kubectl get service ' + svc['Name'] + ' -n ' + namespace + ' -o json ' + kubeconfig
			#print(cmd)
			try:
				out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										stderr=subprocess.PIPE, shell=True).communicate()[0]

				if out:
					json_output = json.loads(out)
					service_type = json_output['spec']['type']
					if service_type == 'NodePort':
						nodePort = json_output['spec']['ports'][0]['nodePort']
						#print("NodePort:" + str(nodePort))
						break
			except Exception as e:
				print(e)
		return nodePort

if __name__ == '__main__':
	appPortFinder = AppPortFinder()
	#crLogs.get_logs(sys.argv[1], sys.argv[2])
	#resources = sys.argv[1]
	relation = sys.argv[1]
	kind = sys.argv[2]
	instance = sys.argv[3]
	namespace = sys.argv[4]
	kubeconfig = sys.argv[5]
	#print(kind + " " + instance + " " + namespace + " " + kubeconfig)
	resources = {}
	if relation == 'connections':
		resources = appPortFinder.get_resources_connections(kind, instance, namespace, kubeconfig)
		#print(resources)
	svcs = appPortFinder.get_svc(resources)
	svcPort = appPortFinder.get_svc_port(svcs, namespace, kubeconfig)
	#print("App port:" + str(svcPort))
	print(svcPort)
