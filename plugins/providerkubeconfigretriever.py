import sys
import json
import subprocess
import os


class ProviderKubeconfigRetriever(object):

	def retrieve_kubeconfig(self, kubeplusNS, serverURL):

		cmd = "kubectl get configmaps kubeplus-saas-provider-kubeconfig -n " + kubeplusNS + " -o jsonpath=\"{.data.kubeplus-saas-provider\.json}\""
		out = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
		#print(out)
		out = out.decode('utf-8')
		json_output = json.loads(out)
		if serverURL != '-1':
			parts = serverURL.split("=")
			sURL = parts[1].strip()
			if sURL != '':
				json_output["clusters"][0]["cluster"]["server"] = sURL
		try:
			pkubeconfig = json.dumps(json_output)
			print(pkubeconfig)
		except Exception as e:
			print(e)

if __name__ == '__main__':

	kubeplusNS = sys.argv[1]
	serverURL = sys.argv[2] # --server=<api server url>
	providerKfgRetriever = ProviderKubeconfigRetriever()
	providerKfgRetriever.retrieve_kubeconfig(kubeplusNS, serverURL)
