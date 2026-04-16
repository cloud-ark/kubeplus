import argparse
import json
import os
import subprocess
import sys
import time
import yaml

from logging.config import dictConfig

RESOURCE_NAME_SEP = "/resourceName::"
NON_RESOURCE_URL_PREFIX = "nonResourceURL::"

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
    proc = subprocess.run(cmd, shell=True, capture_output=True, encoding="utf-8")
    return proc.stdout or "", proc.stderr or ""


def run_command_with_code(cmd):
    """Execute a shell command. Returns (stdout_str, stderr_str, returncode)."""
    proc = subprocess.run(cmd, shell=True, capture_output=True, encoding="utf-8")
    return proc.stdout or "", proc.stderr or "", proc.returncode


class KubeconfigGenerator(object):
        def _load_permission_data(self, permissionfile):
                """Load permissions definition from JSON or YAML file."""
                with open(permissionfile, "r", encoding="utf-8") as fp:
                    contents = fp.read()
                try:
                    perms_data = json.loads(contents)
                except json.JSONDecodeError as json_exc:
                    try:
                        perms_data = yaml.safe_load(contents)
                    except yaml.YAMLError as yaml_exc:
                        raise ValueError(
                            f"Permission file is neither valid JSON ({json_exc}) nor YAML ({yaml_exc})."
                        ) from yaml_exc
                if not isinstance(perms_data, dict) or "perms" not in perms_data:
                    raise ValueError("Permission file must define a top-level 'perms' object.")
                if not isinstance(perms_data["perms"], dict):
                    raise ValueError("'perms' must be a mapping of apiGroup to permission rules.")
                return perms_data["perms"]

        def _parse_permission_rules(self, perms):
                """Convert permission mapping to k8s rule list and tracked resources list."""
                rule_list = []
                resources = []
                for api_group, res_actions in perms.items():
                    for res in res_actions:
                        for resource, verbs in res.items():
                            verbs = verbs or []
                            resource_key = resource.strip()
                            if resource_key not in resources:
                                resources.append(resource_key)
                            rule_group = {}
                            if api_group == "non-apigroup":
                                if NON_RESOURCE_URL_PREFIX in resource:
                                    parts = resource.split(NON_RESOURCE_URL_PREFIX)
                                    non_res = parts[1].strip() if len(parts) > 1 else parts[0].strip()
                                    rule_group["nonResourceURLs"] = [non_res]
                                    rule_group["verbs"] = verbs
                            else:
                                rule_group["apiGroups"] = [api_group]
                                rule_group["verbs"] = verbs
                                if RESOURCE_NAME_SEP in resource:
                                    parts = resource.split(RESOURCE_NAME_SEP)
                                    rule_group["resources"] = [parts[0].strip()]
                                    rule_group["resourceNames"] = [parts[1].strip()]
                                else:
                                    rule_group["resources"] = [resource_key]
                            rule_list.append(rule_group)
                return rule_list, resources

        def _kubectl_or_raise(self, args, kubeconfig, *, ignore_not_found=True):
                """Run a kubectl subcommand and raise on failure."""
                _, err, rc = run_command_with_code("kubectl " + args + kubeconfig)
                if rc == 0:
                        return
                if ignore_not_found and "(NotFound)" in err:
                        return
                raise RuntimeError(f"kubectl {args} failed: {err.strip()}")

        def _fetch_clusterrole(self, name, kubeconfig):
                """Return ClusterRole JSON dict, or None when role does not exist."""
                out, err, rc = run_command_with_code("kubectl get clusterrole " + name + " -o json" + kubeconfig)
                if rc == 0 and out:
                        return json.loads(out)
                if "(NotFound)" in err:
                        return None
                raise RuntimeError(f"Failed to fetch clusterrole {name!r}: {err.strip()}")

        def _rules_changed(self, existing_rules, remaining_rules):
                return self._normalize_rule_list(existing_rules) != self._normalize_rule_list(remaining_rules)

        def _rebuild_perm_configmap_from_roles(self, sa, namespace, kubeconfig):
                """Rebuild <sa>-perms from current <sa> and <sa>-update roles."""
                resources = set()
                for role_name in [sa, sa + "-update"]:
                        role_obj = self._fetch_clusterrole(role_name, kubeconfig)
                        if role_obj is None:
                                continue
                        for rule in role_obj.get("rules") or []:
                                resources.update(self._configmap_keys_for_rule(rule))
                self._write_perm_configmap_resources(sa, namespace, kubeconfig, sorted(resources))

        def _read_perm_configmap_resources(self, sa, namespace, kubeconfig):
                cfg_map_name = sa + "-perms"
                cfg_map_filename = sa + "-perms.txt"
                out, err, rc = run_command_with_code(
                    "kubectl get configmap " + cfg_map_name + " -o json -n " + namespace + kubeconfig
                )
                if rc != 0 and "(NotFound)" not in err:
                    raise RuntimeError(f"Failed to read configmap {cfg_map_name!r}: {err.strip()}")
                kubeplus_perms = []
                if out:
                    json_op = json.loads(out)
                    perms_str = json_op.get("data", {}).get(cfg_map_filename, "")
                    for p in perms_str.replace("'", "").replace("[", "").replace("]", "").split(","):
                        p = p.strip()
                        if p:
                            kubeplus_perms.append(p)
                return kubeplus_perms

        def _write_perm_configmap_resources(self, sa, namespace, kubeconfig, resources):
                cfg_map_name = sa + "-perms"
                cfg_map_filename = sa + "-perms.txt"
                self._kubectl_or_raise(
                    "delete configmap " + cfg_map_name
                    + " -n " + namespace + " --ignore-not-found",
                    kubeconfig,
                )
                if not resources:
                    return
                with open(cfg_map_filename, "w", encoding="utf-8") as fp:
                    fp.write(str(sorted(set(resources))))
                self._kubectl_or_raise(
                    "create configmap " + cfg_map_name + " -n " + namespace
                    + " --from-file=" + cfg_map_filename,
                    kubeconfig,
                    ignore_not_found=False,
                )



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
                    "apiGroups": tuple(sorted(set(rule.get("apiGroups", [])))),
                    "resources": tuple(sorted(set(rule.get("resources", [])))),
                    "verbs": tuple(sorted(set(rule.get("verbs", [])))),
                    "resourceNames": tuple(sorted(set(rule.get("resourceNames", [])))),
                    "nonResourceURLs": tuple(sorted(set(rule.get("nonResourceURLs", [])))),
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
                # Keep old path as source of truth.
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
                # Keep old path as source of truth.
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
                """Add permissions from JSON/YAML file to provider (update command)."""
                perms = self._load_permission_data(permissionfile)
                rule_list, new_resources = self._parse_permission_rules(perms)

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

                kubeplus_perms = self._read_perm_configmap_resources(sa, namespace, kubeconfig)
                new_resources.extend(kubeplus_perms)
                self._write_perm_configmap_resources(sa, namespace, kubeconfig, new_resources)

        def _remove_revoked_verbs(self, existing_verbs, revoke_verbs):
                """Remove revoked verbs from an existing verb list.

                Returns None when the rule should be dropped entirely.
                """
                existing_list = list(existing_verbs)
                revoke_set = set(revoke_verbs)
                if "*" in existing_list and revoke_set:
                        return None
                if "*" in revoke_set:
                        return None
                remaining = [v for v in existing_list if v not in revoke_set]
                return remaining or None

        def _configmap_keys_for_rule(self, rule):
                """Return configmap keys represented by a rule."""
                keys = []
                for url in rule.get("nonResourceURLs") or []:
                        keys.append(NON_RESOURCE_URL_PREFIX + url)
                for res in rule.get("resources") or []:
                        keys.append(res)
                return keys

        def _revoke_targets_rule(self, existing_norm, revoke_norm_rule):
                """Return True when revoke should apply to this rule.

                Matching rules:
                - For URL rules, compare only nonResourceURLs.
                - For normal resource rules, compare apiGroups + resources.
                - A '*' on either side means "match all" for that field.
                """
                def _overlaps(revoke_values, existing_values):
                        if not revoke_values or not existing_values:
                                return False
                        if "*" in revoke_values or "*" in existing_values:
                                return True
                        return bool(set(revoke_values).intersection(existing_values))

                if revoke_norm_rule["nonResourceURLs"]:
                        return _overlaps(revoke_norm_rule["nonResourceURLs"], existing_norm["nonResourceURLs"])
                if not _overlaps(revoke_norm_rule["apiGroups"], existing_norm["apiGroups"]):
                        return False
                if not _overlaps(revoke_norm_rule["resources"], existing_norm["resources"]):
                        return False
                return True

        def _build_remaining_rules_for_role(self, role_obj, revoke_norm):
                """Return (existing_rules, remaining_rules) for one role after revoke."""
                existing_rules = role_obj.get("rules") or []
                remaining_rules = []
                for rule in existing_rules:
                        existing_norm = self._normalize_rule(rule)
                        matched_revoke_rules = [
                            rule_norm for rule_norm in revoke_norm
                            if self._revoke_targets_rule(existing_norm, rule_norm)
                        ]
                        # Case 1: revoke does not target this rule.
                        if not matched_revoke_rules:
                                remaining_rules.append(rule)
                                continue

                        existing_verbs = rule.get("verbs", [])
                        existing_urls = rule.get("nonResourceURLs") or []
                        if existing_urls:
                                # Case 2: URL rules (nonResourceURLs).
                                if "*" in existing_urls:
                                        # URL wildcard ("*"): keep one rule and trim verbs only.
                                        combined_revoke_verbs = [v for mr in matched_revoke_rules for v in mr["verbs"]]
                                        stripped_verbs = self._remove_revoked_verbs(existing_verbs, combined_revoke_verbs)
                                        if stripped_verbs:
                                                updated_rule = dict(rule)
                                                updated_rule["verbs"] = stripped_verbs
                                                remaining_rules.append(updated_rule)
                                        continue

                                # Specific URL list: process each URL separately so unrelated URLs stay.
                                remaining_by_verbs = {}
                                for url in existing_urls:
                                        revoke_verbs = []
                                        for matched_rule in matched_revoke_rules:
                                                matched_urls = set(matched_rule["nonResourceURLs"])
                                                if not matched_urls:
                                                        continue
                                                url_matches = "*" in matched_urls or url in matched_urls
                                                if not url_matches:
                                                        continue
                                                revoke_verbs.extend(matched_rule["verbs"])
                                        new_verbs = self._remove_revoked_verbs(existing_verbs, revoke_verbs)
                                        if new_verbs:
                                                remaining_by_verbs.setdefault(tuple(new_verbs), []).append(url)
                                for verbs_key, urls in sorted(remaining_by_verbs.items()):
                                        updated_rule = dict(rule)
                                        updated_rule["nonResourceURLs"] = sorted(urls)
                                        updated_rule["verbs"] = list(verbs_key)
                                        remaining_rules.append(updated_rule)
                                continue

                        existing_resources = rule.get("resources") or []
                        existing_api_groups = rule.get("apiGroups") or []
                        is_atomic_scope = (not existing_resources) or ("*" in existing_api_groups) or ("*" in existing_resources)
                        if is_atomic_scope:
                                # Case 3: wildcard/no-resource rule.
                                # We cannot remove just one sub-part, so trim verbs on whole rule.
                                combined_revoke_verbs = [v for mr in matched_revoke_rules for v in mr["verbs"]]
                                stripped_verbs = self._remove_revoked_verbs(existing_verbs, combined_revoke_verbs)
                                if stripped_verbs:
                                        updated_rule = dict(rule)
                                        updated_rule["verbs"] = stripped_verbs
                                        remaining_rules.append(updated_rule)
                                continue

                        # Case 4: explicit resource list.
                        # Revoke per resource so one resource change does not affect others.
                        remaining_by_verbs = {}
                        for resource in existing_resources:
                                revoke_verbs = []
                                for matched_rule in matched_revoke_rules:
                                        matched_resources = set(matched_rule["resources"])
                                        if not matched_resources:
                                                continue
                                        resource_matches = "*" in matched_resources or resource in matched_resources
                                        if not resource_matches:
                                                continue
                                        if not matched_rule["apiGroups"] or not existing_api_groups:
                                                continue
                                        api_group_matches = (
                                            ("*" in matched_rule["apiGroups"])
                                            or ("*" in existing_api_groups)
                                            or bool(set(matched_rule["apiGroups"]).intersection(existing_api_groups))
                                        )
                                        if not api_group_matches:
                                                continue
                                        revoke_verbs.extend(matched_rule["verbs"])
                                new_verbs = self._remove_revoked_verbs(existing_verbs, revoke_verbs)
                                if new_verbs:
                                        remaining_by_verbs.setdefault(tuple(new_verbs), []).append(resource)
                        for verbs_key, resources in sorted(remaining_by_verbs.items()):
                                updated_rule = dict(rule)
                                updated_rule["resources"] = sorted(resources)
                                updated_rule["verbs"] = list(verbs_key)
                                remaining_rules.append(updated_rule)
                return existing_rules, remaining_rules

        def _revoke_rbac(self, permissionfile, sa, namespace, kubeconfig):
                """Revoke permissions from permission file.

                Flow:
                1) Select target role (<sa>-update first, then <sa> fallback).
                2) Build rewritten rules after revoke match/removal.
                3) Apply updated role or delete empty role/binding.
                4) Rebuild perms configmap from live role state.
                """
                permission_map = self._load_permission_data(permissionfile)
                revoke_rule_list, _ = self._parse_permission_rules(permission_map)
                revoke_norm = self._normalize_rule_list(revoke_rule_list)

                # Supported revoke behavior:
                # - Supports resource rules (apiGroups/resources/verbs).
                # - Supports URL rules (nonResourceURLs/verbs).
                # - Does not do name-specific matching inside a resource type.
                # - Partial verb revoke is supported (e.g. delete removed, get retained).
                # - Multi-resource rules are split by resource when only a subset is revoked.
                # - If verbs contain '*', revoking any verb drops that matched scope.
                # - Revoke target preference: <sa>-update first, then fallback to base <sa>.
                # - ConfigMap is rebuilt from live role state after revoke.

                # Prefer revoking from <sa>-update. If that role doesn't change, try base <sa>.
                update_role = self._fetch_clusterrole(sa + "-update", kubeconfig)
                base_role = self._fetch_clusterrole(sa, kubeconfig)
                if update_role is None and base_role is None:
                        print(f"No clusterrole {sa!r} or {sa + '-update'!r} found; nothing to revoke.")
                        return False

                candidates = []
                if update_role is not None:
                        existing_rules, remaining_rules = self._build_remaining_rules_for_role(update_role, revoke_norm)
                        candidates.append((sa + "-update", update_role, existing_rules, remaining_rules))
                if base_role is not None:
                        existing_rules, remaining_rules = self._build_remaining_rules_for_role(base_role, revoke_norm)
                        candidates.append((sa, base_role, existing_rules, remaining_rules))

                target = next(
                    (
                        (name, role_obj, remaining_rules)
                        for name, role_obj, existing_rules, remaining_rules in candidates
                        if self._rules_changed(existing_rules, remaining_rules)
                    ),
                    None,
                )
                if target is None:
                        present = [n for n, _, _, _ in candidates]
                        # No role changed means revoke input did not remove any granted verbs.
                        print(f"Nothing to revoke: permission file didn't strip any verbs from {' or '.join(present)}.")
                        return False

                role_name, role_obj, remaining_rules = target
                rolebinding_name = role_name

                # Rewrite target role with surviving rules, or delete role/binding if empty.
                role_filename = role_name + "-role.yaml"
                if remaining_rules:
                        labels = role_obj.get("metadata", {}).get("labels")
                        annotations = role_obj.get("metadata", {}).get("annotations") or {}
                        annotations = {
                                k: v for k, v in annotations.items()
                                if k != "kubectl.kubernetes.io/last-applied-configuration"
                        }
                        metadata = {"name": role_name}
                        if labels:
                                metadata["labels"] = labels
                        if annotations:
                                metadata["annotations"] = annotations
                        role = {
                                "apiVersion": "rbac.authorization.k8s.io/v1",
                                "kind": "ClusterRole",
                                "metadata": metadata,
                                "rules": remaining_rules,
                        }
                        with open(role_filename, "w", encoding="utf-8") as f:
                                yaml.dump(role, f, default_flow_style=False)
                        self._kubectl_or_raise("apply -f " + role_filename, kubeconfig)
                else:
                        self._kubectl_or_raise("delete clusterrole " + role_name, kubeconfig)
                        self._kubectl_or_raise("delete clusterrolebinding " + rolebinding_name, kubeconfig)

                # Keep configmap simple and authoritative: rebuild from current roles.
                self._rebuild_perm_configmap_from_roles(sa, namespace, kubeconfig)
                return True

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
        parser.add_argument("action", help="command", choices=['create', 'delete', 'update', 'revoke', 'extract'])
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
        permission_help = "Permissions file - use with update or revoke command. "
        permission_help += "JSON/YAML with top-level 'perms'; revoke matches apiGroups/resources or nonResourceURLs "
        permission_help += "(resourceNames-specific revoke matching is not supported)."
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

        if action in ['update', 'revoke'] and permission_file == '':
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
                    print("Permissions file should be used with update or revoke command.")
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

        if action == "revoke":
                if kubeconfigGenerator._revoke_rbac(permission_file, sa, namespace, kubeconfigString):
                        print("kubeconfig permissions revoked: " + filename)


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

