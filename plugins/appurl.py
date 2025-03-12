import subprocess
import sys
import json
import yaml
import platform
import os
from crmetrics import CRBase

class AppURLFinder(CRBase):

    def _get_node_ips(self, kubeconfig):
        node_ip_types = []
        node_ips = []
        node_list = []

        # get all the nodes
        get_nodes = "kubectl get nodes " + kubeconfig
        out, err = self.run_command(get_nodes)
        for line in out.split("\n"):
            line1 = ' '.join(line.split())
            if 'NAME' not in line1 and line1 != '':
                parts = line1.split(" ")
                nodename = parts[0].strip()
                node_list.append(nodename)

        # for each node, build list of external and internal ips
        for node in node_list:
            describe_node = "kubectl describe node " + node + " " + kubeconfig
            out1, err1 = self.run_command(describe_node)
            ip_types = {}
            for line in out1.split("\n"):
                line1 = ' '.join(line.split())
                if 'ExternalIP' in line1:
                    parts = line1.split(":")
                    nodeIP = parts[1].strip()
                    ip_types["external"] = nodeIP
                if 'InternalIP' in line1:
                    parts = line1.split(":")
                    nodeIP = parts[1].strip()
                    ip_types["internal"] = nodeIP
            node_ip_types.append(ip_types)

        # prefer external ip; otherwise return internal ip
        for item in node_ip_types:
            if "external" in item:
                node_ips.append(item["external"])
            elif "internal" in item:
                node_ips.append(item["internal"])

        return node_ips


    def get_service_endpoints(self, kind, service_instance_name, kubeconfig):
        endpoints = []
        service_name_in_annotation = self.get_service_name_from_ns(service_instance_name, kubeconfig)

        if service_name_in_annotation.lower() != kind.lower():
            print("Instance does not belong to the Service:" + kind)
            sys.exit(1)

        ingress_cmd = "kubectl get ingress -n " + service_instance_name + " " + kubeconfig

        out, err = self.run_command(ingress_cmd)
        if "No resources found" not in err:
            for line in out.split("\n"):
                line1 = ' '.join(line.split())
                if 'NAME' not in line1 and line1 != '':
                    parts = line1.split(" ")
                    hostname = parts[2].strip()
                    protocol = "https://"
                    endpoint = protocol + hostname
                    endpoints.append(endpoint)
        else:
            service_cmd = "kubectl get service -n " + service_instance_name + " " + kubeconfig
            out1, err1 = self.run_command(service_cmd)
            for line in out1.split("\n"):
                line1 = ' '.join(line.split())
                endpoint = ''
                if 'NAME' not in line1 and line1 != '':
                    parts = line1.split(" ")
                    port_parts = parts[4].split(",")
                    for protocol in port_parts:
                        proto_ports = protocol.split(":")
                        if len(proto_ports) == 2:
                            proto_port = proto_ports[0].strip()
                            app_port = (proto_ports[1].strip()).split("/")[0].strip()
                            if parts[1] == 'LoadBalancer':
                                if proto_port == '80':
                                    endpoint = "http://" + parts[2].strip() + ":" + app_port
                                if proto_port == '443':
                                    endpoint = "https://" + parts[2].strip() + ":" + app_port
                            if parts[1] == 'NodePort':
                                ip_address_list = self._get_node_ips(kubeconfig)
                                for ipaddr in ip_address_list:
                                    if proto_port == '80':
                                        endpoint = "http://" + ipaddr.strip() + ":" + app_port
                                    if proto_port == '443':
                                        endpoint = "https://" + ipaddr.strip() + ":" + app_port
                                    endpoints.append(endpoint)

        return endpoints


if __name__ == '__main__':
    appURLFinder = AppURLFinder()
    kind = sys.argv[1]
    instance = sys.argv[2]
    kubeconfig = sys.argv[3]
    endpoints = appURLFinder.get_service_endpoints(kind, instance, kubeconfig)
    print(str(endpoints))
