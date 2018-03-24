import subprocess
import sys
import json
import yaml
import platform
import os
from crmetrics import CRBase

class AppURLFinder(CRBase):

	def get_ingresses(self, resources):
		ingress_list = []
		for resource in resources:
			#print(resource)
			if resource['Kind'] == 'Ingress':
				present = False
				for s in ingress_list:
					if s['Name'] == resource['Name']:
						present = True
						break
				if not present:
					ingress_list.append(resource)
		#print(ingress_list)
		return ingress_list

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
		#print(svc_list)
		return svc_list

	def get_host_from_ingress(self, ingresses, namespace, kubeconfig):
		appURL = ""
		for ingress in ingresses:
			cmd = 'kubectl get ingress ' + ingress['Name'] + ' -n ' + ingress['Namespace'] + ' -o json ' + kubeconfig
			#print(cmd)
			try:
				out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										stderr=subprocess.PIPE, shell=True).communicate()[0]

				if out:
					json_output = json.loads(out)
					#print(json_output)
					if 'tls' in json_output['spec']:
						host = json_output['spec']['tls'][0]['hosts'][0]
						appURL = "https://" + host.strip()
					else:
						host = json_output['spec']['rules'][0]['host']
						appURL = "http://" + host.strip()
					break
			except Exception as e:
				print(e)
		return appURL

	def get_svc_port(self, svcs, namespace, kubeconfig):
		nodePort = -1
		for svc in svcs:
			cmd = 'kubectl get service ' + svc['Name'] + ' -n ' + svc['Namespace'] + ' -o json ' + kubeconfig
			#print(cmd)
			try:
				out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										stderr=subprocess.PIPE, shell=True).communicate()[0]

				if out:
					json_output = json.loads(out)
					#print(json_output)
					service_type = json_output['spec']['type']
					if service_type == 'NodePort':
						nodePort = json_output['spec']['ports'][0]['nodePort']
						#print("NodePort:" + str(nodePort))
						break
			except Exception as e:
				print(e)
		return nodePort

	def get_node_ip(self, kubeconfig):
		cmd = 'kubectl describe node ' + kubeconfig
		out = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
		out = out.decode('utf-8')
		#print("Node o/p:" + out)
		for line in out.split("\n"):
			if 'ExternalIP' in line:
				parts = line.split(":")
				nodeIP = parts[1].strip()
				#print("Node IP:" + nodeIP)
				return nodeIP
		return ""

	def get_server_ip(self, kubeconfig):
		server_ip = ''
		parts = kubeconfig.split("=")
		kcfg = parts[1].strip()
		#print(parts)
		fp = open(kcfg, "r")
		#fp = open("/root/.kube/config", "r")
		contents = fp.read()
		content = ''
		try:
			content = json.loads(contents)
		except:
			content = yaml.safe_load(contents)

		cluster_list = content['clusters']
		for cluster in cluster_list:
			cluster_name = cluster['name']
			#print("Cluster name:" + cluster_name)
			if cluster_name == 'kubeplus-saas-consumer' or cluster_name == 'kubeplus-saas-provider':
				server_url = cluster['cluster']['server']
				#print(server_url)
				server_url = server_url.strip()
				parts = server_url.split(":")
				server_ip = parts[1].strip()
				#print(server_ip)
		return server_ip

if __name__ == '__main__':
	appURLFinder = AppURLFinder()
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
		resources = appURLFinder.get_resources_connections(kind, instance, namespace, kubeconfig)
		#print(resources)
	try:
		ingresses = appURLFinder.get_ingresses(resources)
		if len(ingresses) > 0:
			appURL = appURLFinder.get_host_from_ingress(ingresses, namespace, kubeconfig)
			print(appURL)
		else:
			svcs = appURLFinder.get_svc(resources)
			svcPort = appURLFinder.get_svc_port(svcs, namespace, kubeconfig)
			appIP = appURLFinder.get_node_ip(kubeconfig)
			if appIP == "":
				appIP = appURLFinder.get_server_ip(kubeconfig)
			if appIP == '':
				print("KubePlus SaaS Consumer context not found in the kubeconfig.")
				print("Cannot form app url.")
				exit()
			else:
				if "//" not in appIP:
					appIP = "//" + appIP
				appURL = "http:" + appIP + ":" + str(svcPort)
				appURL = appURL.strip()
				#print("App port:" + str(svcPort))
				print(appURL)
	except Exception as e:
		print(e)
