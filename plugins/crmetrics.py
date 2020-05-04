import subprocess
import sys
import json
import re
import platform

class CRMetrics(object):

	def _get_identity(self, custom_resource, custom_res_instance, namespace):

		cmd = 'kubectl get ' + custom_resource + ' ' + custom_res_instance + ' -n ' + namespace + ' -o json'

		accountidentity = ''
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]

			json_output = json.loads(out)

			if 'metadata' in json_output:
				metadata = json_output['metadata']
				#print(metadata)
				if 'annotations' in json_output['metadata']:
					annotations = json_output['metadata']['annotations']
					#print(annotations)
					if 'accountidentity' in json_output['metadata']['annotations']:
						accountidentity = json_output['metadata']['annotations']['accountidentity']
						#print(accountidentity)

		except Exception as e:
			print(e)

		return accountidentity

	def _get_composition(self, custom_resource, custom_res_instance, namespace):

		cmd = "\"/apis/platform-as-code/v1/composition?kind=" + custom_resource + "&instance=" + custom_res_instance + "&namespace=" + namespace + "\""
		full_cmd = 'kubectl get --raw ' + cmd

		out = ''
		try:
			out = subprocess.Popen(full_cmd, stdout=subprocess.PIPE,
								   stderr=subprocess.PIPE, shell=True).communicate()[0]
		except Exception as e:
			print(e)

		json_output = json.loads(out)

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

	def _get_pod(self, pod):
		cmd = 'kubectl get pods ' + pod['Name'] + ' -n ' + pod['Namespace'] + ' -o json'
		out = ''
		try:
			out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
								   stderr=subprocess.PIPE, shell=True).communicate()[0]
		except Exception as e:
			print(e)

		json_output = ''
		if out != '':
			json_output = json.loads(out)
		return json_output

	def _parse_number_of_hosts(self, pod_list):
		num_of_hosts = 0

		host_list = []
		for pod in pod_list:
			json_output = self._get_pod(pod)
			if 'status' in json_output:
				if 'hostIP' in json_output['status']:
					hostIP = json_output['status']['hostIP']
					if hostIP not in host_list:
						host_list.append(hostIP)

		return len(host_list)

	def _parse_number_of_containers(self, pod_list):
		num_of_containers = 0

		for pod in pod_list:
			json_output = self._get_pod(pod)
			containers = json_output['spec']['containers']
			#print(len(containers))
			num_of_containers = num_of_containers + len(containers)

			init_containers = json_output['spec']['containers']
			#print(len(init_containers))
			num_of_containers = num_of_containers + len(init_containers)

		return num_of_containers

	def _get_cpu_memory_usage(self, pod_list):
		pod_usage_map = {}

		total_cpu = 0
		total_mem = 0
		for pod in pod_list:
			#print(pod['Name'])
			cmd = "kubectl top pods " +  pod['Name'] + ' -n ' + pod['Namespace'] + " | grep -v NAME"

			out = ''
			try:
				out = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									   stderr=subprocess.PIPE, shell=True).communicate()[0]
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
				#print(cpu_nums)

				temp = re.findall(r'\d+', parts_trimmed[2]) 
				mem_nums = list(map(int, temp)) 
				#print(mem_nums)

				total_cpu = total_cpu + cpu_nums[0]
				#print(total_cpu)

				total_mem = total_mem + mem_nums[0]
				#print(total_mem)

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
			#print(" Number of Pods: " + str(len(pod_list)))

			cpu, memory = self._get_cpu_memory_usage(pod_list)
			#print("    Custom Resource: " + kind)
			#print("    Kind: " + kind)
			#print("    Name: " + instance)
			#print("    Namespace: " + namespace)
			#print("    CPU(cores): " + str(cpu) + "m")
			#print("    MEMORY(bytes): " + str(memory) + "Mi")
			#print("    ------    ")

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
			except Exception as e:
				print(e)

			for line in out.split("\n"):
				if line != 'No resources found.': #or err != 'No resources found.':
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

	def _get_metrics_pods(self, account):
		cpu = 0
		mem = 0
		count = 0
		pod_list = []

		cmd = 'kubectl get pods --all-namespaces | grep -v NAME | awk \'{print $1 \" \" $2}\''
		out = ''
		try:
			out, _ = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()
			out = out.strip()
		except Exception as e:
			print(e)

		for p in out.split("\n"):
			#print(p)
			parts = p.split(" ")
			pod = {}
			pod['Namespace'] = parts[0]
			pod['Name'] = parts[1]
			json_output = self._get_pod(pod)
			if 'metadata' in json_output:
				if 'annotations' in json_output['metadata']:
					if 'accountidentity' in json_output['metadata']['annotations']:
						account_identity = json_output['metadata']['annotations']['accountidentity']
						if account_identity == account:
							pod_list.append(pod)
							count = count + 1
							#print("    Name: " + parts[1])
							#print("    Namespace: " + parts[0])

		cpu, mem = self._get_cpu_memory_usage(pod_list)
		#print("    CPU(cores): " + str(cpu) + "m")
		#print("    MEMORY(bytes): " + str(mem) + "Mi")
		#print("    ------    ")

		return cpu, mem, count

	def _get_metrics_kind(self, kind, account):
		total_mem = 0
		total_cpu = 0
		total_count = 0

		cmd = 'kubectl get ' + kind + ' --all-namespaces | grep -v NAME | awk \'{print $1 \" \" $2}\''

		instances = ''
		try:
			instances = subprocess.Popen(cmd, stdout=subprocess.PIPE,
									stderr=subprocess.PIPE, shell=True).communicate()[0]
			instances = instances.strip("\n")
		except Exception as e:
			print(e)

		for line in instances.split("\n"):
			parts = line.split(" ")
			namespace = parts[0]
			instance = parts[1]

			cmd1 = 'kubectl get ' + kind + ' ' + instance + ' -n ' + namespace + ' -o json'
			cpu, memory, count = self._get_cpu_memory_usage_rootres(kind, cmd1, instance, namespace, account)

			total_cpu = total_cpu + cpu
			total_mem = total_mem + memory
			total_count = total_count + count

		return total_cpu, total_mem, total_count

	def _get_pods_for_service(self, service_name, namespace):
		pod_list = []
		platf = platform.system()
		cmd = ''
		if platf == "Darwin":
			cmd = './plugins/kubediscovery-macos connections Service ' + service_name + ' ' + namespace
		if platf == "Linux":
			cmd = './plugins/kubediscovery-linux connections Service ' + service_name + ' ' + namespace

		if cmd:
			output = ''
			try:
				output = subprocess.Popen(cmd, stdout=subprocess.PIPE,
										  stderr=subprocess.PIPE, shell=True).communicate()[0]
				output = output.strip("\n")
			except Exception as e:
				print(e)

			for line in output.split("\n"):
				if line:
					parts = line.split(" ")
					kind = parts[1]
					instance = parts[2]
					kind_parts = kind.split(":")
					if kind_parts[1] == "Pod":
						instance_parts = instance.split(":")
						instance_name = instance_parts[1]
						pod = {}
						pod['Namespace'] = namespace
						pod['Name'] = instance_name
						pod_list.append(pod)
		return pod_list

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

		print(" Number of Custom Resources: " + str(cr_count))
		print(" Number of Deployments: " + str(dep_count))
		print(" Number of StatefulSets: " + str(ss_count))
		print(" Number of ReplicaSets: " + str(rsets_count))
		print(" Number of DaemonSets: " + str(dsets_count))
		print(" Number of ReplicationControllers: " + str(rcsets_count))
		print(" Number of Pods: " + str(p_count))
		print("Total CPU(cores): " + str(all_cpu) + "m")
		print("Total MEMORY(bytes): " + str(all_mem) + "Mi")


	def get_metrics_cr(self, custom_resource, custom_res_instance, namespace):

		print("---------------------------------------------------------- ")
		accountidentity = self._get_identity(custom_resource, custom_res_instance, namespace)
		print(" Creator Account Identity: " + accountidentity)
		print("---------------------------------------------------------- ")

		composition = self._get_composition(custom_resource, custom_res_instance, namespace)
		num_of_resources = self._parse_number_of_resources(composition)
		print(" Number of Sub-resources: " + str(num_of_resources))

		pod_list = self._parse_number_of_pods(composition)
		print(" Number of Pods: " + str(len(pod_list)))

		num_of_containers = self._parse_number_of_containers(pod_list)
		print(" Number of Containers: " + str(num_of_containers))

		num_of_hosts = self._parse_number_of_hosts(pod_list)
		print(" Number of Nodes: " + str(num_of_hosts))

		cpu, memory = self._get_cpu_memory_usage(pod_list)

		print("Total CPU(cores): " + str(cpu) + "m")
		print("Total MEMORY(bytes): " + str(memory) + "Mi")
		print("---------------------------------------------------------- ")


	def get_metrics_service(self, service_name, namespace):

		print("---------------------------------------------------------- ")
		pod_list = self._get_pods_for_service(service_name, namespace)
		print(" Number of Pods: " + str(len(pod_list)))

		num_of_containers = self._parse_number_of_containers(pod_list)
		print(" Number of Containers: " + str(num_of_containers))

		num_of_hosts = self._parse_number_of_hosts(pod_list)
		print(" Number of Nodes: " + str(num_of_hosts))

		cpu, memory = self._get_cpu_memory_usage(pod_list)

		print("Total CPU(cores): " + str(cpu) + "m")
		print("Total MEMORY(bytes): " + str(memory) + "Mi")
		print("---------------------------------------------------------- ")



if __name__ == '__main__':
	crLogs = CRMetrics()

	res_type = sys.argv[1]

	if res_type == "cr":
		custom_resource = sys.argv[2]
		custom_resource_instance = sys.argv[3]
		namespace = sys.argv[4]
		crLogs.get_metrics_cr(custom_resource, custom_resource_instance, namespace)
	
	if res_type == "account":
		creator_account = sys.argv[2]
		crLogs.get_metrics_creator_account(creator_account)

	if res_type == "service":
		service_name = sys.argv[2]
		namespace = sys.argv[3]
		crLogs.get_metrics_service(service_name, namespace)