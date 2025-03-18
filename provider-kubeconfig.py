import sys
import json
import subprocess
import sys
import os
import yaml
import time
import argparse

from logging.config import dictConfig

dictConfig({
    'version': 1,
    'formatters': {'default': {
        'format': '[%(asctime)s] %(levelname)s in %(module)s: %(message)s',
    }},
    'handlers': {
     'file.handler': {
            'class': 'logging.handlers.RotatingFileHandler',
            'filename': 'provider-kubeconfig.log',
            'maxBytes': 10000000,
            'backupCount': 5,
            'level': 'DEBUG',
        },
    },
    'root': {
        'level': 'INFO',
        'handlers': ['file.handler']
    }
})



def create_role_rolebinding(contents, name, kubeconfig):
    filePath = os.getcwd() + "/" + name
    fp = open(filePath, "w")
    #json_content = json.dumps(contents)
    #fp.write(json_content)
    yaml_content = yaml.dump(contents)
    fp.write(yaml_content)
    fp.close()
    #print("---")
    #print(yaml_content)
    #print("---")
    cmd = " kubectl apply -f " + filePath + kubeconfig
    run_command(cmd)


def run_command(cmd):
    #print("Inside run_command")
    #print(cmd)
    cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
    out = cmdOut[0].decode('utf-8')
    err = cmdOut[1].decode('utf-8')
    #print(out)
    #print("---")
    #print(err)
    return out, err


class KubeconfigGenerator(object):

        def run_command(self, cmd):
                #print("Inside run_command")
                #print(cmd)
                cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
                out = cmdOut[0]
                err = cmdOut[1]
                #print(out)
                #print("---")
                #print(err)
                return out, err

        def _create_kubecfg_file(self, sa, namespace, filename, token, ca, server, kubeconfig, cluster_name=None):
                #print("Creating kubecfg file")
                top_level_dict = {}
                top_level_dict["apiVersion"] = "v1"
                top_level_dict["kind"] = "Config"

                contextName = cluster_name if cluster_name else sa

                usersList = []
                usertoken = {}
                usertoken["token"] = token
                userInfo = {}
                userInfo["name"] = sa
                userInfo["user"] = usertoken
                usersList.append(userInfo)
                top_level_dict["users"] = usersList

                clustersList = []
                cluster_details = {}
                cluster_details["server"] = server
                
                # TODO: Use the certificate authority to perform tls 
                # cluster_details["certificate-authority-data"] = ca
                cluster_details["insecure-skip-tls-verify"] = True

                clusterInfo = {}
                clusterInfo["cluster"] = cluster_details
                clusterInfo["name"] = cluster_name if cluster_name else sa
                clustersList.append(clusterInfo)
                top_level_dict["clusters"] = clustersList

                context_details = {}
                context_details["cluster"] = cluster_name if cluster_name else sa
                context_details["user"] = sa
                context_details["namespace"] = namespace
                contextInfo = {}
                contextInfo["context"] = context_details
                contextInfo["name"] = contextName
                contextList = []
                contextList.append(contextInfo)
                top_level_dict["contexts"] = contextList

                top_level_dict["current-context"] = contextName

                json_file = json.dumps(top_level_dict)
                #print("kubecfg file:" + json_file)

                fp = open(os.getcwd() + "/" + filename, "w")
                fp.write(json_file)
                fp.close()

                configmapName = sa
                created = False 
                while not created:        
                        cmd = "kubectl create configmap " + configmapName + " -n " + namespace + " --from-file=" + os.getcwd() + "/" + filename + kubeconfig
                        self.run_command(cmd)
                        get_cmd = "kubectl get configmap " + configmapName + " -n "  + namespace + kubeconfig
                        output, err = self.run_command(get_cmd)
                        output = output.decode('utf-8')    
                        if 'Error from server (NotFound)' in output:
                                time.sleep(2)
                                print("Trying again..")
                        else:
                                created = True


        def _apply_consumer_rbac(self, sa, namespace, kubeconfig):
                role = {}
                role["apiVersion"] = "rbac.authorization.k8s.io/v1"
                role["kind"] = "ClusterRole"
                metadata = {}
                metadata["name"] = sa
                metadata["namespace"] = namespace
                role["metadata"] = metadata

                # all resources
                all_resources = []

                # Read all resources
                ruleGroup1 = {}
                apiGroup1 = ["*",""]
                resourceGroup1 = ["*"]
                verbsGroup1 = ["get","watch","list"]
                ruleGroup1["apiGroups"] = apiGroup1
                ruleGroup1["resources"] = resourceGroup1
                ruleGroup1["verbs"] = verbsGroup1
                all_resources.extend(resourceGroup1)

                ruleGroup8 = {}
                apiGroup8 = ["apps"]
                resourceGroup8 = ["deployments","daemonsets","deployments/rollback","deployments/scale","replicasets","replicasets/scale","statefulsets","statefulsets/scale"]
                verbsGroup8 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup8["apiGroups"] = apiGroup8
                ruleGroup8["resources"] = resourceGroup8
                ruleGroup8["verbs"] = verbsGroup8
                all_resources.extend(resourceGroup8)

                # Impersonate users, groups, serviceaccounts
                ruleGroup9 = {}
                apiGroup9 = [""]
                resourceGroup9 = ["users","groups","serviceaccounts"]
                verbsGroup9 = ["impersonate"]
                ruleGroup9["apiGroups"] = apiGroup9
                ruleGroup9["resources"] = resourceGroup9
                ruleGroup9["verbs"] = verbsGroup9
                all_resources.extend(resourceGroup9)

                # Pod/portforward to open consumerui
                ruleGroup10 = {}
                apiGroup10 = [""]
                resourceGroup10 = ["pods/portforward"]
                verbsGroup10 = ["create","get"]
                ruleGroup10["apiGroups"] = apiGroup10
                ruleGroup10["resources"] = resourceGroup10
                ruleGroup10["verbs"] = verbsGroup10
                all_resources.extend(resourceGroup10)

                ruleList = []
                ruleList.append(ruleGroup1)
                ruleList.append(ruleGroup9)
                ruleList.append(ruleGroup10)
                ruleList.append(ruleGroup8)
                role["rules"] = ruleList

                roleName = sa + "-role-impersonate.yaml"
                create_role_rolebinding(role, roleName, kubeconfig)

                roleBinding = {}
                roleBinding["apiVersion"] = "rbac.authorization.k8s.io/v1"
                roleBinding["kind"] = "ClusterRoleBinding"
                metadata = {}
                metadata["name"] = sa
                metadata["namespace"] = namespace
                roleBinding["metadata"] = metadata

                subject = {}
                subject["kind"] = "ServiceAccount"
                subject["name"] = sa
                subject["apiGroup"] = ""
                subject["namespace"] = namespace
                subjectList = []
                subjectList.append(subject)
                roleBinding["subjects"] = subjectList

                roleRef = {}
                roleRef["kind"] = "ClusterRole"
                roleRef["name"] = sa
                roleRef["apiGroup"] = "rbac.authorization.k8s.io"
                roleBinding["roleRef"] = roleRef

                roleBindingName = sa + "-rolebinding-impersonate.yaml"
                create_role_rolebinding(roleBinding, roleBindingName, kubeconfig)

               # create configmap to store all resources
                cfg_map_filename = sa + "-perms.txt"
                fp = open(cfg_map_filename, "w")
                all_resources.sort()
                all_resources_uniq = []
                [all_resources_uniq.append(x) for x in all_resources if x not in all_resources_uniq]
                fp.write(str(all_resources_uniq))
                fp.close()
                cfg_map_name = sa + "-perms"
                cmd = "kubectl create configmap " + cfg_map_name + " -n " + namespace  + " --from-file=" + cfg_map_filename
                self.run_command(cmd)


        def _apply_provider_rbac(self, sa, namespace, kubeconfig):
                role = {}
                role["apiVersion"] = "rbac.authorization.k8s.io/v1"
                role["kind"] = "ClusterRole"
                metadata = {}
                metadata["name"] = sa
                metadata["namespace"] = namespace
                role["metadata"] = metadata

                # all resources
                all_resources = []

                # Read all resources
                ruleGroup1 = {}
                apiGroup1 = ["*",""]
                resourceGroup1 = ["*"]
                verbsGroup1 = ["get","watch","list"]
                ruleGroup1["apiGroups"] = apiGroup1
                ruleGroup1["resources"] = resourceGroup1
                ruleGroup1["verbs"] = verbsGroup1

                # CRUD on resourcecompositions et. al.
                ruleGroup2 = {}
                apiGroup2 = ["workflows.kubeplus"]
                resourceGroup2 = ["resourcecompositions","resourcemonitors","resourcepolicies","resourceevents"]
                verbsGroup2 = ["get","watch","list","create","delete","update","patch"]
                ruleGroup2["apiGroups"] = apiGroup2
                ruleGroup2["resources"] = resourceGroup2
                ruleGroup2["verbs"] = verbsGroup2
                all_resources.extend(resourceGroup2)

                # CRUD on clusterroles and clusterrolebindings
                ruleGroup3 = {}
                apiGroup3 = ["rbac.authorization.k8s.io"]
                resourceGroup3 = ["clusterroles","clusterrolebindings","roles","rolebindings"]
                verbsGroup3 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup3["apiGroups"] = apiGroup3
                ruleGroup3["resources"] = resourceGroup3
                ruleGroup3["verbs"] = verbsGroup3
                all_resources.extend(resourceGroup3)

                # CRUD on Port forward
                ruleGroup4 = {}
                apiGroup4 = [""]
                resourceGroup4 = ["pods/portforward"]
                verbsGroup4 = ["get","watch","list","create","delete","update","patch"]
                ruleGroup4["apiGroups"] = apiGroup4
                ruleGroup4["resources"] = resourceGroup4
                ruleGroup4["verbs"] = verbsGroup4
                all_resources.extend(resourceGroup4)

                # CRUD on platformapi.kubeplus
                ruleGroup5 = {}
                apiGroup5 = ["platformapi.kubeplus"]
                resourceGroup5 = ["*"]
                verbsGroup5 = ["get","watch","list","create","delete","update","patch"]
                ruleGroup5["apiGroups"] = apiGroup5
                ruleGroup5["resources"] = resourceGroup5
                ruleGroup5["verbs"] = verbsGroup5

                # CRUD on secrets, serviceaccounts, configmaps
                ruleGroup6 = {}
                apiGroup6 = [""]
                resourceGroup6 = ["secrets", "serviceaccounts", "configmaps","events","persistentvolumeclaims","serviceaccounts/token","services","services/proxy","endpoints"]
                verbsGroup6 = ["get","watch","list","create","delete","update","patch", "deletecollection"]
                ruleGroup6["apiGroups"] = apiGroup6
                ruleGroup6["resources"] = resourceGroup6
                ruleGroup6["verbs"] = verbsGroup6
                all_resources.extend(resourceGroup6)

                # CRUD on namespaces
                ruleGroup7 = {}
                apiGroup7 = [""]
                resourceGroup7 = ["namespaces"]
                verbsGroup7 = ["get","watch","list","create","delete","update","patch"]
                ruleGroup7["apiGroups"] = apiGroup7
                ruleGroup7["resources"] = resourceGroup7
                ruleGroup7["verbs"] = verbsGroup7
                all_resources.extend(resourceGroup7)

                # CRUD on Deployments
                ruleGroup8 = {}
                apiGroup8 = ["apps"]
                resourceGroup8 = ["deployments","daemonsets","deployments/rollback","deployments/scale","replicasets","replicasets/scale","statefulsets","statefulsets/scale"]
                verbsGroup8 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup8["apiGroups"] = apiGroup8
                ruleGroup8["resources"] = resourceGroup8
                ruleGroup8["verbs"] = verbsGroup8
                all_resources.extend(resourceGroup8)

                # Impersonate users, groups, serviceaccounts
                ruleGroup9 = {}
                apiGroup9 = [""]
                resourceGroup9 = ["users","groups","serviceaccounts"]
                verbsGroup9 = ["impersonate"]
                ruleGroup9["apiGroups"] = apiGroup9
                ruleGroup9["resources"] = resourceGroup9
                ruleGroup9["verbs"] = verbsGroup9
                all_resources.extend(resourceGroup9)

                # Exec into the Pods and others in the "" apiGroup
                ruleGroup10 = {}
                apiGroup10 = [""]
                resourceGroup10 = ["pods","pods/attach","pods/exec","pods/portforward","pods/proxy","pods/eviction","replicationcontrollers","replicationcontrollers/scale"]
                verbsGroup10 = ["get","list","create","update","delete","watch","patch","deletecollection"]
                ruleGroup10["apiGroups"] = apiGroup10
                ruleGroup10["resources"] = resourceGroup10
                ruleGroup10["verbs"] = verbsGroup10
                all_resources.extend(resourceGroup10)

                # AdmissionRegistration
                ruleGroup11 = {}
                apiGroup11 = ["admissionregistration.k8s.io"]
                resourceGroup11 = ["mutatingwebhookconfigurations"]
                verbsGroup11 = ["get","create","delete","update"]
                ruleGroup11["apiGroups"] = apiGroup11
                ruleGroup11["resources"] = resourceGroup11
                ruleGroup11["verbs"] = verbsGroup11
                all_resources.extend(resourceGroup11)

                # APIExtension
                ruleGroup12 = {}
                apiGroup12 = ["apiextensions.k8s.io"]
                resourceGroup12 = ["customresourcedefinitions"]
                verbsGroup12 = ["get","create","delete","update", "patch"]
                ruleGroup12["apiGroups"] = apiGroup12
                ruleGroup12["resources"] = resourceGroup12
                ruleGroup12["verbs"] = verbsGroup12
                all_resources.extend(resourceGroup12)

                # Certificates
                ruleGroup13 = {}
                apiGroup13 = ["certificates.k8s.io"]
                resourceGroup13 = ["signers"]
                resourceNames13 = ["kubernetes.io/legacy-unknown","kubernetes.io/kubelet-serving","kubernetes.io/kube-apiserver-client","cloudark.io/kubeplus"]
                verbsGroup13 = ["get","create","delete","update", "patch", "approve"]
                ruleGroup13["apiGroups"] = apiGroup13
                ruleGroup13["resources"] = resourceGroup13
                ruleGroup13["resourceNames"] = resourceNames13
                ruleGroup13["verbs"] = verbsGroup13
                all_resources.extend(resourceGroup13)

                # Read all
                ruleGroup14 = {}
                apiGroup14 = ["*"]
                resourceGroup14 = ["*"]
                verbsGroup14 = ["get"]
                ruleGroup14["apiGroups"] = apiGroup14
                ruleGroup14["resources"] = resourceGroup14
                ruleGroup14["verbs"] = verbsGroup14

                ruleGroup15 = {}
                apiGroup15 = ["certificates.k8s.io"]
                resourceGroup15 = ["certificatesigningrequests", "certificatesigningrequests/approval"]
                verbsGroup15 = ["create","delete","update", "patch"]
                ruleGroup15["apiGroups"] = apiGroup15
                ruleGroup15["resources"] = resourceGroup15
                ruleGroup15["verbs"] = verbsGroup15
                all_resources.extend(resourceGroup15)

                ruleGroup16 = {}
                apiGroup16 = ["extensions"]
                resourceGroup16 = ["deployments","daemonsets","deployments/rollback","deployments/scale","replicasets","replicasets/scale","replicationcontrollers/scale","ingresses","networkpolicies"]
                verbsGroup16 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup16["apiGroups"] = apiGroup16
                ruleGroup16["resources"] = resourceGroup16
                ruleGroup16["verbs"] = verbsGroup16
                all_resources.extend(resourceGroup16)

                ruleGroup17 = {}
                apiGroup17 = ["networking.k8s.io"]
                resourceGroup17 = ["ingresses","networkpolicies"]
                verbsGroup17 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup17["apiGroups"] = apiGroup17
                ruleGroup17["resources"] = resourceGroup17
                ruleGroup17["verbs"] = verbsGroup17
                all_resources.extend(resourceGroup17)

                ruleGroup18 = {}
                apiGroup18 = ["authorization.k8s.io"]
                resourceGroup18 = ["localsubjectaccessreviews"]
                verbsGroup18 = ["create"]
                ruleGroup18["apiGroups"] = apiGroup18
                ruleGroup18["resources"] = resourceGroup18
                ruleGroup18["verbs"] = verbsGroup18
                all_resources.extend(resourceGroup18)

                ruleGroup19 = {}
                apiGroup19 = ["autoscaling"]
                resourceGroup19 = ["horizontalpodautoscalers"]
                verbsGroup19 = ["create", "delete", "deletecollection", "patch", "update"]
                ruleGroup19["apiGroups"] = apiGroup19
                ruleGroup19["resources"] = resourceGroup19
                ruleGroup19["verbs"] = verbsGroup19
                all_resources.extend(resourceGroup19)

                ruleGroup20 = {}
                apiGroup20 = ["batch"]
                resourceGroup20 = ["cronjobs","jobs"]
                verbsGroup20 = ["create", "delete", "deletecollection", "patch", "update"]
                ruleGroup20["apiGroups"] = apiGroup20
                ruleGroup20["resources"] = resourceGroup20
                ruleGroup20["verbs"] = verbsGroup20
                all_resources.extend(resourceGroup20)

                ruleGroup21 = {}
                apiGroup21 = ["policy"]
                resourceGroup21 = ["poddisruptionbudgets"]
                verbsGroup21 = ["create", "delete", "deletecollection", "patch", "update"]
                ruleGroup21["apiGroups"] = apiGroup21
                ruleGroup21["resources"] = resourceGroup21
                ruleGroup21["verbs"] = verbsGroup21
                all_resources.extend(resourceGroup21)

                ruleGroup22 = {}
                apiGroup22 = [""]
                resourceGroup22 = ["resourcequotas"]
                verbsGroup22 = ["create", "delete", "deletecollection", "patch", "update"]
                ruleGroup22["apiGroups"] = apiGroup22
                ruleGroup22["resources"] = resourceGroup22
                ruleGroup22["verbs"] = verbsGroup22
                all_resources.extend(resourceGroup22)

                # PersistentVolumes and PersistentVolumeClaims for charts storage in helmer container 
                ruleGroup23 = {}
                apiGroup23 = [""]
                resourceGroup23 = ["persistentvolumes", "persistentvolumeclaims"]
                verbsGroup23 = ["get", "watch", "list", "create", "delete", "update", "patch"]
                ruleGroup23["apiGroups"] = apiGroup23
                ruleGroup23["resources"] = resourceGroup23
                ruleGroup23["verbs"] = verbsGroup23
                all_resources.extend(resourceGroup23)

                

                ruleList = []
                ruleList.append(ruleGroup1)
                ruleList.append(ruleGroup2)
                ruleList.append(ruleGroup3)
                ruleList.append(ruleGroup4)
                ruleList.append(ruleGroup5)
                ruleList.append(ruleGroup6)
                ruleList.append(ruleGroup7)
                ruleList.append(ruleGroup8)
                ruleList.append(ruleGroup9)
                ruleList.append(ruleGroup10)
                ruleList.append(ruleGroup11)
                ruleList.append(ruleGroup12)
                ruleList.append(ruleGroup13)
                ruleList.append(ruleGroup14)
                ruleList.append(ruleGroup15)
                ruleList.append(ruleGroup16)
                ruleList.append(ruleGroup17)
                ruleList.append(ruleGroup18)
                ruleList.append(ruleGroup19)
                ruleList.append(ruleGroup20)
                ruleList.append(ruleGroup21)
                ruleList.append(ruleGroup22)
                ruleList.append(ruleGroup23)

                role["rules"] = ruleList

                roleName = sa + "-role.yaml"
                create_role_rolebinding(role, roleName, kubeconfig)

                roleBinding = {}
                roleBinding["apiVersion"] = "rbac.authorization.k8s.io/v1"
                roleBinding["kind"] = "ClusterRoleBinding"
                metadata = {}
                metadata["name"] = sa
                metadata["namespace"] = namespace
                roleBinding["metadata"] = metadata

                subject = {}
                subject["kind"] = "ServiceAccount"
                subject["name"] = sa
                subject["apiGroup"] = ""
                subject["namespace"] = namespace
                subjectList = []
                subjectList.append(subject)
                roleBinding["subjects"] = subjectList

                roleRef = {}
                roleRef["kind"] = "ClusterRole"
                roleRef["name"] = sa
                roleRef["apiGroup"] = "rbac.authorization.k8s.io"
                roleBinding["roleRef"] = roleRef

                roleBindingName = sa + "-rolebinding.yaml"
                create_role_rolebinding(roleBinding, roleBindingName, kubeconfig)

                # create configmap to store all resources
                cfg_map_filename = sa + "-perms.txt"
                fp = open(cfg_map_filename, "w")
                all_resources.sort()
                all_resources_uniq = []
                [all_resources_uniq.append(x) for x in all_resources if x not in all_resources_uniq]
                fp.write(str(all_resources_uniq))
                fp.close()
                cfg_map_name = sa + "-perms"
                cmd = "kubectl create configmap " + cfg_map_name + " -n " + namespace  + " --from-file=" + cfg_map_filename 
                self.run_command(cmd)

        def _update_rbac(self, permissionfile, sa, namespace, kubeconfig):
                role = {}
                role["apiVersion"] = "rbac.authorization.k8s.io/v1"
                role["kind"] = "ClusterRole"
                metadata = {}
                metadata["name"] = sa + "-update"
                metadata["namespace"] = namespace
                role["metadata"] = metadata
                
                ruleList = []
                ruleGroup = {}

                fp = open(permissionfile, "r")
                data = fp.read()
                perms_data = json.loads(data)
                perms = perms_data["perms"]
                new_resources = []
                for apiGroup, res_actions in perms.items():
                    for res in res_actions:
                        for resource, verbs in res.items():
                            #print(apiGroup + " " + resource + " " + str(verbs))
                            if resource not in new_resources:
                                new_resources.append(resource.strip())
                            ruleGroup = {}
                            if apiGroup == "non-apigroup":
                                if 'nonResourceURL' in resource:
                                    parts = resource.split("nonResourceURL::")
                                    nonRes = parts[0].strip()
                                    ruleGroup['nonResourceURLs'] = [nonRes]
                                    ruleGroup['verbs'] = verbs
                            else:
                                ruleGroup["apiGroups"] = [apiGroup]
                                ruleGroup["verbs"] = verbs
                                if 'resourceName' in resource:
                                    parts = resource.split("/resourceName::")
                                    resNameParent = parts[0].strip()
                                    resName = parts[1].strip()
                                    ruleGroup["resources"] = [resNameParent]
                                    ruleGroup["resourceNames"] = [resName]
                                else:
                                    ruleGroup["resources"] = [resource]

                    
                            ruleList.append(ruleGroup)

                role["rules"] = ruleList

                roleName = sa + "-update-role.yaml"
                create_role_rolebinding(role, roleName, kubeconfig)

                roleBinding = {}
                roleBinding["apiVersion"] = "rbac.authorization.k8s.io/v1"
                roleBinding["kind"] = "ClusterRoleBinding"
                metadata = {}
                metadata["name"] = sa + "-update"
                metadata["namespace"] = namespace
                roleBinding["metadata"] = metadata

                subject = {}
                subject["kind"] = "ServiceAccount"
                subject["name"] = sa
                subject["apiGroup"] = ""
                subject["namespace"] = namespace
                subjectList = []
                subjectList.append(subject)
                roleBinding["subjects"] = subjectList

                roleRef = {}
                roleRef["kind"] = "ClusterRole"
                roleRef["name"] = sa + "-update"
                roleRef["apiGroup"] = "rbac.authorization.k8s.io"
                roleBinding["roleRef"] = roleRef

                roleBindingName = sa + "-update-rolebinding.yaml"
                create_role_rolebinding(roleBinding, roleBindingName, kubeconfig)

                # Read configmap to get earlier permissions; delete it and create it with all new permissions:
                cfg_map_name = sa + "-perms"
                cfg_map_filename = sa + "-perms.txt"
                cmd = "kubectl get configmap " + cfg_map_name + " -o json -n " + namespace
                out1, err1 = self.run_command(cmd)
                #print("Original Perms Out:" + str(out1))
                #print("Perms Err:" + str(err1))
                kubeplus_perms = []
                if out1 != '':
                    json_op = json.loads(out1)
                    perms = json_op['data'][cfg_map_filename]
                    #print(perms)
                    k_perms = perms.split(",")
                    for p in k_perms:
                        p = p.replace("'","")
                        p = p.replace("[","")
                        p = p.replace("]","")
                        p = p.strip()
                        kubeplus_perms.append(p)

                new_resources.extend(kubeplus_perms)

                #print("New perms:" + str(new_resources))

                cmd = "kubectl delete configmap " + cfg_map_name + " -n " + namespace
                self.run_command(cmd)

                # create configmap to store all resources
                fp = open(cfg_map_filename, "w")
                new_resources.sort()
                new_resources_uniq = []
                [new_resources_uniq.append(x) for x in new_resources if x not in new_resources_uniq]
                fp.write(str(new_resources_uniq))
                fp.close()
                cmd = "kubectl create configmap " + cfg_map_name + " -n " + namespace  + " --from-file=" + cfg_map_filename 
                self.run_command(cmd)
    

        def _apply_rbac(self, sa, namespace, entity='', kubeconfig=''):
                if entity == 'provider':
                        self._apply_provider_rbac(sa, namespace, kubeconfig)
                if entity == 'consumer':
                        self._apply_consumer_rbac(sa, namespace, kubeconfig)

        def _create_secret(self, sa, namespace, kubeconfig):

                annotations = {}
                annotations['kubernetes.io/service-account.name'] = sa

                metadata = {}
                metadata['name'] = sa
                metadata['namespace'] = namespace
                metadata['annotations'] = annotations

                secret = {}
                secret['apiVersion'] = "v1"
                secret['kind'] = "Secret"
                secret['metadata'] = metadata
                secret['type'] = 'kubernetes.io/service-account-token'

                secretName = sa + "-secret.yaml"

                filePath = os.getcwd() + "/" + secretName
                fp = open(filePath, "w")
                yaml_content = yaml.dump(secret)
                fp.write(yaml_content)
                fp.close()
                #print("---")
                #print(yaml_content)
                #print("---")
                created = False
                count = 0
                while not created and count < 5:
                        cmd = " kubectl create -f " + filePath + kubeconfig
                        out, err = self.run_command(cmd)
                        if out != '':
                                out = out.decode('utf-8').strip()
                                #print(out)
                                if 'created' in out:
                                    created = True
                                else:
                                    time.sleep(2)
                                    count = count + 1
                #print("Create secret:" + out)
                if not created and count >= 5:
                    print(err)
                    sys.exit()
                return out

        def _extract_kubeconfig(self, sa, namespace, filename, serverip='', kubecfg='', cluster_name=None):
            #print("Extracting kubeconfig")
            secretName = sa
            tokenFound = False
            kubeconfig = kubecfg
            api_server_ip = serverip
            cmdprefix = ""
            while not tokenFound:
                cmd1 = " kubectl describe secret " + secretName + " -n " + namespace + kubeconfig
                cmdToRun = cmdprefix + " " + cmd1
                out1 = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
                out1 = out1.decode('utf-8')
                #print(out1)
                token = ''
                for line in out1.split("\n"):
                    if 'token' in line:
                        parts = line.split(":")
                        token = parts[1].strip()
                    if token != '':
                        tokenFound = True
                    else:
                        time.sleep(2)

            cmd1 = " kubectl get secret " + secretName + " -n " + namespace + " -o json " + kubeconfig
            cmdToRun = cmdprefix + " " + cmd1
            out1 = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
            out1 = out1.decode('utf-8')
            json_output1 = json.loads(out1)
            ca_cert = json_output1["data"]["ca.crt"].strip()
            #print("CA Cert:" + ca_cert)

            #cmd2 = " kubectl config view --minify -o json "
            server = ''
            if api_server_ip == '':
                cmd2 = "kubectl -n default get endpoints kubernetes " + kubeconfig + " | awk '{print $2}' | grep -v ENDPOINTS"
                cmdToRun = cmdprefix + " " + cmd2
                out2 = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]
                #print("Config view Minify:")
                #print(out2)
                out2 = out2.decode('utf-8')
                #json_output2 = json.loads(out2)
                #server = json_output2["clusters"][0]["cluster"]["server"].strip()
                server = out2.strip()
                server = "https://" + server
            else:
                if "https" not in api_server_ip:
                    server = "https://" + api_server_ip
                else:
                    server = api_server_ip
                    #print("Kube API Server:" + server)
            self._create_kubecfg_file(sa, namespace, filename, token, ca_cert, server, kubeconfig, cluster_name)


        def _generate_kubeconfig(self, sa, namespace, filename, api_server_ip='', kubeconfig='', cluster_name=None):
                cmdprefix = ""
                cmd = " kubectl create sa " + sa + " -n " + namespace + kubeconfig
                cmdToRun = cmdprefix + " " + cmd
                self.run_command(cmdToRun)

                #cmd = " kubectl get sa " + sa + " -n " + namespace + " -o json "
                #cmdToRun = cmdprefix + " " + cmd
                #out = subprocess.Popen(cmdToRun, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()[0]

                secretName = sa
                out = self._create_secret(secretName, namespace, kubeconfig)
                #print("Create secret:" + out)
                if 'secret/' + sa + ' created' in out:
                        #json_output = json.loads(out)
                        #secretName = json_output["secrets"][0]["name"]
                        #print("Secret Name:" + secretName)

                        # Moving from here
                        #print("Got secret token")
                        self._extract_kubeconfig(sa, namespace, filename, serverip=api_server_ip, kubecfg=kubeconfig, cluster_name=cluster_name)


if __name__ == '__main__':

        kubeconfigPath = os.getenv("HOME") + "/.kube/config"
        parser = argparse.ArgumentParser()
        parser.add_argument("action", help="command", choices=['create', 'delete', 'update', 'extract'])
        parser.add_argument("namespace", help="namespace in which KubePlus will be installed.")
        parser.add_argument("-k", "--kubeconfig", help='''This flag is used to specify the path
                of the kubeconfig file that should be used for executing steps in provider-kubeconfig.
                The default value is ~/.kube/config''')
        parser.add_argument("-s", "--apiserverurl", help='''This flag is to be used to pass the API Server URL of the
                API server on which KubePlus is installed. This API Server URL will be used in constructing the
                server endpoint in the provider kubeconfig. Use the command
                `kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'`
                to retrieve the API Server URL.''')
        parser.add_argument("-f", "--filename", help='''This flag is used to specify the
                output file name in which generated provider kubeconfig will be store
                (The default value is kubeplus-saas-provider.json)''')
        parser.add_argument("-x", "--clustername", help='''This flag is used to specify the name of the cluster.
                This name will be used in setting the value of the context attribute, along with the cluster name,
                in the generated kubeconfig file.''')
        permission_help = "permissions file - use with update command.\n"
        permission_help = permission_help + "Should be a JSON file with the following structure:\n"
        permission_help = permission_help + "{perms:{<apiGroup1>:[{resource1|resource/resourceName::<resourceName>: [verb1, verb2, ...]}, {resource2: [..]}], {<apiGroup2>:[...]}}}"
        parser.add_argument("-p", "--permissionfile", help=permission_help)
        parser.add_argument("-c", "--consumer", help="Generate kubeconfig for consumer")
        pargs = parser.parse_args()
        #print(pargs.action)
        #print(pargs.namespace)

        action = pargs.action
        namespace = pargs.namespace

        if pargs.kubeconfig:
            #print("Kubeconfig file:" + pargs.kubeconfig)
            kubeconfigPath = pargs.kubeconfig

        kubeconfigString = " --kubeconfig=" + kubeconfigPath

        api_s_ip = ''
        if pargs.apiserverurl:
            #print("Server ip:" + pargs.serverip)
            api_s_ip = pargs.apiserverurl

        permission_file = ''
        if pargs.permissionfile:
            #print("Permission file:" + pargs.permissionfile)
            permission_file = pargs.permissionfile

        cluster_name = ''
        if pargs.clustername:
            #print("Cluster name:" + pargs.clustername)
            cluster_name = pargs.clustername

        if action == 'update' and permission_file == '':
            print("Permission file missing. Please provide permission file.")
            print(permission_help)
            exit(0)

        kubeconfigGenerator = KubeconfigGenerator()

        sa = 'kubeplus-saas-provider'
        if pargs.consumer:
            sa = pargs.consumer

        filename = sa
        if pargs.filename:
            filename = pargs.filename
        if not filename.endswith(".json"):
            filename += ".json"

        if action == "create":
                if permission_file:
                    print("Permissions file should be used with update command.")
                    exit(1)

                create_ns = "kubectl get ns " + namespace + kubeconfigString
                out, err = run_command(create_ns)
                if 'not found' in out or 'not found' in err:
                        run_command(create_ns)

                cmd = "kubectl label --overwrite=true ns " + namespace + " managedby=kubeplus " + kubeconfigString
                run_command(cmd)

                # 1. Generate Provider kubeconfig
                if sa == "kubeplus-saas-provider":
                    kubeconfigGenerator._generate_kubeconfig(sa, namespace, filename, api_server_ip=api_s_ip, kubeconfig=kubeconfigString, cluster_name=cluster_name)
                    kubeconfigGenerator._apply_rbac(sa, namespace, entity='provider', kubeconfig=kubeconfigString)
                    print("Provider kubeconfig created: " + filename)
                else:
                    kubeconfigGenerator._generate_kubeconfig(sa, namespace, filename, api_server_ip=api_s_ip, kubeconfig=kubeconfigString, cluster_name=cluster_name)
                    kubeconfigGenerator._apply_rbac(sa, namespace, entity='consumer', kubeconfig=kubeconfigString)
                    print("Consumer kubeconfig created: " + filename)

        if action == "extract":
                kubeconfigGenerator._extract_kubeconfig(sa, namespace, filename, serverip=api_s_ip, kubecfg=kubeconfigString)
                print("Provider kubeconfig created: " + filename)

        if action == "update":
                kubeconfigGenerator._update_rbac(permission_file, sa, namespace, kubeconfigString)
                print("kubeconfig permissions updated: " + filename)


        if action == "delete":
                run_command("kubectl delete sa " + sa + " -n " + namespace + kubeconfigString)
                run_command("kubectl delete configmap " + sa + " -n " + namespace + kubeconfigString)
                run_command("kubectl delete clusterrole " + sa + " -n " + namespace + kubeconfigString)
                run_command("kubectl delete clusterrolebinding " + sa + " -n " + namespace + kubeconfigString)
                run_command("kubectl delete clusterrole " + sa + "-update" + " -n " + namespace + kubeconfigString)
                run_command("kubectl delete clusterrolebinding " + sa + "-update" +  " -n " + namespace + kubeconfigString)
                perms_cfg_map = sa + "-perms"
                run_command("kubectl delete configmap " + perms_cfg_map + " -n " + namespace)
                cwd = os.getcwd()
                run_command("rm " + cwd + "/" + sa + "-secret.yaml")
                run_command("rm " + cwd + "/" + filename)
                run_command("rm " + cwd + "/" + sa + "-role.yaml")
                run_command("rm " + cwd + "/" + sa + "-update-role.yaml")
                run_command("rm " + cwd + "/" + sa + "-rolebinding.yaml")
                run_command("rm " + cwd + "/" + sa + "-role-impersonate.yaml")
                run_command("rm " + cwd + "/" + sa + "-rolebinding-impersonate.yaml")
                run_command("rm " + cwd + "/" + sa + "-update-rolebinding.yaml")
                run_command("rm " + cwd + "/" + sa + "-perms.txt")
                run_command("rm " + cwd + "/" + sa + "-perms-update.txt")

