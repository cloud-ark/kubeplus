"""
Tests for kubeconfig generation: equivalence with provider-kubeconfig, CLI parity, provider/consumer split.
"""
import json
import os
import subprocess
import sys
import time
import unittest
from unittest.mock import patch, MagicMock

# Add project root so "from kubeconfig_helpers import ..." works when run from any directory
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from kubeconfig_helpers import build_kubeconfig_dict, run_command
from kubeconfig_generator import apply_consumer_rbac, apply_provider_rbac


def _cluster_available(kubeconfig=""):
    """True if kubectl can reach a cluster."""
    path = os.path.expanduser(kubeconfig).strip() if kubeconfig else ""
    k_opt = " --kubeconfig=" + path if path else ""
    out, err = run_command("kubectl get ns default" + k_opt)
    return out and "default" in out and "NotFound" not in (err or "")


def _norm_rules(rules):
    """Normalize rules for order-independent comparison."""
    return sorted(json.dumps(r, sort_keys=True) for r in rules)


# --- build_kubeconfig_dict ---

class TestBuildKubeconfigDict(unittest.TestCase):
    """Kubeconfig structure is valid (users, clusters, contexts)."""

    def test_produces_valid_kubeconfig_structure(self):
        out = build_kubeconfig_dict(
            sa="test-sa", namespace="default", token="x", ca_cert="",
            server="https://api.example.com", cluster_name="my-cluster",
        )
        self.assertEqual(out["apiVersion"], "v1")
        self.assertEqual(out["kind"], "Config")
        self.assertEqual(out["users"][0]["name"], "test-sa")
        self.assertEqual(out["users"][0]["user"]["token"], "x")
        self.assertEqual(out["clusters"][0]["cluster"]["server"], "https://api.example.com")
        self.assertEqual(out["current-context"], "my-cluster")


# --- Provider vs consumer split ---

class TestProviderConsumerSplit(unittest.TestCase):
    """Provider has platform APIs; consumer has read + apps + impersonate only."""

    @patch("kubeconfig_generator.create_role_rolebinding")
    def test_provider_has_platform_apis_consumer_does_not(self, mock_create):
        apply_provider_rbac("prov", "ns", " ", run_cmd=MagicMock(return_value=("", "")))
        apply_consumer_rbac("cons", "ns", " ", run_cmd=MagicMock(return_value=("", "")))
        prov_role = mock_create.call_args_list[0][0][0]
        cons_role = mock_create.call_args_list[2][0][0]
        prov_groups = {g for r in prov_role["rules"] for g in (r.get("apiGroups", []) or [])}
        cons_groups = {g for r in cons_role["rules"] for g in (r.get("apiGroups", []) or [])}

        self.assertIn("platformapi.kubeplus", prov_groups)
        self.assertNotIn("platformapi.kubeplus", cons_groups)
        self.assertIn("workflows.kubeplus", prov_groups)
        self.assertNotIn("workflows.kubeplus", cons_groups)


# --- CLI parity ---

class TestCliParity(unittest.TestCase):
    """Both scripts expose the same CLI (actions and flags)."""

    def test_same_actions_and_flags(self):
        root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        for script in ["provider-kubeconfig.py", "kubeconfig_generator.py"]:
            proc = subprocess.run(
                [sys.executable, os.path.join(root, script), "--help"],
                capture_output=True, text=True, cwd=root,
            )
            self.assertEqual(proc.returncode, 0)
            out = proc.stdout or ""
            for x in ["create", "delete", "update", "extract", "-k", "-s", "-c", "namespace"]:
                self.assertIn(x, out)


# --- Full integration equivalence (requires live cluster) ---

class TestFullIntegrationEquivalence(unittest.TestCase):
    """
    provider-kubeconfig and kubeconfig_generator produce identical ClusterRole rules
    for both provider and consumer. Skips if no cluster reachable.
    """

    @classmethod
    def setUpClass(cls):
        raw = os.environ.get("KUBECONFIG", "")
        cls.kubeconfig = os.path.expanduser(raw) if raw else ""
        cls.kubeconfig_flag = " --kubeconfig=" + cls.kubeconfig if cls.kubeconfig else ""
        cls.kubeconfig_arg = ["-k", cls.kubeconfig] if cls.kubeconfig else []
        cls.has_cluster = _cluster_available(cls.kubeconfig)

    def _run_create_delete_capture_rules(self, root, ns, consumer_sa=None):
        """Run create with provider, capture rules, delete; then create with generator, capture, delete. Return (provider_rules, generator_rules)."""
        sa = consumer_sa if consumer_sa else "kubeplus-saas-provider"
        create_args = ["create", ns] + (["-c", consumer_sa] if consumer_sa else [])

        # Provider create
        proc = subprocess.run(
            [sys.executable, os.path.join(root, "provider-kubeconfig.py")] + create_args + self.kubeconfig_arg,
            cwd=root, capture_output=True, text=True, timeout=120,
        )
        if proc.returncode != 0:
            raise AssertionError(f"provider create failed: {proc.stderr}")
        # ClusterRole may take a moment to appear
        out = ""
        for _ in range(5):
            out, err = run_command("kubectl get clusterrole " + sa + " -o json" + self.kubeconfig_flag)
            if out and "metadata" in out:
                break
            time.sleep(2)
        if not out or "metadata" not in out:
            raise AssertionError(
                f"ClusterRole {sa} not found after provider create. "
                f"stdout: {proc.stdout!r} stderr: {proc.stderr!r} kubectl_err: {err!r}"
            )
        provider_rules = json.loads(out).get("rules", [])

        # Provider delete
        delete_args = ["delete", ns] + (["-c", consumer_sa] if consumer_sa else [])
        subprocess.run(
            [sys.executable, os.path.join(root, "provider-kubeconfig.py")] + delete_args + self.kubeconfig_arg,
            cwd=root, capture_output=True, timeout=60,
        )

        # Generator create
        proc = subprocess.run(
            [sys.executable, os.path.join(root, "kubeconfig_generator.py")] + create_args + self.kubeconfig_arg,
            cwd=root, capture_output=True, text=True, timeout=120,
        )
        if proc.returncode != 0:
            raise AssertionError(f"generator create failed: {proc.stderr}")
        out = ""
        for _ in range(5):
            out, err = run_command("kubectl get clusterrole " + sa + " -o json" + self.kubeconfig_flag)
            if out and "metadata" in out:
                break
            time.sleep(2)
        if not out or "metadata" not in out:
            raise AssertionError(
                f"ClusterRole {sa} not found after generator create. stderr: {err!r}"
            )
        generator_rules = json.loads(out).get("rules", [])

        # Generator delete (cleanup)
        subprocess.run(
            [sys.executable, os.path.join(root, "kubeconfig_generator.py")] + delete_args + self.kubeconfig_arg,
            cwd=root, capture_output=True, timeout=60,
        )

        return provider_rules, generator_rules

    def test_provider_equivalent(self):
        if not self.has_cluster:
            self.skipTest("No cluster reachable (set KUBECONFIG or use default kubeconfig)")
        root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        ns = "kubeplus-equiv-prov-" + str(os.getpid())
        try:
            prov_rules, gen_rules = self._run_create_delete_capture_rules(root, ns)
            self.assertEqual(_norm_rules(prov_rules), _norm_rules(gen_rules), "Provider ClusterRole rules differ")
        finally:
            subprocess.run(
                [sys.executable, os.path.join(root, "kubeconfig_generator.py"), "delete", ns] + self.kubeconfig_arg,
                cwd=root, capture_output=True, timeout=60,
            )

    def test_consumer_equivalent(self):
        if not self.has_cluster:
            self.skipTest("No cluster reachable (set KUBECONFIG or use default kubeconfig)")
        root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        ns = "kubeplus-equiv-cons-" + str(os.getpid())
        consumer_sa = "equiv-test-consumer"
        try:
            prov_rules, gen_rules = self._run_create_delete_capture_rules(root, ns, consumer_sa=consumer_sa)
            self.assertEqual(_norm_rules(prov_rules), _norm_rules(gen_rules), "Consumer ClusterRole rules differ")
        finally:
            subprocess.run(
                [sys.executable, os.path.join(root, "kubeconfig_generator.py"), "delete", ns, "-c", consumer_sa] + self.kubeconfig_arg,
                cwd=root, capture_output=True, timeout=60,
            )


if __name__ == "__main__":
    unittest.main()
