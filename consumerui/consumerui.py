import os
import time
import subprocess
import json
import re

from flask import request
from flask import Flask, render_template

application = Flask(__name__)
app = application


def run_command(cmd):
	print(cmd)
	proc_out = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
	out = proc_out[0]
	err = proc_out[1]
	out = out.decode('utf-8')
	err = err.decode('utf-8')
	return out, err

def get_metrics(service, res, namespace):
	cpu = 0
	memory = 0
	storage = 0
	nwTransmitBytes = 0
	nwReceiveBytes = 0
	cmd = 'kubectl metrics ' + service + ' ' + res + ' ' + namespace + ' -o json'
	out, err = run_command(cmd)
	if out != '':
		json_output = ''
		try:
			json_output = json.loads(out)
		except Exception as e:
			print(e)
		print(json_output)

		#{"networkReceiveBytes": "185065618.0 bytes", "accountidentity": "", "storage": "0 Gi", "networkTransmitBytes": "7157799.0 bytes", "subresources": "-", "memory": "75.5078125 Mi", "nodes": "1", "pods": "1", "cpu": "48.820286 m", "containers": "6"}
		cpu = json_output["cpu"].split(" ")[0].strip()
		memory = json_output["memory"].split(" ")[0].strip()
		storage = json_output["storage"].split(" ")[0].strip()
		nwTransmitBytes = json_output["networkTransmitBytes"].split(" ")[0].strip()
		nwReceiveBytes = json_output["networkReceiveBytes"].split(" ")[0].strip()
	return cpu, memory, storage, nwTransmitBytes, nwReceiveBytes

def get_connections_op(resource, instance, namespace):
	cmd = 'kubectl connections ' + resource + ' ' + instance + ' ' + namespace + ' -o html -i Namespace:default,ServiceAccount:default -n label,specproperty,envvariable,annotation'
	out, err = run_command(cmd)
	data = ''
	if out != '' and err == '':
		if 'Output available in:' in out:
			parts = out.split(':')
			filepath = parts[1].strip()
			fp = open(filepath, "r")
			data = fp.read()
			fp.close()
	return data

def get_app_url(resource, instance, namespace):
	cmd = 'kubectl appurl ' + resource + ' ' + instance + ' ' + namespace 
	out, err = run_command(cmd)
	data = ''
	if out != '' and err == '':
		print(out)
		out = out.strip()
	return out

def get_logs(resource, instance, namespace):
	cmd = 'kubectl applogs ' + resource + ' ' + instance + ' ' + namespace 
	out, err = run_command(cmd)
	data = ''
	if out != '' and err == '':
		print(out)
		out = out.strip()
	return out

def get_total_resources(service):
	instance_dict = get_all_resources(service)
	instance_list = instance_dict[service]
	num_of_instances = len(instance_list)
	total_cpu = 0
	total_memory = 0
	total_storage = 0
	total_nw_ingress = 0
	total_nw_egress = 0

	for instance in instance_list:
		res = instance['name']
		namespace = instance['namespace']

		cpu, memory, storage, nwTransmitBytes, nwReceiveBytes = get_metrics(service, res, namespace)

		total_cpu = total_cpu + float(cpu)
		total_memory = total_memory + float(memory)
		total_storage = total_storage + float(storage)
		total_nw_ingress = total_nw_ingress + float(nwReceiveBytes)
		total_nw_egress = total_nw_egress + float(nwTransmitBytes)

	return num_of_instances, total_cpu, total_memory, total_storage, total_nw_ingress, total_nw_egress

def process_manpage_line(line):
	parts = line.split(":")
	part = parts[1].strip()
	return part

def get_input_fields(serviceName):
	cmd = 'kubectl man ' + serviceName
	out, err = run_command(cmd)

	kind = ""
	group = ""
	version = ""
	apiVersion = ""
	fieldList = []
        # invariant we want to maintain - fieldList should not contain namespace
	if out != "":
		lines = out.split("\n")
		specFields = False
		for line in lines:
			print(line)
			if "KIND:" in line:
				kind = process_manpage_line(line)
			if "GROUP:" in line:
				group = process_manpage_line(line)
			if "VERSION:" in line:
				version = process_manpage_line(line)
			if specFields:
				parts = line.split(":")
				print("LINE:" + line)
				if len(parts) >= 2:
					field = parts[0].strip()
					if field != '':
						fieldList.append(field)
			if "/values.yaml" in line:
				specFields = True

	apiVersion = group + "/" + version
	print("kind:" + kind)
	print("apiVersion:" + apiVersion)
	print("fieldList:" + str(fieldList))
	return kind, apiVersion, fieldList

def get_all_resources(resource):
	cmd = 'kubectl get ' + resource + " -A -o json"
	out, err = run_command(cmd)
	service_instance_out_list = []
	instances = {}
	if err != '':
		instances[resource] = service_instance_out_list
		return instances
	else:
		json_output = ''
		try:
			json_output = json.loads(out)
		except Exception as e:
			print(e)
		print(json_output)
		for instance in json_output['items']:
			out_dict = {}
			out_dict['name'] = instance['metadata']['name']
			if 'namespace' in instance['metadata']:
				out_dict['namespace'] = instance['metadata']['namespace']
			if 'status' in instance:
				if 'phase' in instance['status']:
					out_dict['running'] = instance['status']['phase']
			service_instance_out_list.append(out_dict)
		print(service_instance_out_list)
		instances[resource] = service_instance_out_list
		return instances

@app.route("/service/<service>/field_names", methods=['GET'])
def get_field_names(service):
	print(service)
	kind, apiVersion, fields = get_input_fields(service)

	# the name of the resource needs to be input as well.
        # We are not requiring namespace to be input since the consumer
        # does not really know what value should be provided to the namespace value.
	#if 'namespace' not in fields:
	#	fields.insert(0,"namespace")
	if 'name' not in fields:
		fields.insert(0,"name") 
	fields_dict = {}
	fields_dict["fields"] = fields
	return fields_dict

@app.route("/get_resource_manpage", methods=['GET'])
def get_resource_manpage():
	resource = request.args.get('resource')
	cmd = 'kubectl man ' + resource
	out, err = run_command(cmd)

	manPage = {}
	if err != "":
		manPage[resource] = err
		print("Error:" + err)
	elif out != "":
		manPage[resource] = out
		print("Man page:" + out)
	else:
		manPage["resource"] = "No information available."
	return manPage

@app.route("/service/create_instance", methods=['POST'])
def create_instance():
	print("Received request.")
	serviceName = request.form["serviceName"]
	print("Service Name:" + serviceName)
	kind, apiVersion, fields = get_input_fields(serviceName);

	fieldMap = {}
	for f in fields:
		fieldMap[f] = request.form[f]
	#apiVersion: platformapi.kubeplus/v1alpha1
	#kind: WordpressService 
	#metadata:
	#  name: abc-org-tenant1
	#spec:
	#  namespace: default 
	#  tenantName: tenant1
	#  nodeName: gke-abc-org-default-pool-d0114ae7-0dl9

	res = {}
	res["apiVersion"] = apiVersion
	res["kind"] = serviceName
	metadata = {}
	resName = request.form["name"]
	namespace = "default"
	metadata["name"] = resName
	res["metadata"] = metadata
	#if 'namespace' in fieldMap:
	#	namespace = fieldMap["namespace"]
	metadata["namespace"] = namespace
	spec = {}
	for k,v in fieldMap.items():
		spec[k] = v
	res["spec"] = spec

	fp = open("resource.json","w")
	fp.write(json.dumps(res))
	fp.close()

	cmd = "kubectl create -f ./resource.json "
	out, err = run_command(cmd)
	create_status = ""
	if err == "":
		create_status = "Resource " + resName + " created successfully."
	else:
		create_status = err

	resource_dict = get_all_resources(serviceName)
	resource_list = resource_dict[serviceName]

	template = render_template('welcome.html',
							resource_list=resource_list,
							service_name=serviceName,
							create_status=create_status,
							create_status_display="display:block;",
							form_display="display:block;",
							table_display="display:none;")
	fp = open("tmpl.html","w")
	fp.write(template)
	fp.close()

	return template

@app.route("/getAll", methods=['GET'])
def getAllResources():
	print("Received request.")
	resource = request.args.get('resource')
	instances = get_all_resources(resource)
	return instances

@app.route("/get_all_service_instances", methods=['POST'])
def get_all_service_instances():
	print("Received request.")
	service = request.form["service"]
	cmd = 'kubectl get ' + service + " -A -o json"
	out, err = run_command(cmd)
	if err != '':
		return render_template('consumeruiack.html',get_all_error_message=err)
	else:
		json_output = ''
		try:
			json_output = json.loads(out)
		except Exception as e:
			print(e)
			return render_template('consumeruiack.html',get_all_error_message=e)

	service_instance_out_list = []
	for instance in json_output['items']:
		out_dict = {}
		out_dict['name'] = instance['metadata']['name']
		if 'namespace' in instance['metadata']:
			out_dict['namespace'] = instance['metadata']['namespace']
		if 'status' in instance:
			if 'phase' in instance['status']:
				out_dict['running'] = instance['status']['phase']
		service_instance_out_list.append(out_dict)
	print(service_instance_out_list)
	if len(service_instance_out_list) > 0:
		return render_template('consumeruiack.html',get_all_error_message='',table_header='true',service_instance_list=service_instance_out_list)
	else:
		return render_template('consumeruiack.html',get_all_error_message='',table_header='false',no_data='true')


@app.route("/service/instance_logs", methods=['GET'])
def get_instance_logs():
	resource = request.args.get('resource').strip()
	instance = request.args.get('instance').strip()
	namespace = request.args.get('namespace').strip()

	logs = get_logs(resource, instance, namespace)
	instance_logs = {}
	instance_logs['logs'] = logs
	return instance_logs

@app.route("/service/instance_data", methods=['GET'])
def get_instance_data():
	resource = request.args.get('resource').strip()
	instance = request.args.get('instance').strip()
	namespace = request.args.get('namespace').strip()

	cpu, memory, storage, nwTransmitBytes, nwReceiveBytes = get_metrics(resource, instance, namespace)

	instance_data = {}
	instance_data['cpu'] = cpu
	instance_data['memory'] = memory
	instance_data['storage'] = storage
	instance_data['nw_egress'] = nwTransmitBytes
	instance_data['nw_ingress'] = nwReceiveBytes

	#connections_op = get_connections_op(resource, instance, namespace)
	#instance_data['connections_op'] = connections_op

	app_url = get_app_url(resource, instance, namespace)
	print("APP URL:" + app_url)
	instance_data['app_url'] = app_url.strip()

	logs = get_logs(resource, instance, namespace)
	instance_data['logs'] = logs
	return instance_data

@app.route("/get_instance_status", methods=['POST'])
def get_instance_status():
	print("Received request.")
	service = request.form["service"]
	instance = request.form["instance"]
	namespace = request.form["namespace"]
	cmd = 'kubectl get ' + service + " " + instance + " -n " + namespace
	out, err = run_command(cmd)
	lines = []
	if err != '':
		lines = err.split("\n")
	else:
		lines = out.split("\n")
	return render_template('consumeruiack.html',get_all_error_message='',instance_status=lines)

@app.route("/create_service_instance", methods=['POST'])
def create_service_instance():
	print("Received request.")
	print(request.form)
	service_instance = request.form["service_instance"]
	fp = open("/root/service_instance.yaml","w")
	fp.write(service_instance)
	fp.close()
	cmd = 'kubectl create -f /root/service_instance.yaml '
	out, err = run_command(cmd)
	instance_creation_status = out
	if err != '':
		instance_creation_status = err
	return render_template('consumeruiack.html',get_all_error_message='',service_instance=service_instance,instance_creation_status=instance_creation_status)

@app.route("/register_kubeconfig", methods=['POST'])
def register_kubeconfig():
	print("Received request.")
	print(request.form)
	kubeconfig = request.form["kubeconfig"]
	fp = open("/root/.kube/config","w")
	fp.write(kubeconfig)
	fp.close()
	return render_template('consumeruiack.html',get_all_error_message='',kubeconfig=kubeconfig,kubeconfig_received="Input received")

#@app.route("/")
def index1():
    print("Inside hello")
    print("Printing available environment variables")
    print(os.environ)
    return render_template('consumeruiack.html',get_all_error_message='',instance_status='')

@app.route("/service/<service>/namespace/<namespace>/instance/<instance>")
def get_resource_info(service, namespace, instance):
	cmd = ''
	pass

@app.route("/")
def index():
	return render_template('welcome.html')

@app.route("/service/<service>")
def service_index(service):
	print(service)
	num_of_instances, total_cpu, total_memory, total_storage, total_nw_ingress, total_nw_egress = get_total_resources(service)
	resource_dict = get_all_resources(service)
	resource_list = resource_dict[service]
	consumption_string = "Consumption based metrics for " + service + " (aggregate of all instances) "
	num_of_instances_string = "Number of instances: " + str(len(resource_list))
	return render_template('welcome.html',
							consumption_string=consumption_string,
							service_name=service,num_of_instances=num_of_instances,
							total_cpu=total_cpu,
							total_memory=total_memory,
							total_storage=total_storage,
							total_nw_ingress=total_nw_ingress,
							total_nw_egress=total_nw_egress,
							resource_list=resource_list,
							num_of_instances_string=num_of_instances_string)

def get_kubeplus_namespace():
	cmd = " kubectl get deployments -A "
	#print(cmd)
	out = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
	#print(out)
	out = out.decode('utf-8')
	kubeplusNamespace = ''
	for line in out.split("\n"):
		line1 = re.sub(' +', ' ', line)
		parts = line1.split()
		if parts[1] == 'kubeplus-deployment':
			kubeplusNamespace = parts[0]
			break
	return kubeplusNamespace

def download_consumer_kubeconfig():
	kubeplusNS = get_kubeplus_namespace()
	print("KubePlus NS:" + kubeplusNS)
	cmd = "kubectl get configmaps kubeplus-saas-consumer-kubeconfig -n " + kubeplusNS + " -o jsonpath=\"{.data.kubeplus-saas-consumer\\.json}\""
	out, err = run_command(cmd)
	if err == '':
		consumer_kubeconfig = out.strip()
		print("Consumer kubeconfig")
		print(consumer_kubeconfig)
		kubeconfig_path = "/root/.kube/"
		if os.path.exists(kubeconfig_path):
			fp = open(kubeconfig_path + "/config","w")
			fp.write(consumer_kubeconfig)
			fp.close()


if __name__ == "__main__":
    app.debug = True

    download_consumer_kubeconfig()

    app.run(host='0.0.0.0')
