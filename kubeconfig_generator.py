#!/usr/bin/env python3
"""
Kubeconfig generator: creates provider and consumer kubeconfigs.
Use provider-kubeconfig.py for backward compatibility
"""
import argparse
import json
import os
import sys

from logging.config import dictConfig

from kubeconfig_helpers import (
    create_role_rolebinding,
    extract_token_and_build_kubeconfig,
    generate_kubeconfig,
    run_command,
)

dictConfig({
    "version": 1,
    "formatters": {"default": {"format": "[%(asctime)s] %(levelname)s in %(module)s: %(message)s"}},
    "handlers": {
        "file.handler": {
            "class": "logging.handlers.RotatingFileHandler",
            "filename": "kubeconfig-generator.log",
            "maxBytes": 10000000,
            "backupCount": 5,
            "level": "DEBUG",
        },
    },
    "root": {"level": "INFO", "handlers": ["file.handler"]},
})


# --- Provider RBAC ---

def apply_provider_rbac(sa, namespace, kubeconfig, run_cmd=None):
    """Apply ClusterRole and ClusterRoleBinding for provider (full platform operator permissions)."""
    run = run_cmd or run_command

    all_resources = []
    rule_list = [
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
        {"apiGroups": [""], "resources": ["secrets", "serviceaccounts", "configmaps", "events", "persistentvolumeclaims", "serviceaccounts/token", "services", "services/proxy", "endpoints"], "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"]},
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

    for r in rule_list:
        all_resources.extend(r.get("resources", []))

    role = {"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRole", "metadata": {"name": sa, "namespace": namespace}, "rules": rule_list}
    create_role_rolebinding(role, sa + "-role.yaml", kubeconfig, run_cmd)

    role_binding = {
        "apiVersion": "rbac.authorization.k8s.io/v1",
        "kind": "ClusterRoleBinding",
        "metadata": {"name": sa, "namespace": namespace},
        "subjects": [{"kind": "ServiceAccount", "name": sa, "namespace": namespace, "apiGroup": ""}],
        "roleRef": {"kind": "ClusterRole", "name": sa, "apiGroup": "rbac.authorization.k8s.io"},
    }
    create_role_rolebinding(role_binding, sa + "-rolebinding.yaml", kubeconfig, run_cmd)

    cfg_map_filename = sa + "-perms.txt"
    all_resources = sorted(list(set(all_resources)))
    with open(cfg_map_filename, "w", encoding="utf-8") as fp:
        fp.write(str(all_resources))
    run(
        "kubectl create configmap " + sa + "-perms -n " + namespace
        + " --from-file=" + cfg_map_filename + kubeconfig
    )


def update_provider_rbac(permission_file, sa, namespace, kubeconfig, run_cmd=None):
    """Add permissions from JSON file to provider (update command)."""
    run = run_cmd or run_command

    with open(permission_file, "r", encoding="utf-8") as fp:
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
                        rule_group["nonResourceURLs"] = [parts[0].strip()]
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

    role = {"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRole", "metadata": {"name": sa + "-update", "namespace": namespace}, "rules": rule_list}
    create_role_rolebinding(role, sa + "-update-role.yaml", kubeconfig, run_cmd)

    role_binding = {
        "apiVersion": "rbac.authorization.k8s.io/v1",
        "kind": "ClusterRoleBinding",
        "metadata": {"name": sa + "-update", "namespace": namespace},
        "subjects": [{"kind": "ServiceAccount", "name": sa, "namespace": namespace, "apiGroup": ""}],
        "roleRef": {"kind": "ClusterRole", "name": sa + "-update", "apiGroup": "rbac.authorization.k8s.io"},
    }
    create_role_rolebinding(role_binding, sa + "-update-rolebinding.yaml", kubeconfig, run_cmd)

    cfg_map_name = sa + "-perms"
    cfg_map_filename = sa + "-perms.txt"
    cmd = "kubectl get configmap " + cfg_map_name + " -o json -n " + namespace + kubeconfig
    out, _ = run(cmd)
    kubeplus_perms = []
    if out:
        json_op = json.loads(out)
        perms_str = json_op.get("data", {}).get(cfg_map_filename, "")
        for p in perms_str.replace("'", "").replace("[", "").replace("]", "").split(","):
            p = p.strip()
            if p:
                kubeplus_perms.append(p)
    new_resources.extend(kubeplus_perms)

    run("kubectl delete configmap " + cfg_map_name + " -n " + namespace + kubeconfig)
    new_resources = sorted(list(set(new_resources)))
    with open(cfg_map_filename, "w", encoding="utf-8") as fp:
        fp.write(str(new_resources))
    run(
        "kubectl create configmap " + cfg_map_name + " -n " + namespace
        + " --from-file=" + cfg_map_filename + kubeconfig
    )


# --- Consumer RBAC ---

def apply_consumer_rbac(sa, namespace, kubeconfig, run_cmd=None):
    """Apply ClusterRole and ClusterRoleBinding for consumer (read + apps + impersonate + portforward)."""
    run = run_cmd or run_command

    role = {
        "apiVersion": "rbac.authorization.k8s.io/v1",
        "kind": "ClusterRole",
        "metadata": {"name": sa, "namespace": namespace},
        "rules": [
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
        ],
    }
    create_role_rolebinding(role, sa + "-role-impersonate.yaml", kubeconfig, run_cmd)

    role_binding = {
        "apiVersion": "rbac.authorization.k8s.io/v1",
        "kind": "ClusterRoleBinding",
        "metadata": {"name": sa, "namespace": namespace},
        "subjects": [{"kind": "ServiceAccount", "name": sa, "namespace": namespace, "apiGroup": ""}],
        "roleRef": {"kind": "ClusterRole", "name": sa, "apiGroup": "rbac.authorization.k8s.io"},
    }
    create_role_rolebinding(role_binding, sa + "-rolebinding-impersonate.yaml", kubeconfig, run_cmd)

    cfg_map_filename = sa + "-perms.txt"
    all_resources = ["*", "deployments", "daemonsets", "pods/portforward", "users", "groups", "serviceaccounts"]
    all_resources = sorted(list(set(all_resources)))
    with open(cfg_map_filename, "w", encoding="utf-8") as fp:
        fp.write(str(all_resources))
    run(
        "kubectl create configmap " + sa + "-perms -n " + namespace
        + " --from-file=" + cfg_map_filename + kubeconfig
    )


# --- Main CLI ---

def main():
    """Parse args and dispatch to create, delete, update, or extract."""
    kubeconfig_path = os.path.join(os.getenv("HOME", ""), ".kube", "config")
    parser = argparse.ArgumentParser(
        description="Generate provider or consumer kubeconfig for KubePlus."
    )
    parser.add_argument("action", help="command", choices=["create", "delete", "update", "extract"])
    parser.add_argument(
        "namespace",
        help="Namespace in which the ServiceAccount will be created. "
        "For provider: typically the KubePlus install namespace (e.g. default). "
        "For consumer (-c): the namespace where the consumer SA lives.",
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

    if pargs.kubeconfig:
        kubeconfig_path = pargs.kubeconfig
    kubeconfig_string = " --kubeconfig=" + kubeconfig_path

    api_server_ip = pargs.apiserverurl or ""
    permission_file = pargs.permissionfile or ""
    cluster_name = pargs.clustername or ""

    if action == "update" and not permission_file:
        print("Permission file missing. Please provide -p/--permissionfile.")
        print(permission_help)
        sys.exit(1)

    sa = "kubeplus-saas-provider"
    if pargs.consumer:
        sa = pargs.consumer

    filename = pargs.filename or sa
    if not filename.endswith(".json"):
        filename += ".json"

    if action == "create":
        if permission_file:
            print("Permissions file should be used with update command.")
            sys.exit(1)
        out, err = run_command("kubectl get ns " + namespace + kubeconfig_string)
        if "not found" in (out or "") or "not found" in (err or ""):
            run_command("kubectl create ns " + namespace + kubeconfig_string)
        run_command("kubectl label --overwrite=true ns " + namespace + " managedby=kubeplus " + kubeconfig_string)

        if sa == "kubeplus-saas-provider":
            generate_kubeconfig(sa, namespace, filename, api_server_ip, kubeconfig_string, cluster_name or None)
            apply_provider_rbac(sa, namespace, kubeconfig_string)
            print("Provider kubeconfig created: " + filename)
        else:
            generate_kubeconfig(sa, namespace, filename, api_server_ip, kubeconfig_string, cluster_name or None)
            apply_consumer_rbac(sa, namespace, kubeconfig_string)
            print("Consumer kubeconfig created: " + filename)

    elif action == "extract":
        extract_token_and_build_kubeconfig(
            sa, namespace, filename, api_server_ip, kubeconfig_string,
            cluster_name or None
        )
        print("Kubeconfig extracted: " + filename)

    elif action == "update":
        update_provider_rbac(permission_file, sa, namespace, kubeconfig_string)
        print("Kubeconfig permissions updated: " + filename)

    elif action == "delete":
        cwd = os.getcwd()
        run_command("kubectl delete sa " + sa + " -n " + namespace + kubeconfig_string)
        run_command("kubectl delete configmap " + sa + " -n " + namespace + kubeconfig_string)
        run_command("kubectl delete clusterrole " + sa + kubeconfig_string)
        run_command("kubectl delete clusterrolebinding " + sa + kubeconfig_string)
        run_command("kubectl delete clusterrole " + sa + "-update" + kubeconfig_string)
        run_command("kubectl delete clusterrolebinding " + sa + "-update" + kubeconfig_string)
        run_command("kubectl delete configmap " + sa + "-perms -n " + namespace + kubeconfig_string)
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


if __name__ == "__main__":
    main()
