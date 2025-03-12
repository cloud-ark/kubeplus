import sys
import json
import subprocess
import os
from crmetrics import CRBase


class KubeconfigRetriever(CRBase):

	def retrieve_kubeconfig(self, serverURL, kubeconfigFor, kubeconfig):

		kubeplusNS = self.get_kubeplus_namespace(kubeconfig)
		if kubeconfigFor == 'provider':
			cmd = "kubectl get configmaps kubeplus-saas-provider -n " + kubeplusNS + " -o jsonpath=\"{.data.kubeplus-saas-provider\.json}\""
		if kubeconfigFor == 'consumer':
			cmd = "kubectl get configmaps kubeplus-saas-consumer-kubeconfig -n " + kubeplusNS + " -o jsonpath=\"{.data.kubeplus-saas-consumer\.json}\""			

		#kubeconfigParts = kubeconfig.split("=")
		#kubeconfigPath = kubeconfigParts[1].strip()
		cmd = cmd + " --kubeconfig=" + kubeconfig
		out = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
		#print(out)
		out = out.decode('utf-8')
		json_output = {}
		try:
			json_output = json.loads(out)
		except Exception as e:
			print(e)
			print("KubePlus might not be ready yet. Try again once the KubePlus Pod is running.")
		if serverURL != '-1':
			#parts = serverURL.split("=")
			#sURL = parts[1].strip()
			#if sURL != '':
			json_output["clusters"][0]["cluster"]["server"] = serverURL
		try:
			pkubeconfig = json.dumps(json_output)
			print(pkubeconfig)
		except Exception as e:
			print(e)

if __name__ == '__main__':

	serverURL = sys.argv[1] # <api server url>
	kubeconfigFor = sys.argv[2]
	kubeconfigPath = sys.argv[3] # <complete kubeconfig path>
	kfgRetriever = KubeconfigRetriever()
	kfgRetriever.retrieve_kubeconfig(serverURL, kubeconfigFor, kubeconfigPath)
