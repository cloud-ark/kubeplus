import argparse
import json
import os
import subprocess
import sys
import time
import yaml

from logging.config import dictConfig

dictConfig({
    "version": 1,
    "formatters": {"default": {"format": "[%(asctime)s] %(levelname)s in %(module)s: %(message)s"}},
    "handlers": {
        "file.handler": {
            "class": "logging.handlers.RotatingFileHandler",
            "filename": "provider-kubeconfig.log",
            "maxBytes": 10000000,
            "backupCount": 5,
            "level": "DEBUG",
        },
    },
    "root": {"level": "INFO", "handlers": ["file.handler"]},
})


def create_role_rolebinding(contents, name, kubeconfig):
    """Write Role/ClusterRole YAML to file and apply via kubectl."""
    file_path = os.path.join(os.getcwd(), name)
    with open(file_path, "w", encoding="utf-8") as fp:
        fp.write(yaml.dump(contents))
    run_command(" kubectl apply -f " + file_path + kubeconfig)


def run_command(cmd):
    """Execute a shell command. Returns (stdout_str, stderr_str)."""
    with subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True) as proc:
        cmd_out = proc.communicate()
    out = cmd_out[0].decode("utf-8") if cmd_out[0] else ""
    err = cmd_out[1].decode("utf-8") if cmd_out[1] else ""
    return out, err


class KubeconfigGenerator(object):


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

                file_path = os.path.join(os.getcwd(), filename)
                with open(file_path, "w", encoding="utf-8") as fp:
                    fp.write(json.dumps(top_level_dict))

                configmap_name = sa
                created = False
                while not created:
                        run_command("kubectl create configmap " + configmap_name + " -n " + namespace + " --from-file=" + file_path + kubeconfig)
                        get_cmd = "kubectl get configmap " + configmap_name + " -n " + namespace + kubeconfig
                        output, err = run_command(get_cmd)
                        if "Error from server (NotFound)" in (output or ""):
                                time.sleep(2)
                                print("Trying again..")
                        else:
                                created = True


        def _normalize_rule(self, rule):
                return {
                    "apiGroups": tuple(sorted(rule.get("apiGroups", []))),
                    "resources": tuple(sorted(rule.get("resources", []))),
                    "verbs": tuple(sorted(rule.get("verbs", []))),
                    "resourceNames": tuple(sorted(rule.get("resourceNames", []))),
                    "nonResourceURLs": tuple(sorted(rule.get("nonResourceURLs", []))),
                }

        def _normalize_rule_list(self, rules):
                normalized = [self._normalize_rule(r) for r in rules]
                # Sort deterministically by all tuple fields for stable diffs.
                return sorted(
                    normalized,
                    key=lambda r: (
                        r["apiGroups"],
                        r["resources"],
                        r["verbs"],
                        r["resourceNames"],
                        r["nonResourceURLs"],
                    ),
                )

        def _assert_rule_parity(self, label, old_rules, new_rules):
                old_norm = self._normalize_rule_list(old_rules)
                new_norm = self._normalize_rule_list(new_rules)
                if old_norm != new_norm:
                    old_set = set(tuple(sorted(r.items())) for r in old_norm)
                    new_set = set(tuple(sorted(r.items())) for r in new_norm)
                    old_only = sorted(old_set - new_set)
                    new_only = sorted(new_set - old_set)
                    raise AssertionError(
                        f"{label} RBAC mismatch.\n"
                        f"Only in old: {old_only}\n"
                        f"Only in new: {new_only}"
                    )

        def _all_resources_from_rules(self, rules, skip_wildcard=False):
                resources = []
                for rule in rules:
                    for res in rule.get("resources", []):
                        if skip_wildcard and res == "*":
                            continue
                        resources.append(res)
                return sorted(set(resources))

        def _assert_all_resources_parity(self, label, old_resources, new_resources):
                old_set = set(old_resources)
                new_set = set(new_resources)
                if old_set != new_set:
                    old_only = sorted(old_set - new_set)
                    new_only = sorted(new_set - old_set)
                    raise AssertionError(
                        f"{label} all_resources mismatch.\n"
                        f"Only in old: {old_only}\n"
                        f"Only in new: {new_only}"
                    )

        def _build_consumer_rules_old(self):
                # Read all resources
                ruleGroup1 = {}
                apiGroup1 = ["*",""]
                resourceGroup1 = ["*"]
                verbsGroup1 = ["get","watch","list"]
                ruleGroup1["apiGroups"] = apiGroup1
                ruleGroup1["resources"] = resourceGroup1
                ruleGroup1["verbs"] = verbsGroup1

                ruleGroup8 = {}
                apiGroup8 = ["apps"]
                resourceGroup8 = ["deployments","daemonsets","deployments/rollback","deployments/scale","replicasets","replicasets/scale","statefulsets","statefulsets/scale"]
                verbsGroup8 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup8["apiGroups"] = apiGroup8
                ruleGroup8["resources"] = resourceGroup8
                ruleGroup8["verbs"] = verbsGroup8

                # Impersonate users, groups, serviceaccounts
                ruleGroup9 = {}
                apiGroup9 = [""]
                resourceGroup9 = ["users","groups","serviceaccounts"]
                verbsGroup9 = ["impersonate"]
                ruleGroup9["apiGroups"] = apiGroup9
                ruleGroup9["resources"] = resourceGroup9
                ruleGroup9["verbs"] = verbsGroup9

                # Pod/portforward to open consumerui
                ruleGroup10 = {}
                apiGroup10 = [""]
                resourceGroup10 = ["pods/portforward"]
                verbsGroup10 = ["create","get"]
                ruleGroup10["apiGroups"] = apiGroup10
                ruleGroup10["resources"] = resourceGroup10
                ruleGroup10["verbs"] = verbsGroup10

                ruleList = []
                ruleList.append(ruleGroup1)
                ruleList.append(ruleGroup9)
                ruleList.append(ruleGroup10)
                ruleList.append(ruleGroup8)
                return ruleList

        def _build_consumer_rules_new(self):
                return [
                    {"apiGroups": ["*", ""], "resources": ["*"], "verbs": ["get", "watch", "list"]},
                    {
                        "apiGroups": ["apps"],
                        "resources": [
                            "deployments", "daemonsets", "deployments/rollback", "deployments/scale",
                            "replicasets", "replicasets/scale", "statefulsets", "statefulsets/scale",
                        ],
                        "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"],
                    },
                    {"apiGroups": [""], "resources": ["users", "groups", "serviceaccounts"], "verbs": ["impersonate"]},
                    {"apiGroups": [""], "resources": ["pods/portforward"], "verbs": ["create", "get"]},
                ]

        def _build_provider_rules_old(self):
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

                # CRUD on clusterroles and clusterrolebindings
                ruleGroup3 = {}
                apiGroup3 = ["rbac.authorization.k8s.io"]
                resourceGroup3 = ["clusterroles","clusterrolebindings","roles","rolebindings"]
                verbsGroup3 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup3["apiGroups"] = apiGroup3
                ruleGroup3["resources"] = resourceGroup3
                ruleGroup3["verbs"] = verbsGroup3

                # CRUD on Port forward
                ruleGroup4 = {}
                apiGroup4 = [""]
                resourceGroup4 = ["pods/portforward"]
                verbsGroup4 = ["get","watch","list","create","delete","update","patch"]
                ruleGroup4["apiGroups"] = apiGroup4
                ruleGroup4["resources"] = resourceGroup4
                ruleGroup4["verbs"] = verbsGroup4

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
                resourceGroup6 = ["secrets","serviceaccounts","configmaps","events","persistentvolumeclaims","serviceaccounts/token","services","services/proxy","endpoints"]
                verbsGroup6 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup6["apiGroups"] = apiGroup6
                ruleGroup6["resources"] = resourceGroup6
                ruleGroup6["verbs"] = verbsGroup6

                # CRUD on namespaces
                ruleGroup7 = {}
                apiGroup7 = [""]
                resourceGroup7 = ["namespaces"]
                verbsGroup7 = ["get","watch","list","create","delete","update","patch"]
                ruleGroup7["apiGroups"] = apiGroup7
                ruleGroup7["resources"] = resourceGroup7
                ruleGroup7["verbs"] = verbsGroup7

                # CRUD on Deployments
                ruleGroup8 = {}
                apiGroup8 = ["apps"]
                resourceGroup8 = ["deployments","daemonsets","deployments/rollback","deployments/scale","replicasets","replicasets/scale","statefulsets","statefulsets/scale"]
                verbsGroup8 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup8["apiGroups"] = apiGroup8
                ruleGroup8["resources"] = resourceGroup8
                ruleGroup8["verbs"] = verbsGroup8

                # Impersonate users, groups, serviceaccounts
                ruleGroup9 = {}
                apiGroup9 = [""]
                resourceGroup9 = ["users","groups","serviceaccounts"]
                verbsGroup9 = ["impersonate"]
                ruleGroup9["apiGroups"] = apiGroup9
                ruleGroup9["resources"] = resourceGroup9
                ruleGroup9["verbs"] = verbsGroup9

                # Exec into the Pods and others in the "" apiGroup
                ruleGroup10 = {}
                apiGroup10 = [""]
                resourceGroup10 = ["pods","pods/attach","pods/exec","pods/portforward","pods/proxy","pods/eviction","replicationcontrollers","replicationcontrollers/scale"]
                verbsGroup10 = ["get","list","create","update","delete","watch","patch","deletecollection"]
                ruleGroup10["apiGroups"] = apiGroup10
                ruleGroup10["resources"] = resourceGroup10
                ruleGroup10["verbs"] = verbsGroup10

                # AdmissionRegistration
                ruleGroup11 = {}
                apiGroup11 = ["admissionregistration.k8s.io"]
                resourceGroup11 = ["mutatingwebhookconfigurations"]
                verbsGroup11 = ["get","create","delete","update"]
                ruleGroup11["apiGroups"] = apiGroup11
                ruleGroup11["resources"] = resourceGroup11
                ruleGroup11["verbs"] = verbsGroup11

                # APIExtension
                ruleGroup12 = {}
                apiGroup12 = ["apiextensions.k8s.io"]
                resourceGroup12 = ["customresourcedefinitions"]
                verbsGroup12 = ["get","create","delete","update","patch"]
                ruleGroup12["apiGroups"] = apiGroup12
                ruleGroup12["resources"] = resourceGroup12
                ruleGroup12["verbs"] = verbsGroup12

                # Certificates
                ruleGroup13 = {}
                apiGroup13 = ["certificates.k8s.io"]
                resourceGroup13 = ["signers"]
                resourceNames13 = ["kubernetes.io/legacy-unknown","kubernetes.io/kubelet-serving","kubernetes.io/kube-apiserver-client","cloudark.io/kubeplus"]
                verbsGroup13 = ["get","create","delete","update","patch","approve"]
                ruleGroup13["apiGroups"] = apiGroup13
                ruleGroup13["resources"] = resourceGroup13
                ruleGroup13["resourceNames"] = resourceNames13
                ruleGroup13["verbs"] = verbsGroup13

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
                resourceGroup15 = ["certificatesigningrequests","certificatesigningrequests/approval"]
                verbsGroup15 = ["create","delete","update","patch"]
                ruleGroup15["apiGroups"] = apiGroup15
                ruleGroup15["resources"] = resourceGroup15
                ruleGroup15["verbs"] = verbsGroup15

                ruleGroup16 = {}
                apiGroup16 = ["extensions"]
                resourceGroup16 = ["deployments","daemonsets","deployments/rollback","deployments/scale","replicasets","replicasets/scale","replicationcontrollers/scale","ingresses","networkpolicies"]
                verbsGroup16 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup16["apiGroups"] = apiGroup16
                ruleGroup16["resources"] = resourceGroup16
                ruleGroup16["verbs"] = verbsGroup16

                ruleGroup17 = {}
                apiGroup17 = ["networking.k8s.io"]
                resourceGroup17 = ["ingresses","networkpolicies"]
                verbsGroup17 = ["get","watch","list","create","delete","update","patch","deletecollection"]
                ruleGroup17["apiGroups"] = apiGroup17
                ruleGroup17["resources"] = resourceGroup17
                ruleGroup17["verbs"] = verbsGroup17

                ruleGroup18 = {}
                apiGroup18 = ["authorization.k8s.io"]
                resourceGroup18 = ["localsubjectaccessreviews"]
                verbsGroup18 = ["create"]
                ruleGroup18["apiGroups"] = apiGroup18
                ruleGroup18["resources"] = resourceGroup18
                ruleGroup18["verbs"] = verbsGroup18

                ruleGroup19 = {}
                apiGroup19 = ["autoscaling"]
                resourceGroup19 = ["horizontalpodautoscalers"]
                verbsGroup19 = ["create","delete","deletecollection","patch","update"]
                ruleGroup19["apiGroups"] = apiGroup19
                ruleGroup19["resources"] = resourceGroup19
                ruleGroup19["verbs"] = verbsGroup19

                ruleGroup20 = {}
                apiGroup20 = ["batch"]
                resourceGroup20 = ["cronjobs","jobs"]
                verbsGroup20 = ["create","delete","deletecollection","patch","update"]
                ruleGroup20["apiGroups"] = apiGroup20
                ruleGroup20["resources"] = resourceGroup20
                ruleGroup20["verbs"] = verbsGroup20

                ruleGroup21 = {}
                apiGroup21 = ["policy"]
                resourceGroup21 = ["poddisruptionbudgets"]
                verbsGroup21 = ["create","delete","deletecollection","patch","update"]
                ruleGroup21["apiGroups"] = apiGroup21
                ruleGroup21["resources"] = resourceGroup21
                ruleGroup21["verbs"] = verbsGroup21

                ruleGroup22 = {}
                apiGroup22 = [""]
                resourceGroup22 = ["resourcequotas"]
                verbsGroup22 = ["create","delete","deletecollection","patch","update"]
                ruleGroup22["apiGroups"] = apiGroup22
                ruleGroup22["resources"] = resourceGroup22
                ruleGroup22["verbs"] = verbsGroup22

                # PersistentVolumes and PersistentVolumeClaims for charts storage in helmer container
                ruleGroup23 = {}
                apiGroup23 = [""]
                resourceGroup23 = ["persistentvolumes","persistentvolumeclaims"]
                verbsGroup23 = ["get","watch","list","create","delete","update","patch"]
                ruleGroup23["apiGroups"] = apiGroup23
                ruleGroup23["resources"] = resourceGroup23
                ruleGroup23["verbs"] = verbsGroup23

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
                return ruleList

        def _build_provider_rules_new(self):
                return [
                    {"apiGroups": ["*", ""], "resources": ["*"], "verbs": ["get", "watch", "list"]},
                    {
                        "apiGroups": ["workflows.kubeplus"],
                        "resources": ["resourcecompositions", "resourcemonitors", "resourcepolicies", "resourceevents"],
                        "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"],
                    },
                    {
                        "apiGroups": ["rbac.authorization.k8s.io"],
                        "resources": ["clusterroles", "clusterrolebindings", "roles", "rolebindings"],
                        "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"],
                    },
                    {"apiGroups": [""], "resources": ["pods/portforward"], "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"]},
                    {"apiGroups": ["platformapi.kubeplus"], "resources": ["*"], "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"]},
                    {
                        "apiGroups": [""],
                        "resources": ["secrets", "serviceaccounts", "configmaps", "events", "persistentvolumeclaims", "serviceaccounts/token", "services", "services/proxy", "endpoints"],
                        "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"],
                    },
                    {"apiGroups": [""], "resources": ["namespaces"], "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"]},
                    {
                        "apiGroups": ["apps"],
                        "resources": ["deployments", "daemonsets", "deployments/rollback", "deployments/scale", "replicasets", "replicasets/scale", "statefulsets", "statefulsets/scale"],
                        "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"],
                    },
                    {"apiGroups": [""], "resources": ["users", "groups", "serviceaccounts"], "verbs": ["impersonate"]},
                    {
                        "apiGroups": [""],
                        "resources": ["pods", "pods/attach", "pods/exec", "pods/portforward", "pods/proxy", "pods/eviction", "replicationcontrollers", "replicationcontrollers/scale"],
                        "verbs": ["get", "list", "create", "update", "delete", "watch", "patch", "deletecollection"],
                    },
                    {"apiGroups": ["admissionregistration.k8s.io"], "resources": ["mutatingwebhookconfigurations"], "verbs": ["get", "create", "delete", "update"]},
                    {"apiGroups": ["apiextensions.k8s.io"], "resources": ["customresourcedefinitions"], "verbs": ["get", "create", "delete", "update", "patch"]},
                    {
                        "apiGroups": ["certificates.k8s.io"],
                        "resources": ["signers"],
                        "resourceNames": ["kubernetes.io/legacy-unknown", "kubernetes.io/kubelet-serving", "kubernetes.io/kube-apiserver-client", "cloudark.io/kubeplus"],
                        "verbs": ["get", "create", "delete", "update", "patch", "approve"],
                    },
                    {"apiGroups": ["*"], "resources": ["*"], "verbs": ["get"]},
                    {"apiGroups": ["certificates.k8s.io"], "resources": ["certificatesigningrequests", "certificatesigningrequests/approval"], "verbs": ["create", "delete", "update", "patch"]},
                    {
                        "apiGroups": ["extensions"],
                        "resources": ["deployments", "daemonsets", "deployments/rollback", "deployments/scale", "replicasets", "replicasets/scale", "replicationcontrollers/scale", "ingresses", "networkpolicies"],
                        "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"],
                    },
                    {"apiGroups": ["networking.k8s.io"], "resources": ["ingresses", "networkpolicies"], "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"]},
                    {"apiGroups": ["authorization.k8s.io"], "resources": ["localsubjectaccessreviews"], "verbs": ["create"]},
                    {"apiGroups": ["autoscaling"], "resources": ["horizontalpodautoscalers"], "verbs": ["create", "delete", "deletecollection", "patch", "update"]},
                    {"apiGroups": ["batch"], "resources": ["cronjobs", "jobs"], "verbs": ["create", "delete", "deletecollection", "patch", "update"]},
                    {"apiGroups": ["policy"], "resources": ["poddisruptionbudgets"], "verbs": ["create", "delete", "deletecollection", "patch", "update"]},
                    {"apiGroups": [""], "resources": ["resourcequotas"], "verbs": ["create", "delete", "deletecollection", "patch", "update"]},
                    {"apiGroups": [""], "resources": ["persistentvolumes", "persistentvolumeclaims"], "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"]},
                ]

        def _apply_consumer_rbac(self, sa, namespace, kubeconfig):
                """Apply ClusterRole and ClusterRoleBinding for consumer (read + apps + impersonate + portforward)."""
                old_rule_list = self._build_consumer_rules_old()
                new_rule_list = self._build_consumer_rules_new()
                old_all_resources = self._all_resources_from_rules(old_rule_list)
                new_all_resources = self._all_resources_from_rules(new_rule_list)
                if os.getenv("KUBEPLUS_RBAC_EQ_CHECK", "0") == "1":
                    self._assert_rule_parity("consumer", old_rule_list, new_rule_list)
                    self._assert_all_resources_parity("consumer", old_all_resources, new_all_resources)
                # Keep old path as source of truth in this PR.
                rule_list = old_rule_list
                all_resources = old_all_resources
                role = {
                    "apiVersion": "rbac.authorization.k8s.io/v1",
                    "kind": "ClusterRole",
                    "metadata": {"name": sa, "namespace": namespace},
                    "rules": rule_list,
                }
                create_role_rolebinding(role, sa + "-role-impersonate.yaml", kubeconfig)

                role_binding = {
                    "apiVersion": "rbac.authorization.k8s.io/v1",
                    "kind": "ClusterRoleBinding",
                    "metadata": {"name": sa, "namespace": namespace},
                    "subjects": [{"kind": "ServiceAccount", "name": sa, "namespace": namespace, "apiGroup": ""}],
                    "roleRef": {"kind": "ClusterRole", "name": sa, "apiGroup": "rbac.authorization.k8s.io"},
                }
                create_role_rolebinding(role_binding, sa + "-rolebinding-impersonate.yaml", kubeconfig)
                cfg_map_filename = sa + "-perms.txt"
                with open(cfg_map_filename, "w", encoding="utf-8") as fp:
                    fp.write(str(all_resources))
                run_command(
                    "kubectl create configmap " + sa + "-perms -n " + namespace
                    + " --from-file=" + cfg_map_filename + kubeconfig
                )


        def _apply_provider_rbac(self, sa, namespace, kubeconfig):
                """Apply ClusterRole and ClusterRoleBinding for provider (full platform operator permissions)."""
                old_rule_list = self._build_provider_rules_old()
                new_rule_list = self._build_provider_rules_new()
                old_all_resources = self._all_resources_from_rules(old_rule_list, skip_wildcard=True)
                new_all_resources = self._all_resources_from_rules(new_rule_list, skip_wildcard=True)
                if os.getenv("KUBEPLUS_RBAC_EQ_CHECK", "0") == "1":
                    self._assert_rule_parity("provider", old_rule_list, new_rule_list)
                    self._assert_all_resources_parity("provider", old_all_resources, new_all_resources)
                # Keep old path as source of truth in this PR.
                rule_list = old_rule_list
                all_resources = old_all_resources

                role = {
                    "apiVersion": "rbac.authorization.k8s.io/v1",
                    "kind": "ClusterRole",
                    "metadata": {"name": sa, "namespace": namespace},
                    "rules": rule_list,
                }
                create_role_rolebinding(role, sa + "-role.yaml", kubeconfig)

                role_binding = {
                    "apiVersion": "rbac.authorization.k8s.io/v1",
                    "kind": "ClusterRoleBinding",
                    "metadata": {"name": sa, "namespace": namespace},
                    "subjects": [{"kind": "ServiceAccount", "name": sa, "namespace": namespace, "apiGroup": ""}],
                    "roleRef": {"kind": "ClusterRole", "name": sa, "apiGroup": "rbac.authorization.k8s.io"},
                }
                create_role_rolebinding(role_binding, sa + "-rolebinding.yaml", kubeconfig)

                cfg_map_filename = sa + "-perms.txt"
                with open(cfg_map_filename, "w", encoding="utf-8") as fp:
                    fp.write(str(all_resources))
                run_command(
                    "kubectl create configmap " + sa + "-perms -n " + namespace
                    + " --from-file=" + cfg_map_filename + kubeconfig
                )

        def _update_rbac(self, permissionfile, sa, namespace, kubeconfig):
                """Add permissions from JSON file to provider (update command)."""
                with open(permissionfile, "r", encoding="utf-8") as fp:
                    perms_data = json.load(fp)
                perms = perms_data["perms"]
                rule_list = []
                new_resources = []

                for api_group, res_actions in perms.items():
                    for res in res_actions:
                        for resource, verbs in res.items():
                            if resource not in new_resources:
                                new_resources.append(resource.strip())
                            rule_group = {}
                            if api_group == "non-apigroup":
                                if "nonResourceURL" in resource:
                                    parts = resource.split("nonResourceURL::")
                                    non_res = parts[1].strip() if len(parts) > 1 else parts[0].strip()
                                    rule_group["nonResourceURLs"] = [non_res]
                                    rule_group["verbs"] = verbs
                            else:
                                rule_group["apiGroups"] = [api_group]
                                rule_group["verbs"] = verbs
                                if "resourceName" in resource:
                                    parts = resource.split("/resourceName::")
                                    rule_group["resources"] = [parts[0].strip()]
                                    rule_group["resourceNames"] = [parts[1].strip()]
                                else:
                                    rule_group["resources"] = [resource]
                            rule_list.append(rule_group)

                role = {
                    "apiVersion": "rbac.authorization.k8s.io/v1",
                    "kind": "ClusterRole",
                    "metadata": {"name": sa + "-update", "namespace": namespace},
                    "rules": rule_list,
                }
                create_role_rolebinding(role, sa + "-update-role.yaml", kubeconfig)

                role_binding = {
                    "apiVersion": "rbac.authorization.k8s.io/v1",
                    "kind": "ClusterRoleBinding",
                    "metadata": {"name": sa + "-update", "namespace": namespace},
                    "subjects": [{"kind": "ServiceAccount", "name": sa, "namespace": namespace, "apiGroup": ""}],
                    "roleRef": {"kind": "ClusterRole", "name": sa + "-update", "apiGroup": "rbac.authorization.k8s.io"},
                }
                create_role_rolebinding(role_binding, sa + "-update-rolebinding.yaml", kubeconfig)

                cfg_map_name = sa + "-perms"
                cfg_map_filename = sa + "-perms.txt"
                out, _ = run_command("kubectl get configmap " + cfg_map_name + " -o json -n " + namespace + kubeconfig)
                kubeplus_perms = []
                if out:
                    json_op = json.loads(out)
                    perms_str = json_op.get("data", {}).get(cfg_map_filename, "")
                    for p in perms_str.replace("'", "").replace("[", "").replace("]", "").split(","):
                        p = p.strip()
                        if p:
                            kubeplus_perms.append(p)
                new_resources.extend(kubeplus_perms)

                run_command("kubectl delete configmap " + cfg_map_name + " -n " + namespace + kubeconfig)
                new_resources = sorted(set(new_resources))
                with open(cfg_map_filename, "w", encoding="utf-8") as fp:
                    fp.write(str(new_resources))
                run_command(
                    "kubectl create configmap " + cfg_map_name + " -n " + namespace
                    + " --from-file=" + cfg_map_filename + kubeconfig
                )
    

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

                secret_name = sa + "-secret.yaml"
                file_path = os.path.join(os.getcwd(), secret_name)
                with open(file_path, "w", encoding="utf-8") as fp:
                    fp.write(yaml.dump(secret))
                created = False
                count = 0
                while not created and count < 5:
                        cmd = " kubectl create -f " + file_path + kubeconfig
                        out, err = run_command(cmd)
                        if out and "created" in out:
                                created = True
                        else:
                                time.sleep(2)
                                count += 1
                #print("Create secret:" + out)
                if not created and count >= 5:
                    print(err)
                    sys.exit(1)
                return out

        def _extract_kubeconfig(self, sa, namespace, filename, serverip="", kubecfg="", cluster_name=None):
                """Extract token from secret, determine server URL, build kubeconfig and store in ConfigMap."""
                token_found = False
                token = ""
                while not token_found:
                    out, _ = run_command(
                        " kubectl describe secret " + sa + " -n " + namespace + kubecfg
                    )
                    token = ""
                    for line in (out or "").split("\n"):
                        parts = line.split(":", 1)
                        if len(parts) == 2 and parts[0].strip() == "token":
                            token = parts[1].strip()
                    if token != "":
                        token_found = True
                    else:
                        time.sleep(2)

                out, err = run_command(" kubectl get secret " + sa + " -n " + namespace + " -o json " + kubecfg)
                if not out:
                    raise RuntimeError(
                        f"Failed to fetch secret {sa!r} in ns {namespace!r}: {err}"
                    )
                json_out = json.loads(out)
                ca_cert = json_out["data"]["ca.crt"].strip()

                if serverip:
                    server = serverip if "https" in serverip else "https://" + serverip
                else:
                    out2, _ = run_command(
                        "kubectl -n default get endpoints kubernetes " + kubecfg
                        + " | awk '{print $2}' | grep -v ENDPOINTS"
                    )
                    server = out2.strip() if out2 else ""
                    server = "https://" + server if server else ""
                if not server or server.rstrip("/") == "https:":
                    raise RuntimeError(
                        "Could not determine API server endpoint; pass -s/--apiserverurl"
                    )

                self._create_kubecfg_file(sa, namespace, filename, token, ca_cert, server, kubecfg, cluster_name)


        def _generate_kubeconfig(self, sa, namespace, filename, api_server_ip="", kubeconfig="", cluster_name=None):
                run_command(" kubectl create sa " + sa + " -n " + namespace + kubeconfig)

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


if __name__ == "__main__":
        parser = argparse.ArgumentParser(
            description="Generate kubeconfig files for KubePlus provider and consumer. "
            "Creates ServiceAccounts, RBAC (ClusterRoles/RoleBindings), and kubeconfig files "
            "used by KubePlus for platform operations (provider) or tenant access (consumer). "
            "The generated kubeconfig includes the namespace in the context so kubectl defaults to it; "
            "consumer RBAC restricts what operations are allowed.",
        )
        parser.add_argument("action", help="command", choices=['create', 'delete', 'update', 'extract'])
        parser.add_argument(
            "namespace",
            help="Namespace where the ServiceAccount is created. "
            "Provider: typically the KubePlus install namespace (e.g. default). "
            "Consumer (-c): namespace where the consumer SA lives and access is restricted.",
        )
        parser.add_argument(
            "-k", "--kubeconfig",
            help="Path to kubeconfig for executing steps. Default: ~/.kube/config",
        )
        parser.add_argument(
            "-s", "--apiserverurl",
            help="API Server URL for the generated kubeconfig. Use "
            "kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' to retrieve it.",
        )
        parser.add_argument(
            "-f", "--filename",
            help="Output filename. Default: kubeplus-saas-provider.json (or <consumer>.json with -c)",
        )
        parser.add_argument(
            "-x", "--clustername",
            help="Cluster name for context and cluster in the generated kubeconfig file.",
        )
        permission_help = "Permissions file - use with update command. "
        permission_help += "JSON structure: {perms:{<apiGroup>:[{resource|resource/resourceName::<name>:[verbs]},...]}}"
        parser.add_argument("-p", "--permissionfile", help=permission_help)
        parser.add_argument(
            "-c", "--consumer",
            help="Generate consumer kubeconfig. Use consumer name as value (e.g. -c team1).",
        )
        pargs = parser.parse_args()
        action = pargs.action
        namespace = pargs.namespace
        kubeconfig_path = pargs.kubeconfig or os.path.join(os.path.expanduser("~"), ".kube", "config")
        kubeconfigString = " --kubeconfig=" + kubeconfig_path
        api_s_ip = pargs.apiserverurl or ""
        permission_file = pargs.permissionfile or ""
        cluster_name = pargs.clustername or ""

        if action == 'update' and permission_file == '':
            print("Permission file missing. Please provide -p/--permissionfile.")
            sys.exit(1)

        kubeconfigGenerator = KubeconfigGenerator()

        sa = 'kubeplus-saas-provider'
        if pargs.consumer:
            sa = pargs.consumer

        filename = pargs.filename or sa
        if not filename.endswith(".json"):
            filename += ".json"

        if action == "create":
                if permission_file:
                    print("Permissions file should be used with update command.")
                    sys.exit(1)

                get_ns = "kubectl get ns " + namespace + kubeconfigString
                out, err = run_command(get_ns)
                if 'not found' in (out or '') or 'not found' in (err or ''):
                    run_command("kubectl create ns " + namespace + kubeconfigString)

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
                run_command("kubectl delete clusterrole " + sa + kubeconfigString)
                run_command("kubectl delete clusterrolebinding " + sa + kubeconfigString)
                run_command("kubectl delete clusterrole " + sa + "-update" + kubeconfigString)
                run_command("kubectl delete clusterrolebinding " + sa + "-update" + kubeconfigString)
                run_command("kubectl delete configmap " + sa + "-perms -n " + namespace + kubeconfigString)
                cwd = os.getcwd()
                for f in [
                    sa + "-secret.yaml", filename, sa + "-role.yaml", sa + "-update-role.yaml",
                    sa + "-rolebinding.yaml", sa + "-role-impersonate.yaml",
                    sa + "-rolebinding-impersonate.yaml", sa + "-update-rolebinding.yaml",
                    sa + "-perms.txt", sa + "-perms-update.txt",
                ]:
                    path = os.path.join(cwd, f)
                    if os.path.exists(path):
                        try:
                            os.remove(path)
                        except OSError:
                            pass

