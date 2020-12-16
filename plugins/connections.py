import sys
import json
from graphviz import Digraph
from graphviz import Graph

class ConnectionsGraph(object):

	def draw(self, connections_json, output_folder):
		#print(connections_json)
		fp = open(output_folder + "/" + connections_json, "r")
		json_data = fp.read()
		json_output = json.loads(json_data)

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
					dot.edge(fqpeername, fqnodename, color=color, label=relationshipType)
					dot.edge

		# Create edges
		#dot.edges(['AB', 'AL'])
		#dot.edge('B', 'L', constraint='false')

		filename = connections_json + ".gv"
		dot.render(filename, view=False)
		#print("Output available in " + filename + "." + opformat)

		#fp1 = open(output_folder + "/abc.txt", "w")
		#fp1.write(connections_json)
		#fp1.close()

if __name__ == '__main__':
	graph = ConnectionsGraph()

	#print("Inside connections.py")
	connections_json = sys.argv[1]
	output_folder = sys.argv[2]
	#print("Connections_json:"+ connections_json)
	#print("Output folder:" + output_folder)

	graph.draw(connections_json, output_folder)

