#!/usr/bin/env python3
import argparse
import subprocess
import json
import sys
import os

POLICY_NAME = "restrict-cross-ns-traffic"

def run_command(cmd):
    """Run a shell command and return stdout and stderr as strings."""
    proc = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
    out, err = proc.communicate()
    return out.decode('utf-8'), err.decode('utf-8')

def networkpolicy_exists(policy_name, namespace, kubeconfig):
    cmd = f"kubectl get networkpolicy {policy_name} -n {namespace} {kubeconfig}"
    out, err = run_command(cmd)
    if "NotFound" in err or out.strip() == "":
        return False
    return True

def add_namespace_to_networkpolicy(policy_name, namespace, allowed_namespace, kubeconfig):
    """
    Add an ingress rule for allowed_namespace to the NetworkPolicy in the given namespace,
    preserving any existing ingress rules.
    """
    # Retrieve the current network policy as JSON.
    cmd = f"kubectl get networkpolicy {policy_name} -n {namespace} {kubeconfig} -o json"
    out, err = run_command(cmd)
    if err and not out:
        return "", f"Error retrieving network policy: {err}"
    try:
        policy = json.loads(out)
    except Exception as e:
        return "", f"Failed to parse JSON: {e}"

    ingress = policy.get("spec", {}).get("ingress", [])
    if not isinstance(ingress, list):
        ingress = []

    # Check if allowed_namespace is already allowed.
    for rule in ingress:
        if "from" in rule:
            for src in rule["from"]:
                if "namespaceSelector" in src:
                    expressions = src["namespaceSelector"].get("matchExpressions", [])
                    for expr in expressions:
                        if (expr.get("key") == "kubernetes.io/metadata.name" and
                            expr.get("operator") == "In" and
                            allowed_namespace in expr.get("values", [])):
                            return f"Namespace '{allowed_namespace}' is already allowed in namespace '{namespace}'.", ""

    # Create a new ingress rule for allowed_namespace.
    new_rule = {
        "from": [
            {
                "namespaceSelector": {
                    "matchExpressions": [
                        {
                            "key": "kubernetes.io/metadata.name",
                            "operator": "In",
                            "values": [allowed_namespace]
                        }
                    ]
                }
            }
        ]
    }
    ingress.append(new_rule)

    patch = {"spec": {"ingress": ingress}}
    patch_str = json.dumps(patch)
    patch_cmd = f"kubectl patch networkpolicy {policy_name} -n {namespace} {kubeconfig} --type merge -p '{patch_str}'"
    return run_command(patch_cmd)

def remove_namespace_from_networkpolicy(policy_name, namespace, deny_namespace, kubeconfig):
    """
    Remove any ingress rule that allows traffic from deny_namespace while leaving other rules intact.
    """
    # Retrieve the current network policy as JSON.
    cmd = f"kubectl get networkpolicy {policy_name} -n {namespace} {kubeconfig} -o json"
    out, err = run_command(cmd)
    if err and not out:
        return "", f"Error retrieving network policy: {err}"
    try:
        policy = json.loads(out)
    except Exception as e:
        return "", f"Failed to parse JSON: {e}"

    ingress = policy.get("spec", {}).get("ingress", [])
    new_ingress = []

    for rule in ingress:
        if "from" in rule:
            new_from = []
            for src in rule["from"]:
                if "namespaceSelector" in src:
                    expressions = src["namespaceSelector"].get("matchExpressions", [])
                    remove_entry = any(
                        expr.get("key") == "kubernetes.io/metadata.name" and
                        expr.get("operator") == "In" and
                        deny_namespace in expr.get("values", [])
                        for expr in expressions
                    )
                    if not remove_entry:
                        new_from.append(src)
                else:
                    new_from.append(src)
            if new_from:
                new_rule = rule.copy()
                new_rule["from"] = new_from
                new_ingress.append(new_rule)
        else:
            new_ingress.append(rule)

    if new_ingress == ingress:
        return f"No ingress rule for namespace '{deny_namespace}' found in namespace '{namespace}'.", ""

    patch = {"spec": {"ingress": new_ingress}}
    patch_str = json.dumps(patch)
    patch_cmd = f"kubectl patch networkpolicy {policy_name} -n {namespace} {kubeconfig} --type merge -p '{patch_str}'"
    return run_command(patch_cmd)

def do_allow(ns1, ns2, kubeconfig):
    # For ns1, add allowed ingress rule for ns2.
    print(f"Adding allowed ingress for namespace '{ns2}' in NetworkPolicy '{POLICY_NAME}' of namespace '{ns1}'.")
    out, err = add_namespace_to_networkpolicy(POLICY_NAME, ns1, ns2, kubeconfig)
    print(out)
    if err:
        print("Error:", err, file=sys.stderr)
        sys.exit(1)

    # For ns2, add allowed ingress rule for ns1.
    print(f"Adding allowed ingress for namespace '{ns1}' in NetworkPolicy '{POLICY_NAME}' of namespace '{ns2}'.")
    out, err = add_namespace_to_networkpolicy(POLICY_NAME, ns2, ns1, kubeconfig)
    print(out)
    if err:
        print("Error:", err, file=sys.stderr)
        sys.exit(1)

def do_deny(ns1, ns2, kubeconfig):
    # For ns1, remove the ingress rule that allows traffic from ns2.
    print(f"Removing allowed ingress for namespace '{ns2}' in NetworkPolicy '{POLICY_NAME}' of namespace '{ns1}'.")
    out, err = remove_namespace_from_networkpolicy(POLICY_NAME, ns1, ns2, kubeconfig)
    print(out)
    if err:
        print("Error:", err, file=sys.stderr)
        sys.exit(1)

    # For ns2, remove the ingress rule that allows traffic from ns1.
    print(f"Removing allowed ingress for namespace '{ns1}' in NetworkPolicy '{POLICY_NAME}' of namespace '{ns2}'.")
    out, err = remove_namespace_from_networkpolicy(POLICY_NAME, ns2, ns1, kubeconfig)
    print(out)
    if err:
        print("Error:", err, file=sys.stderr)
        sys.exit(1)

def main():
    # If the first argument is not 'allow' or 'deny', infer it from the executable name.
    valid_commands = {"allow", "deny"}
    if len(sys.argv) > 1 and sys.argv[1] not in valid_commands:
        prog = os.path.basename(sys.argv[0]).lower()
        if "deny" in prog:
            sys.argv.insert(1, "deny")
        elif "allow" in prog:
            sys.argv.insert(1, "allow")

    parser = argparse.ArgumentParser(
        description="Manage bidirectional NetworkPolicy rules for cross-namespace traffic."
    )
    parser.add_argument("-k", "--kubeconfig", help="Path to kubeconfig file", default="")

    subparsers = parser.add_subparsers(dest="command", required=True, help="Subcommand to run: 'allow' or 'deny'.")

    # Subparser for the 'allow' command.
    allow_parser = subparsers.add_parser("allow", help="Allow traffic between two namespaces.")
    allow_parser.add_argument("ns1", help="First namespace (e.g., hs1)")
    allow_parser.add_argument("ns2", help="Second namespace (e.g., hs2)")

    # Subparser for the 'deny' command.
    deny_parser = subparsers.add_parser("deny", help="Deny traffic between two namespaces by removing specific allow rules.")
    deny_parser.add_argument("ns1", help="First namespace (e.g., hs1)")
    deny_parser.add_argument("ns2", help="Second namespace (e.g., hs2)")

    args = parser.parse_args()
    kubeconfig = f"--kubeconfig={args.kubeconfig}" if args.kubeconfig else ""

    # Verify the NetworkPolicy exists in both namespaces.
    if not networkpolicy_exists(POLICY_NAME, args.ns1, kubeconfig):
        print(f"Error: NetworkPolicy '{POLICY_NAME}' does not exist in namespace '{args.ns1}'.", file=sys.stderr)
        sys.exit(1)
    if not networkpolicy_exists(POLICY_NAME, args.ns2, kubeconfig):
        print(f"Error: NetworkPolicy '{POLICY_NAME}' does not exist in namespace '{args.ns2}'.", file=sys.stderr)
        sys.exit(1)

    if args.command == "allow":
        do_allow(args.ns1, args.ns2, kubeconfig)
    elif args.command == "deny":
        do_deny(args.ns1, args.ns2, kubeconfig)
    else:
        print("Invalid command.", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
