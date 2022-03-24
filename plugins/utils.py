def get_pods(resources):
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

def get_resources(resources):
	res_list = []
	for resource in resources:
		#print(resource)
		present = False
		for r in res_list:
			if r['Name'] == resource['Name'] and r['Namespace'] == resource['Namespace']:
				present = True
				break
		if not present:
			res_list.append(resource)
	#print(pod_list)
	return res_list

