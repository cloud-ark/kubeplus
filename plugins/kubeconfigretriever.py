import sys
import json
import subprocess
import os
from crmetrics import CRBase


class KubeconfigRetriever(CRBase):

	def retrieve_kubeconfig(self, serverURL, kubeconfigFor, kubeconfig):

		kubeplusNS = self.get_kubeplus_namespace(kubeconfig)
		if kubeconfigFor == 'provider':
			cmd = "kubectl get configmaps kubeplus-saas-provider -n " + kubeplusNS + r" -o jsonpath='{.data.kubeplus-saas-provider\.json}'"
		if kubeconfigFor == 'consumer':
			cmd = "kubectl get configmaps kubeplus-saas-consumer-kubeconfig -n " + kubeplusNS + r" -o jsonpath='{.data.kubeplus-saas-consumer\.json}'"

		cmd = cmd + " --kubeconfig=" + kubeconfig
		out = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
		out = out.decode('utf-8')
		json_output = {}
		try:
			json_output = json.loads(out)
			if isinstance(json_output, str):
				json_output = json.loads(json_output)
		except Exception as e:
			print(e)
			print("KubePlus might not be ready yet. Try again once the KubePlus Pod is running.")
		if isinstance(json_output, dict):
			if serverURL != '-1':
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
