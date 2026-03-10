"""
Tests for provider-kubeconfig.py: CLI, structure validation, integration tests
that verify non-empty fields, service account creation, and CLI flags in output.
"""
import json
import os
import subprocess
import sys
import unittest

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))


def _run_command(cmd):
    """Run shell command, return (stdout, stderr)."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout or "", result.stderr or ""


def _cluster_available(kubeconfig=""):
    """True if kubectl can reach a cluster."""
    path = os.path.expanduser(kubeconfig).strip() if kubeconfig else ""
    k_opt = " --kubeconfig=" + path if path else ""
    out, err = _run_command("kubectl get ns default" + k_opt)
    return out and "default" in out and "NotFound" not in (err or "")


SCRIPT = "provider-kubeconfig.py"


class TestCli(unittest.TestCase):
    """provider-kubeconfig.py exposes expected CLI (actions and flags)."""

    def test_help_shows_actions_and_flags(self):
        root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        proc = subprocess.run(
            [sys.executable, os.path.join(root, SCRIPT), "--help"],
            capture_output=True, text=True, cwd=root,
        )
        self.assertEqual(proc.returncode, 0)
        out = proc.stdout or ""
        for x in ["create", "delete", "update", "extract", "-k", "-s", "-c", "namespace"]:
            self.assertIn(x, out)


class TestKubeconfigIntegration(unittest.TestCase):
    """
    Verify provider-kubeconfig.py produces valid output with non-empty fields.
    Verifies service account creation and that CLI flags are reflected.
    Skips if no cluster reachable.
    """

    @classmethod
    def setUpClass(cls):
        raw = os.environ.get("KUBECONFIG", "")
        cls.kubeconfig = os.path.expanduser(raw) if raw else ""
        cls.kubeconfig_arg = ["-k", cls.kubeconfig] if cls.kubeconfig else []
        cls.has_cluster = _cluster_available(cls.kubeconfig)
        cls.kubeconfig_flag = " --kubeconfig=" + cls.kubeconfig if cls.kubeconfig else ""

    def _create_and_get_kubeconfig(self, root, ns, sa="kubeplus-saas-provider", extra_args=None):
        """Run create, return (kubeconfig_dict, proc). Caller must delete to cleanup."""
        create_args = ["create", ns]
        if sa != "kubeplus-saas-provider":
            create_args += ["-c", sa]
        if extra_args:
            create_args += extra_args
        proc = subprocess.run(
            [sys.executable, os.path.join(root, SCRIPT)] + create_args + self.kubeconfig_arg,
            cwd=root, capture_output=True, text=True, timeout=120,
        )
        if proc.returncode != 0:
            return None, proc
        filename = sa + ".json"
        kubeconfig_path = os.path.join(root, filename)
        if not os.path.exists(kubeconfig_path):
            return None, proc
        with open(kubeconfig_path, "r", encoding="utf-8") as fp:
            cfg = json.load(fp)
        return cfg, proc

    def _delete_for_cleanup(self, root, ns, sa="kubeplus-saas-provider"):
        delete_args = ["delete", ns]
        if sa != "kubeplus-saas-provider":
            delete_args += ["-c", sa]
        subprocess.run(
            [sys.executable, os.path.join(root, SCRIPT)] + delete_args + self.kubeconfig_arg,
            cwd=root, capture_output=True, timeout=60,
        )

    def _assert_kubeconfig_valid(self, cfg, expected_server=None, expected_cluster_name=None, expected_namespace=None):
        self.assertIsNotNone(cfg)
        self.assertEqual(cfg.get("apiVersion"), "v1")
        self.assertEqual(cfg.get("kind"), "Config")
        self.assertTrue(cfg.get("users"), "users should be non-empty")
        self.assertTrue(cfg.get("clusters"), "clusters should be non-empty")
        self.assertTrue(cfg.get("contexts"), "contexts should be non-empty")
        token = cfg["users"][0].get("user", {}).get("token")
        self.assertTrue(token, "token should be non-empty")
        server = cfg["clusters"][0].get("cluster", {}).get("server")
        self.assertTrue(server, "cluster server should be non-empty")
        if expected_server:
            self.assertEqual(server, expected_server)
        if expected_cluster_name:
            ctx_name = cfg.get("current-context") or (cfg["contexts"][0].get("name") if cfg["contexts"] else "")
            self.assertEqual(ctx_name, expected_cluster_name)
        if expected_namespace:
            ctx_ns = cfg["contexts"][0].get("context", {}).get("namespace", "")
            self.assertEqual(ctx_ns, expected_namespace)

    def _sa_exists(self, sa, ns):
        out, err = _run_command("kubectl get sa " + sa + " -n " + ns + self.kubeconfig_flag)
        return out and sa in out and "NotFound" not in (err or "")

    def test_provider_kubeconfig_has_nonempty_fields(self):
        if not self.has_cluster:
            self.skipTest("No cluster reachable (set KUBECONFIG)")
        root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        ns = "kubeplus-test-prov-" + str(os.getpid())
        try:
            cfg, proc = self._create_and_get_kubeconfig(root, ns)
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(cfg, expected_namespace=ns)
            self.assertTrue(self._sa_exists("kubeplus-saas-provider", ns))
        finally:
            self._delete_for_cleanup(root, ns)

    def test_consumer_kubeconfig_has_namespace_in_context(self):
        if not self.has_cluster:
            self.skipTest("No cluster reachable (set KUBECONFIG)")
        root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        ns = "kubeplus-test-cons-" + str(os.getpid())
        consumer_sa = "test-consumer-sa"
        try:
            cfg, proc = self._create_and_get_kubeconfig(root, ns, sa=consumer_sa)
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(cfg, expected_namespace=ns)
            self.assertTrue(self._sa_exists(consumer_sa, ns))
        finally:
            self._delete_for_cleanup(root, ns, sa=consumer_sa)

    def test_cli_flags_reflected_in_kubeconfig(self):
        if not self.has_cluster:
            self.skipTest("No cluster reachable (set KUBECONFIG)")
        root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        ns = "kubeplus-test-flags-" + str(os.getpid())
        test_server = "https://api.example.com:6443"
        test_cluster = "my-test-cluster"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                root, ns,
                extra_args=["-s", test_server, "-x", test_cluster],
            )
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(
                cfg,
                expected_server=test_server,
                expected_cluster_name=test_cluster,
                expected_namespace=ns,
            )
        finally:
            self._delete_for_cleanup(root, ns)


if __name__ == "__main__":
    unittest.main()
