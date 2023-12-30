import subprocess
import sys
import json
import re
import os
import platform
import pprint
import time

import utils

class CRBase(object):

	def parse_pod_details(self, out, instance):
		pod_list = []
		for line in out.split("\n"):
			if 'NAME' not in line and line != '':
				line1 = ' '.join(line.split())
				parts = line1.split(" ")
				pod_info = {}
				pod_info['Name'] = parts[0]
				pod_info['Namespace'] = instance
				pod_list.append(pod_info)
		return pod_list
        	
	def get_pods_in_ns(self, kind, instance, kubeconfig):
		pod_list = []
		labelSelector = "partof=" + kind + "-" + instance
		labelSelector = labelSelector.lower()
		cmd = "kubectl get pods -n  " + instance + " " + kubeconfig
		#print(cmd)
		out = ''
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,stderr=subprocess.PIPE, shell=True).communicate()[0]
			out = out.decode('utf-8')
		except Exception as e:
			print(e)
		
		pod_list = self.parse_pod_details(out, instance)

		#print(out)

		#print(pod_list)
		return pod_list

	def get_pods(self, kind, instance, kubeconfig):
		pod_list = []
		labelSelector = "partof=" + kind + "-" + instance
		labelSelector = labelSelector.lower()
		cmd = "kubectl get pods -A -l " + labelSelector + " " + kubeconfig
		#print(cmd)
		out = ''
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,stderr=subprocess.PIPE, shell=True).communicate()[0]
			out = out.decode('utf-8')
		except Exception as e:
			print(e)
		
		#pod_list = self.parse_pod_details(out)
		for line in out.split("\n"):
			if 'NAME' not in line and line != '':
				line1 = ' '.join(line.split())
				parts = line1.split(" ")
				pod_info = {}
				pod_info['Name'] = parts[1]
				pod_info['Namespace'] = parts[0]
				pod_list.append(pod_info)

		#print(out)

		#print(pod_list)
		return pod_list

	def _get_kubeplus_namespace(self):
		kb_namespace = 'default'
		cmd = "kubectl get pods -A | grep kubeplus-deployment | awk '{print $1}'"
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]
			out = out.decode('utf-8')
			kb_namespace = out.strip()
			#print("KubePlus NS:" + kb_namespace)
		except Exception as e:
			print(e)
		return kb_namespace

	def get_resources_connections(self, kind, instance, namespace, kubeconfig):
		platf = platform.system()
		kubeplus_home = os.getenv('KUBEPLUS_HOME', '/')
		cmd = ''
		json_output = {}
		if platf == "Darwin":
			cmd = kubeplus_home + '/plugins/kubediscovery-macos connections ' 
		elif platf == "Linux":
			cmd = kubeplus_home + '/plugins/kubediscovery-linux connections '
		else:
			print("OS not supported:" + platf)
			return json_output
		kb_ns = self._get_kubeplus_namespace()
		cmd = cmd + kind + ' ' + instance + ' ' + namespace + ' --output=json ' + kubeconfig + ' --ignore=ServiceAccount:default,Namespace:' + kb_ns
		#print(cmd)
		out = ''
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
								   stderr=subprocess.PIPE, shell=True).communicate()[0]
			out = out.decode('utf-8').strip()
		except Exception as e:
			print(e)
		if out:
			#print(out)
			try:
				json_output = json.loads(out)
			except Exception as e:
				print(e)
		return json_output


class CRMetrics(CRBase):

	def _get_identity(self, custom_resource, custom_res_instance, namespace):
		cmd = 'kubectl get ' + custom_resource + ' ' + custom_res_instance + ' -n ' + namespace + ' -o json'
		accountidentity = ''
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]
			out = out.decode('utf-8')
			if out:
				json_output = json.loads(out)
				if json_output and 'metadata' in json_output:
					metadata = json_output['metadata']
					if 'annotations' in json_output['metadata']:
						annotations = json_output['metadata']['annotations']
						if 'accountidentity' in json_output['metadata']['annotations']:
							accountidentity = json_output['metadata']['annotations']['accountidentity']
		except Exception as e:
			print(e)
		return accountidentity

	def _get_composition(self, custom_resource, custom_res_instance, namespace):
		platf = platform.system()
		kubeplus_home = os.getenv('KUBEPLUS_HOME', '/')
		cmd = ''
		if platf == "Darwin":
			cmd = kubeplus_home + '/plugins/kubediscovery-macos composition ' + custom_resource + ' ' + custom_res_instance + ' ' + namespace
		if platf == "Linux":
			cmd = kubeplus_home + '/plugins/kubediscovery-linux composition ' + custom_resource + ' ' + custom_res_instance + ' ' + namespace
		out = ''
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
								   stderr=subprocess.PIPE, shell=True).communicate()[0]
			out = out.decode('utf-8')
		except Exception as e:
			print(e)
		json_output = {}
		if out:
			#print(out)
			try:
				json_output = json.loads(out)
			except Exception as e:
				print(e)
		return json_output

	def _count_resources(self, json_block):
		num_of_resources = 0
		if 'Children' in json_block:
			child_list = json_block['Children']
			num_of_resources = num_of_resources + len(child_list)
			for child in child_list:
				child_resources = self._count_resources(child)
				num_of_resources = num_of_resources + child_resources
		return num_of_resources

	def _parse_number_of_resources(self, composition):
		num_of_resources = 0
		if composition:
			json_block = composition[0]
			num_of_resources = self._count_resources(json_block)
		return num_of_resources

	def _get_pods(self, json_block):
		pod_list = []
		if json_block['Kind'] == 'Pod':
			pod_list.append(json_block)
		if 'Children' in json_block:
			child_list = json_block['Children']
			for child in child_list:
				child_pod_list = self._get_pods(child)
				pod_list = pod_list + child_pod_list
		return pod_list

	def _parse_number_of_pods(self, composition):
		pod_list = []
		if composition:
			json_block = composition[0]
			pod_list = self._get_pods(json_block)
		return pod_list

	def _get_pod(self, pod, kubeconfig=''):
		cmd = 'kubectl get pods ' + pod['Name'] + ' -n ' + pod['Namespace'] + ' -o json' + ' ' + kubeconfig
		out = ''
		#print(cmd)
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
								   stderr=subprocess.PIPE, shell=True).communicate()[0]
			out = out.decode('utf-8')
		except Exception as e:
			print(e)
		json_output = ''
		if out != '':
			json_output = json.loads(out)
		return json_output

	def _parse_number_of_hosts(self, pod_list, kubecfg=''):
		num_of_hosts = 0
		host_list = []
		for pod in pod_list:
			json_output = self._get_pod(pod, kubeconfig=kubecfg)
			if 'status' in json_output:
				if 'hostIP' in json_output['status']:
					hostIP = json_output['status']['hostIP']
					if hostIP not in host_list:
						host_list.append(hostIP)
		return len(host_list)

	def _parse_number_of_containers(self, pod_list, kubecfg=''):
		num_of_containers = 0
		for pod in pod_list:
			json_output = self._get_pod(pod, kubeconfig=kubecfg)
			containers = json_output['spec']['containers']
			num_of_containers = num_of_containers + len(containers)
			if 'initContainers' in json_output['spec']:
				init_containers = json_output['spec']['initContainers']
				num_of_containers = num_of_containers + len(init_containers)
		return num_of_containers

	def _num_of_not_running_pods(self, pod_list, kubecfg=''):
		num_of_not_running_pods = 0
		for pod in pod_list:
			json_output = self._get_pod(pod, kubeconfig=kubecfg)
			if json_output['status']['phase'] != 'Running':
				num_of_not_running_pods = num_of_not_running_pods + 1

		return num_of_not_running_pods

	def _parse_persistentvolumeclaims(self, pod_list, kubecfg=''):
		total_storage = 0
		pvc_list = []
		for pod in pod_list:
			pvc_details = {}
			json_output = self._get_pod(pod, kubeconfig=kubecfg)
			#print(json_output)
			#print("--------\n")
			if json_output['spec']['volumes']:
				volumes = json_output['spec']['volumes']
				for v in volumes:
					if 'persistentVolumeClaim' in v:
						pvc = v['persistentVolumeClaim']
						pvc_name = pvc['claimName']
						if pvc_name not in pvc_list:
							pvc_list.append(pvc_name)
							pvc_details['name'] = pvc_name
							pvc_details['namespace'] = json_output['metadata']['namespace']

		if len(pod_list) > 0:
			for pvc in pvc_details:
				pvc_name = pvc_details['name']
				pvc_ns = pvc_details['namespace']
				cmd = "kubectl get pvc " + pvc_name + ' -n ' + pvc_ns + " -o json"
				out = ''
				try:
					out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										   stderr=subprocess.PIPE, shell=True).communicate()[0]
					out = out.decode('utf-8')
					out = out.strip("\n")
				except Exception as e:
					print(e)
				if out != '':
					json_output = json.loads(out)
					if 'status' in json_output:
						phase = json_output['status']['phase']
						if phase == "Bound":
							if 'capacity' in json_output['status']:
								capacity = json_output['status']['capacity']
								storage = capacity['storage']
								#print("Storage:" + storage)
								temp = re.findall(r'\d+', storage) 
								storage_nums = list(map(int, temp))
								#print(storage_nums)
								total_storage = total_storage + storage_nums[0]
		return total_storage

	def _get_cpu_memory_usage_kubelet(self, pod_list, kubecfg=''):
		total_cpu = 0
		total_mem = 0

		platf = platform.system()
		kubeplus_home = os.getenv('KUBEPLUS_HOME', '/')
		cmd = ''
		if platf == "Darwin":
			cmd = kubeplus_home + '/plugins/kubediscovery-macos podmetrics '
		if platf == "Linux":
			cmd = kubeplus_home + '/plugins/kubediscovery-linux podmetrics '

		#print("POD LIST:")
		#print(pod_list)
		for pod in pod_list:
			#print("POD:")
			#print(pod)
			pod_json = self._get_pod(pod, kubeconfig=kubecfg)

            # If nodeName is not present in Pod spec then it means that Pod is not scheduled
			if 'nodeName' not in pod_json['spec']:
				continue

			#print(pod_json)
			podName = pod['Name']
			nodeName = pod_json['spec']['nodeName']
			podNS = pod_json['metadata']['namespace']
			#print("PodName:" + podName)
			#print("NodeName:" + nodeName)
			#print("PodNS:" + podNS)
			podMetricsCmd = cmd + " " + nodeName + " " + kubeconfig
			#print(podMetricsCmd)
			try:
				output = subprocess.Popen(podMetricsCmd, stdout=subprocess.PIPE,
										  stderr=subprocess.PIPE, shell=True).communicate()[0]
				output = output.decode('utf-8')
				output = output.strip("\n")
			except Exception as e:
				print(e)

			#print("stats/summary output:")
			#print(output)
			#print("===")
			nodeMetrics = json.loads(output)
			podMetricsJSONList = nodeMetrics["pods"]
			#print("====\n")
			#print(podMetricsJSONList)
			#print("====\n")

			for podMetric in podMetricsJSONList:
				podName1 = podMetric["podRef"]["name"].strip()
				podNamespace1 = podMetric["podRef"]["namespace"].strip()
				if podName1 == podName and podNamespace1 == podNS:
					cpuInNanoCores = podMetric["cpu"]["usageNanoCores"]
					memoryInBytes = podMetric["memory"]["workingSetBytes"]
					#print("CPU:" + str(cpuInNanoCores))
					#print("MEMORY:" + str(memoryInBytes))
					total_cpu = total_cpu + cpuInNanoCores
					total_mem = total_mem + memoryInBytes

		total_cpu_milli_cores = float(total_cpu) / 1000000
		total_mem_mib = float(total_mem) / (1024*1024)

		#print("TOTAL CPU:" + str(total_cpu_milli_cores) + " mi")
		#print("TOTAL MEM:" + str(total_mem_mib) + " Mib")

		return total_cpu_milli_cores, total_mem_mib

	def _get_cpu_memory_usage_kubectl_top(self, pod_list):
		pod_usage_map = {}
		total_cpu = 0
		total_mem = 0
		for pod in pod_list:
			cmd = "kubectl top pods " +  pod['Name'] + ' -n ' + pod['Namespace'] + " | grep -v NAME"
			out = ''
			try:
				#print(cmd)
				out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									   stderr=subprocess.PIPE, shell=True).communicate()[0]
				out = out.decode('utf-8')
				out = out.strip("\n")
			except Exception as e:
				print(e)
			parts_trimmed = []
			for line in out.split('\n'):
				parts = line.split(" ")
				for x in parts: 
					if x != "":
						parts_trimmed.append(x)
			if parts_trimmed:
				temp = re.findall(r'\d+', parts_trimmed[1]) 
				cpu_nums = list(map(int, temp))
				temp = re.findall(r'\d+', parts_trimmed[2]) 
				mem_nums = list(map(int, temp))
				total_cpu = total_cpu + cpu_nums[0]
				total_mem = total_mem + mem_nums[0]
		return total_cpu, total_mem

	def _get_cpu_memory_usage_rootres(self, kind, cmd, instance, namespace, account):
		account_identity = ''
		cpu = 0
		memory = 0
		kind = ''
		count = 0
		try:
			out1 = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]
			out1 = out1.decode('utf-8')
			json_output = json.loads(out1)
			kind = json_output['kind']
			if 'metadata' in json_output:
				if 'annotations' in json_output['metadata']:
					if 'accountidentity' in json_output['metadata']['annotations']:
						account_identity = json_output['metadata']['annotations']['accountidentity']
		except Exception as e:
			print(e)

		if account_identity == account:
			count = count + 1
			composition = self._get_composition(kind, instance, namespace)
			pod_list = self._parse_number_of_pods(composition)
			cpu, memory = self._get_cpu_memory_usage_kubelet(pod_list)
		return cpu, memory, count

	def _get_metrics_cr_instances(self, account):
		total_cpu = 0
		total_mem = 0
		total_count = 0
		cmd = 'kubectl get crds | grep -v NAME | awk \'{print $1}\''
		crds = ''
		try:
			crds = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]
			crds = crds.decode('utf-8')
			crds = crds.strip("\n")
		except Exception as e:
			print(e)

		for crd in crds.split('\n'):
			cmd = 'kubectl get ' + crd + " --all-namespaces | grep -v NAME"
			out = ''
			err = ''
			try:
				out, err = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										stderr=subprocess.PIPE, shell=True).communicate()
				out = out.decode('utf-8')
			except Exception as e:
				print(e)

			for line in out.split("\n"):
				if line != 'No resources found.':
					if line == '':
						continue
					parts = line.split(" ")
					parts_trimmed = []
					for x in parts: 
						if x != "":
							parts_trimmed.append(x)
					if parts_trimmed:
						namespace = parts_trimmed[0]
						instance = parts_trimmed[1]
						cmd1 = 'kubectl get ' + crd + ' ' + instance + ' -n ' + namespace + ' -o json'
						cpu, memory, count = self._get_cpu_memory_usage_rootres(crd, cmd1, instance, namespace, account)

						total_cpu = total_cpu + cpu
						total_mem = total_mem + memory
						total_count = total_count + count

		return total_cpu, total_mem, total_count

	def _get_metrics_deployments(self, account):
		cpu = 0
		mem = 0
		cpu, mem, count = self._get_metrics_kind('deployments', account)
		return cpu, mem, count

	def _get_metrics_statefulsets(self, account):
		cpu = 0
		mem = 0
		cpu, mem, count = self._get_metrics_kind('statefulsets', account)
		return cpu, mem, count

	def _get_metrics_replicasets(self, account):
		cpu = 0
		mem = 0
		cpu, mem, count = self._get_metrics_kind('replicasets', account)
		return cpu, mem, count

	def _get_metrics_daemonsets(self, account):
		cpu = 0
		mem = 0
		cpu, mem, count = self._get_metrics_kind('daemonsets', account)
		return cpu, mem, count

	def _get_metrics_replicationcontrollers(self, account):
		cpu = 0
		mem = 0
		cpu, mem, count = self._get_metrics_kind('replicationcontrollers', account)
		return cpu, mem, count

	def _get_pods_for_account(self, account):
		pod_list = []

		cmd = 'kubectl get pods --all-namespaces | grep -v NAME | awk \'{print $1 \" \" $2}\''
		out = ''
		#print(cmd)
		try:
			out, _ = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()
			out = out.decode('utf-8')
			out = out.strip()
		except Exception as e:
			print(e)

		#print("----")
		#print(out)
		#print("----")
		for p in out.split("\n"):
			parts = p.split(" ")
			pod = {}
			if len(parts) > 0:
				pod['Namespace'] = parts[0]
				pod['Name'] = parts[1]
				json_output = self._get_pod(pod)
				if 'metadata' in json_output:
					if 'annotations' in json_output['metadata']:
						if 'accountidentity' in json_output['metadata']['annotations']:
							account_identity = json_output['metadata']['annotations']['accountidentity']
							if account_identity == account:
								pod_list.append(pod)
							#count = count + 1
		return pod_list

	def _get_metrics_pods(self, account):
		cpu = 0
		mem = 0
		count = 0
		pod_list = self._get_pods_for_account(account)
		#print("Pods:")
		#print(pod_list)
		cpu, mem = self._get_cpu_memory_usage_kubelet(pod_list)
		return cpu, mem, len(pod_list)

	def _get_metrics_kind(self, kind, account):
		total_mem = 0
		total_cpu = 0
		total_count = 0

		cmd = 'kubectl get ' + kind + ' --all-namespaces | grep -v NAME | awk \'{print $1 \" \" $2}\''

		instances = ''
		try:
			instances = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]
			instances = instances.decode('utf-8')
			instances = instances.strip("\n")
		except Exception as e:
			print(e)

		if instances == 'No resources found.':
			return total_cpu, total_mem, total_count

		for line in instances.split("\n"):
			if line != '':
				parts = line.split(" ")
				namespace = parts[0]
				instance = parts[1]

				cmd1 = 'kubectl get ' + kind + ' ' + instance + ' -n ' + namespace + ' -o json'
				cpu, memory, count = self._get_cpu_memory_usage_rootres(kind, cmd1, instance, namespace, account)

				total_cpu = total_cpu + cpu
				total_mem = total_mem + memory
				total_count = total_count + count

		return total_cpu, total_mem, total_count

	def _get_pods_for_cr_connections(self, cr, cr_instance, namespace, kubeconfig, conn_op_format="flat"):
		pod_list = []
		platf = platform.system()
		kubeplus_home = os.getenv('KUBEPLUS_HOME', '/')
		cmd = ''
		kb_ns = self._get_kubeplus_namespace()
		if platf == "Darwin":
			cmd = kubeplus_home + '/plugins/kubediscovery-macos connections ' + cr + ' ' + cr_instance + ' ' + namespace + ' --output=' + conn_op_format + ' --ignore=ServiceAccount:default,Namespace:' + kb_ns
		if platf == "Linux":
			cmd = kubeplus_home + '/plugins/kubediscovery-linux connections ' + cr + ' ' + cr_instance + ' ' + namespace + ' --output=' + conn_op_format + ' --ignore=Namespace:'+ kb_ns
		parts = kubeconfig.split("=")
		cmd = cmd + " " + kubeconfig
		if cmd:
			#print(cmd)
			output = ''
			try:
				output = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										  stderr=subprocess.PIPE, shell=True).communicate()[0]
				output = output.decode('utf-8')
				output = output.strip("\n")
				#print(output)
			except Exception as e:
				print(e)

			if conn_op_format == "flat:":
				pod_list = self._parse_pods_from_connections_op(output)
			if conn_op_format == "json":
				try:
					json_output = json.loads(output)
					pod_list = utils.get_pods(json_output)
				except Exception as e:
					pass
					#print(e)
		return pod_list

	def _get_pods_for_service(self, service_name, namespace):
		pod_list = []
		platf = platform.system()
		kubeplus_home = os.getenv('KUBEPLUS_HOME', '/')
		cmd = ''
		if platf == "Darwin":
			cmd = kubeplus_home + '/plugins/kubediscovery-macos connections Service ' + service_name + ' ' + namespace + ' --output=json'
		if platf == "Linux":
			cmd = kubeplus_home + '/plugins/kubediscovery-linux connections Service ' + service_name + ' ' + namespace + ' --output=json'

		if cmd:
			output = ''
			try:
				output = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										  stderr=subprocess.PIPE, shell=True).communicate()[0]
				output = output.decode('utf-8')
			except Exception as e:
				print(e)

			if output:
				try:
					json_output = json.loads(output)
				except Exception as e:
					print(e)

				pod_list = utils.get_pods(json_output)
				res_list = utils.get_resources(json_output)

			#pod_list = self._parse_pods_from_connections_op(output)
		return pod_list, res_list

	def _parse_pods_from_connections_op(self, output):
		pod_list = []
		#print(output)
		start = False
		for line in output.split("\n"):
			if line.startswith("Level"):
				start = True
			if line.find("Branch") >= 0:
				start = False
			if line and start:
				parts = line.split(" ")
				kind = parts[1]
				instance = parts[2]
				kind_parts = kind.split(":")
				#print("Kind_parts:" + kind)
				if kind_parts[1] == "Pod":
					instance_parts = instance.split(":")
					instance_name = instance_parts[1]
					pod = {}
					pod['Namespace'] = namespace
					pod['Name'] = instance_name
					if len(pod_list) > 0:
						present = False
						for p in pod_list:
							if p['Name'] == instance_name and p['Namespace'] == namespace:
								present = True
								break
						if not present:
							pod_list.append(pod)
					else:
						pod_list.append(pod)
		return pod_list

	def _get_pods_for_helmrelease_2(self, release_name):
		cmd = "helm get " + release_name
		try:
			output = subprocess.Popen(cmd,
									  stdout=subprocess.PIPE,
									  stderr=subprocess.PIPE,
									  shell=True).communicate()[0]
			output = output.decode('utf-8')
			output = output.strip("\n")
		except Exception as e:
			print(e)

		pod_list_to_return = []

		output1 = output.decode("utf-8")
		#print(output1)
		processed_output = []
		start = False
		for line in output1.split("\n"):
			if not start and line == "---":
				start = True
			if start:
				processed_output.append(line)
				processed_output.append("\n")

		#print("CBC")
		#print(processed_output)
		yamls = ''.join(processed_output)
		#print("DEF")
		#print(yamls)
		yamls_bytes = yamls.encode()
		#print("EFG")
		#print(yamls_bytes)

		for project in yaml.load_all(yamls_bytes):
			#pprint.pprint(project)
			if project != None:
				kind = project['kind']
				name = ''
				namespace = 'default'
				if 'metadata' in project:
					if 'name' in project['metadata']:
						name = project['metadata']['name']
					if 'namespace' in project['metadata']:
						name = project['metadata']['namespace']

				if kind not in ['ConfigMap', 'CustomResourceDefinition', 'ClusterRole', 'ClusterRoleBinding']:
					if kind != '' and name != '' and namespace != '':
						#print("Kind:"+ kind + " Namespace:" + namespace + " Instance:" + instance)
						composition = self._get_composition(kind, name, namespace)
						pod_list = self._parse_number_of_pods(composition)
						if pod_list:
							#print(pod_list)
							for p in pod_list:
								pod = {}
								pod['Name'] = p['Name']
								pod['Namespace'] = p['Namespace']
								pod_list_to_return.append(pod)
		return pod_list_to_return

	def _get_pods_for_helmrelease(self, release_name):
		cmd = "helm get manifest " + release_name
		try:
			output = subprocess.Popen(cmd,
									  stdout=subprocess.PIPE,
									  stderr=subprocess.PIPE,
									  shell=True).communicate()[0]
			output = output.decode('utf-8')
			output = output.strip("\n")
		except Exception as e:
			print(e)

		# TODO: Include all the kinds that do not create a Pod.
		skipKinds = ['ConfigMap', 'CustomResourceDefinition', 'ClusterRole', 'ClusterRoleBinding',
					 'Service', 'ServiceAccount', 'Role', 'RoleBinding', 'Secret']
		pod_list_to_return = []
		kind = ''
		instance = ''
		namespace = ''
		reachedEnd = False
		for line in output.split("\n"):
			allparts = line.split(" ")
			parts = [i for i in allparts if i]
			if parts:
				if parts[0].startswith("apiVersion:"):
					kind = ''
					instance = ''
					namespace = ''
					reachedEnd = False
					processed = False
				if parts[0].startswith("kind:"):
					kind = parts[1]
				if parts[0].startswith("name:"):
					instance = parts[1]
				if parts[0].startswith("spec:"):
					reachedEnd = True
					if namespace == '':
						namespace = 'default'
				if parts[0].startswith("namespace:") and not reachedEnd:
					namespace = parts[1]

				if kind not in skipKinds:
					if kind != '' and instance != '' and namespace != '' and not processed:
						processed = True
						#print("Kind:"+ kind + " Namespace:" + namespace + " Instance:" + instance)
						time1 = int(round(time.time() * 1000))
						composition = self._get_composition(kind, instance, namespace)
						time2 = int(round(time.time() * 1000))
						#print("Composition time:" + str(time2-time1))

						time3 = int(round(time.time() * 1000))
						pod_list = self._parse_number_of_pods(composition)
						time4 = int(round(time.time() * 1000))
						#print("Pod list time:" + str(time4-time3))

						if pod_list:
							#print(pod_list)
							for p in pod_list:
								pod = {}
								pod['Name'] = p['Name']
								pod['Namespace'] = p['Namespace']
								pod_list_to_return.append(pod)
		return pod_list_to_return

	def _get_pods_from_connections_pod(self, pod_name, namespace):
		pod_list = []
		platf = platform.system()
		kubeplus_home = os.getenv('KUBEPLUS_HOME', '/')
		cmd = ''
		if platf == "Darwin":
			cmd = kubeplus_home + '/plugins/kubediscovery-macos connections Pod ' + ' ' + pod_name + ' ' + namespace + ' --output=json'
		if platf == "Linux":
			cmd = kubeplus_home + '/plugins/kubediscovery-linux connections Pod ' + ' ' + pod_name + ' ' + namespace + ' --output=json'

		#print(cmd)
		if cmd:
			output = ''
			try:
				output = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										  stderr=subprocess.PIPE, shell=True).communicate()[0]
				output = output.decode('utf-8')
				output = output.strip("\n")
			except Exception as e:
				print(e)

			if output:
				try:
					json_output = json.loads(output)
				except Exception as e:
					print(e)

				pod_list = utils.get_pods(json_output)
		return pod_list

	def _parse_network_bytes(self, cAdvisorLine):
		parts = cAdvisorLine.split("}")
		networkMetric = 0
		#print("PARTS:")
		#print(parts)
		#print("Len:" + str(int(len(parts))))
		if len(parts) == 2:
			metricPart = parts[1]
			part1 = metricPart.split(" ")
			part2 = []
			for p in part1:
				if p != '':
					part2.append(p)
			#print("Len1:" + str(int(len(part2))))
			if len(part2) == 2:
				networkMetric = part2[0]
		#print(networkMetric)
		return float(networkMetric)

	def _get_cadvisor_metrics(self, pod_list, kubecfg=''):
		networkReceiveBytesTotal = 0
		networkTransmitBytesTotal = 0
		oom_events = 0

		#print(pod_list)
		platf = platform.system()
		kubeplus_home = os.getenv('KUBEPLUS_HOME', '/')
		cmd = ''
		if platf == "Darwin":
			cmd = kubeplus_home + '/plugins/kubediscovery-macos networkmetrics ' 
		if platf == "Linux":
			cmd = kubeplus_home + '/plugins/kubediscovery-linux networkmetrics '

		for pod in pod_list:
			pod_json = self._get_pod(pod, kubeconfig=kubecfg)

            # If nodeName is not present in Pod spec then it means that Pod is not scheduled
			if 'nodeName' not in pod_json['spec']:
				continue

			#print(pod_json)
			podName = pod['Name']
			nodeName = pod_json['spec']['nodeName']
			podNS = pod_json['metadata']['namespace']
			#print("PodName:" + podName)
			#print("NodeName:" + nodeName)
			#print("PodNS:" + podNS)
			networkMetricsCmd = cmd + " " + nodeName + " " + kubeconfig
			try:
				output = subprocess.Popen(networkMetricsCmd, stdout=subprocess.PIPE,
										  stderr=subprocess.PIPE, shell=True).communicate()[0]
				output = output.decode('utf-8')
				output = output.strip("\n")
			except Exception as e:
				print(e)

			pod_found_network_receive_bytes = {}
			pod_found_network_transmit_bytes = {}
			pod_oom_events = {}
			if output != '':
				#print("-------\n")
				#print(output)
				for line in output.split("\n"):
					if 'container_network_receive_bytes_total' in line and podName in line:
						if podName not in pod_found_network_receive_bytes:
							#print("ABC ---\n")
							#print(line)
							pod_found_network_receive_bytes[podName] = line
							networkReceiveBytes = self._parse_network_bytes(line)
							networkReceiveBytesTotal = networkReceiveBytesTotal + networkReceiveBytes
					if 'container_network_transmit_bytes_total' in line and podName in line:
						if podName not in pod_found_network_transmit_bytes:
							#print("DEF ---\n")
							#print(line)
							pod_found_network_transmit_bytes[podName] = line
							networkTransmitBytes = self._parse_network_bytes(line)
							networkTransmitBytesTotal = networkTransmitBytesTotal + networkTransmitBytes
					if 'container_oom_events_total' in line and podName in line:
						if podName not in pod_oom_events:
							#print("DEF ---\n")
							#print(line)
							pod_oom_events[podName] = line
							parts = line.split(" ")
							oom_events = parts[1].strip()
				#print("--------\n")

		#print("NetworkReceiveBytesTotal:" + str(networkReceiveBytesTotal))
		#print("NetworkTransmitBytesTotal:" + str(networkTransmitBytesTotal))
		return networkReceiveBytesTotal, networkTransmitBytesTotal, oom_events

	def _get_metrics_creator_account_with_connections(self, account):

		#print("---------------------------------------------------------- ")
		#print(" Creator Account Identity: " + account)
		#print("---------------------------------------------------------- ")

		#print("Finding Pods created by account.." + account)

		print("---------------------------------------------------------- ")
		pod_list = self._get_pods_for_account(account)
		print(pod_list)
		pod_list_for_metrics = []

		for pod in pod_list:
			#print("Discovering connections for..." + pod['Namespace'] + "/" + pod['Name'])
			connections_pod_list = self._get_pods_from_connections_pod(pod['Name'], pod['Namespace'])
			#print(connections_pod_list)

			for p in connections_pod_list:
				seen = False
				for p1 in pod_list_for_metrics:
					if p1['Name'] == p['Name'] and p1['Namespace'] == p['Namespace']:
						seen = True
				if not seen:
					pod_list_for_metrics.append(p)

		print("Finding metrics for:")
		for pod in pod_list_for_metrics:
			print("    " + pod['Name'])
		#print(pod_list_for_metrics)

		cpu, mem = self._get_cpu_memory_usage_kubelet(pod_list_for_metrics)
		storage = 0
		for p in pod_list_for_metrics:
			stor = self._parse_persistentvolumeclaims([p])
			storage = storage + stor

		num_of_containers = self._parse_number_of_containers(pod_list_for_metrics)
		num_of_hosts = self._parse_number_of_hosts(pod_list_for_metrics)

		print("Kubernetes Resources created:")
		print("    Number of Pods: " + str(len(pod_list_for_metrics)))
		print("        Number of Containers: " + str(num_of_containers))
		print("        Number of Nodes: " + str(num_of_hosts))
		print("Underlying Physical Resoures consumed:")
		print("    Total CPU(cores): " + str(cpu) + "m")
		print("    Total MEMORY(bytes): " + str(mem) + "Mi")
		print("    Total Storage(bytes): " + str(storage) + "Gi")
		print("---------------------------------------------------------- ")

	def get_metrics_creator_account(self, account):
		# 1. Get all custom resource instances - their count
		all_cpu = 0
		all_mem = 0

		print("---------------------------------------------------------- ")
		print(" Creator Account Identity: " + account)
		print("---------------------------------------------------------- ")

		print("Checking Custom Resources..")
		cr_cpu, cr_mem, cr_count = self._get_metrics_cr_instances(account)

		# 2. Get all deployments
		print("Checking Deployments..")
		dep_cpu, dep_mem, dep_count = self._get_metrics_deployments(account)

		# 3. Get all statefulsets
		print("Checking StatefulSets..")
		ssets_cpu, ssets_mem, ss_count = self._get_metrics_statefulsets(account)

		# 4. Get all replicasets - 
		# Replicasets seem to carry over annotations from deployment. 
		# So we will ignore the memory and cpu of deployments from final totals.
		# To correctly count replicasets we will subtract dep_count from rsets_count
		print("Checking ReplicaSets..")
		rsets_cpu, rsets_mem, rsets_count = self._get_metrics_replicasets(account)
		rsets_count = rsets_count - dep_count

		# 5. Get all daemonsets
		print("Checking DaemonSets..")
		dsets_cpu, dsets_mem, dsets_count = self._get_metrics_daemonsets(account)

		# 6. Get all replicationcontrollers
		print("Checking ReplicationControllers..")
		rcsets_cpu, rcsets_mem, rcsets_count = self._get_metrics_replicationcontrollers(account)

		# 7. Get all pods
		print("Checking Pods..")
		p_cpu, p_mem, p_count = self._get_metrics_pods(account)

		all_cpu = cr_cpu + ssets_cpu + rsets_cpu + dsets_cpu + rcsets_cpu + p_cpu
		all_mem = cr_mem + ssets_mem + rsets_mem + dsets_mem + rcsets_mem + p_mem

		print("Kubernetes Resources created:")
		print("    Number of Custom Resources: " + str(cr_count))
		print("    Number of Deployments: " + str(dep_count))
		print("    Number of StatefulSets: " + str(ss_count))
		print("    Number of ReplicaSets: " + str(rsets_count))
		print("    Number of DaemonSets: " + str(dsets_count))
		print("    Number of ReplicationControllers: " + str(rcsets_count))
		print("    Number of Pods: " + str(p_count))
		print("Underlying Physical Resoures consumed:")
		print("    Total CPU(cores): " + str(all_cpu) + "m")
		print("    Total MEMORY(bytes): " + str(all_mem) + "Mi")
		print("    Total Storage(bytes): (Upcoming)")

	def get_metrics_cr(self, custom_resource, custom_res_instance, namespace, follow_connections, opformat, kubeconfig):
		accountidentity = self._get_identity(custom_resource, custom_res_instance, namespace)
		accountidentity = ''
		pod_list = []
		if follow_connections == "false":
			composition = self._get_composition(custom_resource, custom_res_instance, namespace)
			num_of_resources = self._parse_number_of_resources(composition)
			pod_list = self._parse_number_of_pods(composition)
		if follow_connections == "true":
			num_of_resources = "-"
			conn_op_format = "json"
			#pod_list = self._get_pods_for_cr_connections(custom_resource, custom_res_instance, namespace, kubeconfig, conn_op_format)
			#pod_list = self.get_pods(custom_resource, custom_res_instance, kubeconfig) # uses label selectors
			pod_list = self.get_pods_in_ns(custom_resource, custom_res_instance, kubeconfig) # queries Pods in the CR NS 
			if len(pod_list) == 0:
			    # uses kubectl connections plugin - slower than label selectors
			    pod_list = self._get_pods_for_cr_connections(custom_resource, custom_res_instance, namespace, kubeconfig, conn_op_format)

		#cpu, memory = self._get_cpu_memory_usage(pod_list)
		#print(pod_list)        
		num_of_containers_conn = self._parse_number_of_containers(pod_list, kubecfg=kubeconfig)
		total_storage_conn = self._parse_persistentvolumeclaims(pod_list, kubecfg=kubeconfig)
		num_of_hosts_conn = self._parse_number_of_hosts(pod_list, kubecfg=kubeconfig)
		cpu_conn, memory_conn = self._get_cpu_memory_usage_kubelet(pod_list, kubecfg=kubeconfig)
		networkReceiveBytesTotal, networkTransmitBytesTotal, oom_events = self._get_cadvisor_metrics(pod_list, kubecfg=kubeconfig)

		num_of_not_running_pods = self._num_of_not_running_pods(pod_list, kubecfg=kubeconfig)

		num_of_pods = len(pod_list)
		num_of_containers = num_of_containers_conn
		num_of_hosts = num_of_hosts_conn
		cpu = cpu_conn
		memory = memory_conn
		total_storage = total_storage_conn

		if opformat == 'json':
			op = {}
			op['accountidentity'] = accountidentity
			op['subresources'] = num_of_resources
			op['pods'] = str(num_of_pods)
			op['containers'] = str(num_of_containers)
			op['nodes'] = str(num_of_hosts)
			op['cpu'] = str(cpu) + " m"
			op['memory'] = str(memory) + " Mi"
			op['storage'] = str(total_storage) + " Gi"
			op['networkReceiveBytes'] = str(networkReceiveBytesTotal) + " bytes"
			op['networkTransmitBytes'] = str(networkTransmitBytesTotal) + " bytes"
			op['notRunningPods'] = str(num_of_not_running_pods)
			op['oom_events'] = str(oom_events)
			json_op = json.dumps(op)
			print(json_op)
		elif opformat == 'prometheus':
			millis = int(round(time.time() * 1000))
			timeInMillis = str(millis)
			metricsToReturn = ''
			cpuMetrics = 'cpu{custom_resource="'+custom_res_instance+'"} ' + str(cpu) + ' ' + timeInMillis
			memoryMetrics = 'memory{custom_resource="'+custom_res_instance+'"} ' + str(memory) + ' ' + timeInMillis
			storageMetrics = 'storage{custom_resource="'+custom_res_instance+'"} ' + str(total_storage) + ' ' + timeInMillis
			networkReceiveBytes = 'network_receive_bytes_total{custom_resource="'+custom_res_instance+'"} ' + str(networkReceiveBytesTotal) + ' ' + timeInMillis
			networkTransmitBytes = 'network_transmit_bytes_total{custom_resource="'+custom_res_instance+'"} ' + str(networkTransmitBytesTotal) + ' ' + timeInMillis
			numOfPods = 'pods{custom_resource="'+custom_res_instance+'"} ' + str(num_of_pods) + ' ' + timeInMillis
			numOfContainers = 'containers{custom_resource="'+custom_res_instance+'"} ' + str(num_of_containers) + ' ' + timeInMillis
			numOfNotRunningPods = 'not_running_pods{custom_resource="'+custom_res_instance+'"} ' + str(num_of_not_running_pods) + ' ' + timeInMillis

			oomEvents = 'oom_events{custom_resource="'+custom_res_instance+'"} ' + str(oom_events) + ' ' + timeInMillis
                         
			metricsToReturn = cpuMetrics + "\n" + memoryMetrics + "\n" + storageMetrics + "\n" + numOfPods + "\n" + numOfContainers + "\n" + networkReceiveBytes + "\n" + networkTransmitBytes + "\n" + numOfNotRunningPods + "\n" + oomEvents
			print(metricsToReturn)
		elif opformat == 'pretty':
			#print("---------------------------------------------------------- ")
			#print(" Creator Account Identity: " + accountidentity)
			#print("---------------------------------------------------------- ")
			print("---------------------------------------------------------- ")
			print("Kubernetes Resources created:")
			print("    Number of Sub-resources: " + str(num_of_resources))
			print("    Number of Pods: " + str(num_of_pods))
			print("        Number of Containers: " + str(num_of_containers))
			print("        Number of Nodes: " + str(num_of_hosts))
			print("        Number of Not Running Pods: " + str(num_of_not_running_pods))
			print("Underlying Physical Resoures consumed:")
			print("    Total CPU(cores): " + str(cpu) + "m")
			print("    Total MEMORY(bytes): " + str(memory) + "Mi")
			print("    Total Storage(bytes): " + str(total_storage) + "Gi")
			print("    Total Network bytes received: " + str(networkReceiveBytesTotal))
			print("    Total Network bytes transferred: " + str(networkTransmitBytesTotal))
			print("---------------------------------------------------------- ")
		else:
			print("Unknown output format specified. Accepted values: pretty, json, prometheus")

	def get_metrics_service(self, service_name, namespace):
		print("---------------------------------------------------------- ")
		pod_list, res_list = self._get_pods_for_service(service_name, namespace)
		#print(res_list)
		print("Kubernetes Resources created:")
		print("    Total Number of Resources: " + str(len(res_list)))
		print("    Number of Pods: " + str(len(pod_list)))

		num_of_containers = self._parse_number_of_containers(pod_list)
		print("        Number of Containers: " + str(num_of_containers))

		total_storage = self._parse_persistentvolumeclaims(pod_list)

		num_of_hosts = self._parse_number_of_hosts(pod_list)
		print("        Number of Nodes: " + str(num_of_hosts))

		cpu, memory = self._get_cpu_memory_usage(pod_list)

		print("Underlying Physical Resoures consumed:")
		print("    Total CPU(cores): " + str(cpu) + "m")
		print("    Total MEMORY(bytes): " + str(memory) + "Mi")
		print("    Total Storage(bytes): " + str(total_storage) + "Gi")
		print("---------------------------------------------------------- ")

	def get_metrics_helmrelease(self, release_name):
		pod_list = self._get_pods_for_helmrelease(release_name)

		time1 = int(round(time.time() * 1000))
		num_of_containers = self._parse_number_of_containers(pod_list)
		time2 = int(round(time.time() * 1000))

		#print("      time:" + str(time2-time1))

		# TODO: What should be the namespace parameter for Helm releases?
		# Currently setting to "default"
		time1 = int(round(time.time() * 1000))
		total_storage = self._parse_persistentvolumeclaims(pod_list)
		time2 = int(round(time.time() * 1000))
		#print("      pvc time:" + str(time2-time1))

		time1 = int(round(time.time() * 1000))
		num_of_hosts = self._parse_number_of_hosts(pod_list)
		time2 = int(round(time.time() * 1000))
		#print("      time:" + str(time2-time1))

		cpu, memory = self._get_cpu_memory_usage(pod_list)

		metrics_helm_release = {}
		metrics_helm_release['num_of_pods'] = str(len(pod_list))
		metrics_helm_release['num_of_containers'] = str(num_of_containers)
		metrics_helm_release['num_of_hosts'] = str(num_of_hosts)
		metrics_helm_release['cpu'] = str(cpu)
		metrics_helm_release['memory'] = str(memory)
		metrics_helm_release['storage'] = str(total_storage)
		return metrics_helm_release

	def print_metrics_helmrelease(self, metrics_helm_release):
		num_of_pods = metrics_helm_release['num_of_pods']
		num_of_containers = metrics_helm_release['num_of_containers']
		num_of_hosts = metrics_helm_release['num_of_hosts']
		cpu = metrics_helm_release['cpu']
		memory = metrics_helm_release['memory']
		total_storage = metrics_helm_release['storage']

		print("---------------------------------------------------------- ")
		print("Kubernetes Resources created:")
		print("    Number of Pods: " + num_of_pods)
		print("        Number of Containers: " + num_of_containers)
		print("    Number of Nodes: " + num_of_hosts)
		print("Underlying Physical Resoures consumed:")
		print("    Total CPU(cores): " + str(cpu) + "m")
		print("    Total MEMORY(bytes): " + str(memory) + "Mi")
		print("    Total Storage(bytes): " + str(total_storage) + "Gi")
		print("---------------------------------------------------------- ")

	def prometheus_metrics_helmrelease(self, release_name, metrics_helm_release):
		millis = int(round(time.time() * 1000))
		metricsToReturn = ''
		cpu = metrics_helm_release['cpu']
		memory = metrics_helm_release['memory']
		storage = metrics_helm_release['storage']
		num_of_pods = metrics_helm_release['num_of_pods']
		num_of_containers = metrics_helm_release['num_of_containers']
		cpuMetrics = 'cpu{helmrelease="'+release_name+'"} ' + str(cpu) + ' ' + str(millis)
		memoryMetrics = 'memory{helmrelease="'+release_name+'"} ' + str(memory) + ' ' + str(millis)
		storageMetrics = 'storage{helmrelease="'+release_name+'"} ' + str(storage) + ' ' + str(millis)
		numOfPods = 'pods{helmrelease="'+release_name+'"} ' + str(num_of_pods) + ' ' + str(millis)
		numOfContainers = 'containers{helmrelease="'+release_name+'"} ' + str(num_of_containers) + ' ' + str(millis)
		metricsToReturn = cpuMetrics + "\n" + memoryMetrics + "\n" + storageMetrics + "\n" + numOfPods + "\n" + numOfContainers
		return metricsToReturn

if __name__ == '__main__':
	crMetrics = CRMetrics()

	res_type = sys.argv[1]

	if res_type == "cr":
		custom_resource = sys.argv[2]
		custom_resource_instance = sys.argv[3]
		namespace = sys.argv[4]
		outputformat = sys.argv[5]
		follow_connections = sys.argv[6]
		kubeconfig = sys.argv[7]
		crMetrics.get_metrics_cr(custom_resource, custom_resource_instance, namespace, follow_connections, outputformat, kubeconfig)
	
	if res_type == "account":
		creator_account = sys.argv[2]
		if len(sys.argv) == 4:
			crMetrics._get_metrics_creator_account_with_connections(creator_account)
		else:
			crMetrics.get_metrics_creator_account(creator_account)

	if res_type == "service":
		service_name = sys.argv[2]
		namespace = sys.argv[3]
		crMetrics.get_metrics_service(service_name, namespace)

	if res_type == "helmrelease":
		release_name = sys.argv[2]
		op_format = sys.argv[3]
		metrics_helm_release = crMetrics.get_metrics_helmrelease(release_name)
		if op_format == "pretty":
			crMetrics.print_metrics_helmrelease(metrics_helm_release)
		elif op_format == "prometheus":
			prom_metrics = crMetrics.prometheus_metrics_helmrelease(release_name, metrics_helm_release)
			print(prom_metrics)
		elif op_format == "json":
			metrics_helm_release_json = json.dumps(metrics_helm_release)
			print(metrics_helm_release_json)
		else:
			print("Unrecognized output format. Supported formats - pretty/json/prometheus")
