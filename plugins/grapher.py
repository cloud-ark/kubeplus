import sys
import json
import subprocess
import sys
import os
from graphviz import Digraph
from graphviz import Graph

class ConnectionsGraph(object):

	def draw(self, connections_json, output_folder, relsToHide):
		#print(connections_json)
		cmd = "ls -ltr /root/"
		out = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
		#print(out)
		fp = open(output_folder + "/" + connections_json, "r")
		json_data = fp.read()
		json_output = json.loads(json_data)
		#print(json_output)

		nodemap = {}
		for n in json_output:
			level = n['Level']
			if level in nodemap.keys():
				nodelist = nodemap[level]
			else:
				nodelist = []
			nodelist.append(n)
			nodemap[level] = nodelist

		#print(nodemap)
		opformat = 'png'
		dot = Graph(comment='Connections Graph', format=opformat)
#		dot.node('A', 'King Shivaji')
#		dot.node('B', 'Sir Bedevere the Wise')
#		dot.node('L', 'Sir Lancelot the Brave')

		relsToHideList1 = relsToHide.split(",")
		relsToHideList = []
		for rel in relsToHideList1:
			relsToHideList.append(rel.strip())
		#print(relsToHideList)
		# Create Nodes
		for level, nodelist in nodemap.items():
			for n in nodelist:
				fqnodename = n['Kind'] + " " + n['Name']
				fqpeername = n['PeerKind'] + " " + n['PeerName']
				#print(fqnodename + " " + fqpeername)
				if n['Kind'] == 'Pod':
					dot.node(fqnodename, fqnodename, shape='box', style='filled', color='lightcyan1')
				else:
					dot.node(fqnodename, fqnodename, shape='box', style='filled', color='snow2')
				if level > 0:
					color = 'gray0'
					relationshipType = n['RelationType']
					relationshipDetails = n['RelationDetails']
					relationInfo = relationshipType
					if relationshipDetails != '' and relationshipType not in relsToHideList:
						relationInfo = relationInfo + " (" + relationshipDetails + ")"
					if relationshipType == 'specproperty':
						color = 'crimson'
					if relationshipType == 'label':
						color = 'darkgreen'
					if relationshipType == 'envvariable':
						color = 'gold4'
					if relationshipType == 'annotation':
						color = 'indigo'
					if relationshipType == 'owner reference':
						color = 'blue'
					dot.edge(fqpeername, fqnodename, color=color, label=relationInfo)

		# Create edges
		#dot.edges(['AB', 'AL'])
		#dot.edge('B', 'L', constraint='false')
		#print(dot.source)

		filename = connections_json + ".gv"
		#rendered_file_path = dot.render('/root/' + filename, view=False)
		rendered_file_path = dot.render(output_folder + filename, view=False)
		#print("FILENAME:" + filename)
		#print("Rendered file path:" + rendered_file_path)
		#print("Output available in " + filename + "." + opformat)

		#fp1 = open(output_folder + "/abc.txt", "w")
		#fp1.write(connections_json)
		#fp1.close()

if __name__ == '__main__':
	graph = ConnectionsGraph()

	#print("Inside connections.py")
	connections_json = sys.argv[1]
	output_folder = sys.argv[2]
	if len(sys.argv) == 4:
		relsToHide = sys.argv[3]
	else:
		relsToHide = ""
	#print("Connections_json:"+ connections_json)
	#print("Output folder:" + output_folder)
	#print(relsToHide)

	graph.draw(connections_json, output_folder, relsToHide)

