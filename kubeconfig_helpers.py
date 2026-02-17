"""
Shared helpers for kubeconfig generation.
Used by both provider and consumer kubeconfig flows.

run_cmd: Helpers that run kubectl accept optional run_cmd for dependency injection.
When None, uses run_command. Pass a mock in tests to avoid executing shell commands.
"""
import json
import os
import subprocess
import time
import yaml


def run_command(cmd):
    """Execute a shell command. Returns (stdout_str, stderr_str)."""
    with subprocess.Popen(
        cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True
    ) as proc:
        cmd_out = proc.communicate()
    out = cmd_out[0].decode("utf-8") if cmd_out[0] else ""
    err = cmd_out[1].decode("utf-8") if cmd_out[1] else ""
    return out, err


def create_role_rolebinding(contents, name, kubeconfig, run_cmd=None):
    """Write Role/RoleBinding YAML to file and apply via kubectl."""
    run = run_cmd or run_command
    file_path = os.path.join(os.getcwd(), name)
    with open(file_path, "w", encoding="utf-8") as fp:
        fp.write(yaml.dump(contents))
    cmd = " kubectl apply -f " + file_path + kubeconfig
    run(cmd)


def create_service_account(sa, namespace, kubeconfig, run_cmd=None):
    """Create a ServiceAccount in the given namespace."""
    run = run_cmd or run_command
    cmd = " kubectl create sa " + sa + " -n " + namespace + kubeconfig
    run(cmd)


def create_sa_token_secret(sa, namespace, kubeconfig, run_cmd=None):
    """Create a ServiceAccount token secret. Returns output string from kubectl create."""
    run = run_cmd or run_command
    secret = {
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {
            "name": sa,
            "namespace": namespace,
            "annotations": {"kubernetes.io/service-account.name": sa},
        },
        "type": "kubernetes.io/service-account-token",
    }
    secret_name = sa + "-secret.yaml"
    file_path = os.path.join(os.getcwd(), secret_name)
    with open(file_path, "w", encoding="utf-8") as fp:
        fp.write(yaml.dump(secret))
    created = False
    count = 0
    while not created and count < 5:
        out, err = run(" kubectl create -f " + file_path + kubeconfig)
        if out:
            out_str = out if isinstance(out, str) else out.decode("utf-8")
            if "created" in out_str:
                created = True
        else:
            time.sleep(2)
            count += 1
    if not created and count >= 5:
        raise SystemExit(err)
    return out if isinstance(out, str) else (out.decode("utf-8") if out else "")


def build_kubeconfig_dict(
    sa, namespace, token, ca_cert, server, cluster_name=None
):
    """Build kubeconfig structure as a dict (no I/O).

    TODO: Use ca_cert for TLS (certificate-authority-data) instead of insecure-skip-tls-verify.
    """
    _ = ca_cert  # Reserved for TODO above
    context_name = cluster_name if cluster_name else sa
    return {
        "apiVersion": "v1",
        "kind": "Config",
        "users": [{"name": sa, "user": {"token": token}}],
        "clusters": [
            {
                "name": cluster_name if cluster_name else sa,
                "cluster": {
                    "server": server,
                    "insecure-skip-tls-verify": True,
                },
            }
        ],
        "contexts": [
            {
                "name": context_name,
                "context": {
                    "cluster": cluster_name if cluster_name else sa,
                    "user": sa,
                    "namespace": namespace,
                },
            }
        ],
        "current-context": context_name,
    }


def store_kubeconfig_configmap(
    sa, namespace, filename, kubeconfig_str, kubeconfig, run_cmd=None
):
    """Store kubeconfig in a ConfigMap in the namespace."""
    run = run_cmd or run_command
    configmap_name = sa
    file_path = os.path.join(os.getcwd(), filename)
    with open(file_path, "w", encoding="utf-8") as fp:
        fp.write(kubeconfig_str)
    created = False
    while not created:
        cmd = (
            "kubectl create configmap "
            + configmap_name
            + " -n "
            + namespace
            + " --from-file="
            + file_path
            + kubeconfig
        )
        run(cmd)
        get_cmd = (
            "kubectl get configmap " + configmap_name
            + " -n " + namespace + kubeconfig
        )
        output, _ = run(get_cmd)
        output_str = (
            output if isinstance(output, str)
            else (output.decode("utf-8") if output else "")
        )
        if "Error from server (NotFound)" not in output_str:
            created = True
        else:
            time.sleep(2)


def extract_token_and_build_kubeconfig(
    sa, namespace, filename, api_server_ip, kubeconfig,
    cluster_name=None, run_cmd=None
):
    """Extract token from secret, determine server URL, build kubeconfig and store in ConfigMap."""
    run = run_cmd or run_command
    # Wait for token in secret
    token = ""
    while not token:
        cmd = " kubectl describe secret " + sa + " -n " + namespace + kubeconfig
        out, _ = run(cmd)
        out_str = out if isinstance(out, str) else (out.decode("utf-8") if out else "")
        for line in out_str.split("\n"):
            if "token" in line:
                parts = line.split(":")
                if len(parts) > 1:
                    token = parts[1].strip()
                    break
        if not token:
            time.sleep(2)
    # Get CA cert (TODO: use in build_kubeconfig_dict for TLS verification)
    cmd = " kubectl get secret " + sa + " -n " + namespace + " -o json " + kubeconfig
    out, _ = run(cmd)
    out_str = out if isinstance(out, str) else (out.decode("utf-8") if out else "{}")
    json_out = json.loads(out_str)
    ca_cert = json_out.get("data", {}).get("ca.crt", "").strip()
    # Server URL
    if api_server_ip:
        server = api_server_ip if "https" in api_server_ip else "https://" + api_server_ip
    else:
        cmd = (
            "kubectl -n default get endpoints kubernetes " + kubeconfig
            + " | awk '{print $2}' | grep -v ENDPOINTS"
        )
        out, _ = run(cmd)
        server = ("https://" + out.strip()) if out and out.strip() else ""
    # Build and store kubeconfig
    kubeconfig_dict = build_kubeconfig_dict(sa, namespace, token, ca_cert, server, cluster_name)
    kubeconfig_str = json.dumps(kubeconfig_dict)
    store_kubeconfig_configmap(sa, namespace, filename, kubeconfig_str, kubeconfig, run_cmd)


def generate_kubeconfig(
    sa, namespace, filename, api_server_ip="", kubeconfig="",
    cluster_name=None, run_cmd=None
):
    """Create SA, token secret, extract kubeconfig, store in ConfigMap."""
    create_service_account(sa, namespace, kubeconfig, run_cmd)
    out = create_sa_token_secret(sa, namespace, kubeconfig, run_cmd)
    if "secret/" + sa + " created" in (out if isinstance(out, str) else ""):
        extract_token_and_build_kubeconfig(
            sa, namespace, filename, api_server_ip, kubeconfig, cluster_name, run_cmd
        )
