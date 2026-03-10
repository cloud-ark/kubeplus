"""
Tests for provider-kubeconfig.py: CLI, structure validation, integration tests
that verify non-empty fields, service account creation, and CLI flags in output.
"""
import json
import os
import subprocess
import sys
import unittest


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
ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

# All CLI elements that must appear in --help
HELP_ELEMENTS = [
    "create", "delete", "update", "extract",
    "namespace",
    "-k", "--kubeconfig",
    "-s", "--apiserverurl",
    "-f", "--filename",
    "-x", "--clustername",
    "-p", "--permissionfile",
    "-c", "--consumer",
]


class TestCli(unittest.TestCase):
    """provider-kubeconfig.py exposes expected CLI (actions and flags)."""

    def test_help_shows_all_actions_flags_and_namespace(self):
        """--help must show actions, flags, and namespace argument."""
        proc = subprocess.run(
            [sys.executable, os.path.join(ROOT, SCRIPT), "--help"],
            capture_output=True, text=True, cwd=ROOT,
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        out = proc.stdout or ""
        for elem in ["create", "delete", "update", "extract"]:
            self.assertIn(elem, out, f"Action {elem} should appear in help")
        for elem in HELP_ELEMENTS:
            self.assertIn(elem, out, f"Help should mention {elem}")
        self.assertIn("namespace", out.lower())

    def test_update_without_permissionfile_exits_with_error(self):
        """update action without -p exits with code 1."""
        proc = subprocess.run(
            [sys.executable, os.path.join(ROOT, SCRIPT), "update", "default"],
            capture_output=True, text=True, cwd=ROOT,
        )
        self.assertNotEqual(proc.returncode, 0)
        self.assertIn("permission", (proc.stdout or proc.stderr or "").lower())


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

    def setUp(self):
        if not self.has_cluster:
            self.skipTest("No cluster reachable (set KUBECONFIG)")

    def _create_and_get_kubeconfig(self, root, ns, sa="kubeplus-saas-provider", extra_args=None, output_filename=None):
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
        filename = output_filename or (sa + ".json")
        if not filename.endswith(".json"):
            filename += ".json"
        kubeconfig_path = os.path.join(root, filename)
        if not os.path.exists(kubeconfig_path):
            return None, proc
        with open(kubeconfig_path, "r", encoding="utf-8") as fp:
            cfg = json.load(fp)
        return cfg, proc

    def _delete_for_cleanup(self, root, ns, sa="kubeplus-saas-provider", filename=None):
        """Delete k8s resources and local files. Pass filename when -f was used (e.g. custom-provider-kubeconfig)."""
        delete_args = ["delete", ns]
        if sa != "kubeplus-saas-provider":
            delete_args += ["-c", sa]
        if filename:
            delete_args += ["-f", filename]
        subprocess.run(
            [sys.executable, os.path.join(root, SCRIPT)] + delete_args + self.kubeconfig_arg,
            cwd=root, capture_output=True, timeout=60,
        )

    def _assert_kubeconfig_valid(
        self,
        cfg,
        expected_server=None,
        expected_cluster_name=None,
        expected_namespace=None,
        expected_user_name=None,
    ):
        """Assert all kubeconfig fields are non-empty and optionally match expected values."""
        self.assertIsNotNone(cfg, "kubeconfig should not be None")

        # Top-level
        self.assertEqual(cfg.get("apiVersion"), "v1", "apiVersion should be v1")
        self.assertEqual(cfg.get("kind"), "Config", "kind should be Config")
        self.assertTrue(cfg.get("current-context"), "current-context should be non-empty")

        # Users
        self.assertTrue(cfg.get("users"), "users should be non-empty")
        user = cfg["users"][0]
        self.assertTrue(user.get("name"), "users[0].name should be non-empty")
        self.assertTrue(user.get("user"), "users[0].user should be non-empty")
        token = user.get("user", {}).get("token")
        self.assertTrue(token, "users[0].user.token should be non-empty")
        if expected_user_name:
            self.assertEqual(user.get("name"), expected_user_name)

        # Clusters
        self.assertTrue(cfg.get("clusters"), "clusters should be non-empty")
        cluster_entry = cfg["clusters"][0]
        self.assertTrue(cluster_entry.get("name"), "clusters[0].name should be non-empty")
        self.assertTrue(cluster_entry.get("cluster"), "clusters[0].cluster should be non-empty")
        cluster = cluster_entry["cluster"]
        self.assertTrue(cluster.get("server"), "clusters[0].cluster.server should be non-empty")
        self.assertIn("insecure-skip-tls-verify", cluster, "cluster should have insecure-skip-tls-verify")
        if expected_server:
            self.assertEqual(cluster.get("server"), expected_server)

        # Contexts
        self.assertTrue(cfg.get("contexts"), "contexts should be non-empty")
        ctx_entry = cfg["contexts"][0]
        self.assertTrue(ctx_entry.get("name"), "contexts[0].name should be non-empty")
        self.assertTrue(ctx_entry.get("context"), "contexts[0].context should be non-empty")
        ctx = ctx_entry["context"]
        self.assertTrue(ctx.get("cluster"), "contexts[0].context.cluster should be non-empty")
        self.assertTrue(ctx.get("user"), "contexts[0].context.user should be non-empty")
        self.assertTrue(ctx.get("namespace"), "contexts[0].context.namespace should be non-empty")
        if expected_cluster_name:
            self.assertEqual(cfg.get("current-context"), expected_cluster_name)
            self.assertEqual(ctx_entry.get("name"), expected_cluster_name)
        if expected_namespace:
            self.assertEqual(ctx.get("namespace"), expected_namespace)

    def _sa_exists(self, sa, ns):
        out, err = _run_command("kubectl get sa " + sa + " -n " + ns + self.kubeconfig_flag)
        return out and sa in out and "NotFound" not in (err or "")

    def test_provider_kubeconfig_all_fields_nonempty(self):
        """Provider kubeconfig: every field that should exist is non-empty."""
        ns = "kubeplus-test-prov-" + str(os.getpid())
        try:
            cfg, proc = self._create_and_get_kubeconfig(ROOT, ns)
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(
                cfg,
                expected_namespace=ns,
                expected_user_name="kubeplus-saas-provider",
            )
            self.assertTrue(self._sa_exists("kubeplus-saas-provider", ns))
        finally:
            self._delete_for_cleanup(ROOT, ns)

    def test_consumer_kubeconfig_all_fields_nonempty(self):
        """Consumer kubeconfig: every field non-empty, user name matches SA, namespace in context."""
        ns = "kubeplus-test-cons-" + str(os.getpid())
        consumer_sa = "test-consumer-sa"
        try:
            cfg, proc = self._create_and_get_kubeconfig(ROOT, ns, sa=consumer_sa)
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(
                cfg,
                expected_namespace=ns,
                expected_user_name=consumer_sa,
            )
            self.assertTrue(self._sa_exists(consumer_sa, ns))
        finally:
            self._delete_for_cleanup(ROOT, ns, sa=consumer_sa)

    def test_flag_s_apiserverurl_reflected_in_kubeconfig(self):
        """-s/--apiserverurl sets cluster server in kubeconfig."""
        ns = "kubeplus-test-s-" + str(os.getpid())
        test_server = "https://api.example.com:6443"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                ROOT, ns,
                extra_args=["-s", test_server],
            )
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(cfg, expected_server=test_server, expected_namespace=ns)
        finally:
            self._delete_for_cleanup(ROOT, ns)

    def test_flag_x_clustername_reflected_in_kubeconfig(self):
        """-x/--clustername sets context name and cluster name in kubeconfig."""
        ns = "kubeplus-test-x-" + str(os.getpid())
        test_cluster = "my-test-cluster"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                ROOT, ns,
                extra_args=["-x", test_cluster],
            )
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(
                cfg,
                expected_cluster_name=test_cluster,
                expected_namespace=ns,
            )
        finally:
            self._delete_for_cleanup(ROOT, ns)

    def test_flag_f_filename_uses_custom_output_file(self):
        """-f/--filename writes kubeconfig to specified file."""
        ns = "kubeplus-test-f-" + str(os.getpid())
        custom_name = "custom-provider-kubeconfig"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                ROOT, ns,
                extra_args=["-f", custom_name],
                output_filename=custom_name,
            )
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(cfg, expected_namespace=ns)
            self.assertTrue(os.path.exists(os.path.join(ROOT, custom_name + ".json")))
        finally:
            self._delete_for_cleanup(ROOT, ns, filename=custom_name)

    def test_flags_s_and_x_combined(self):
        """-s and -x together: both server and cluster name in kubeconfig."""
        ns = "kubeplus-test-sx-" + str(os.getpid())
        test_server = "https://api.example.com:6443"
        test_cluster = "my-test-cluster"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                ROOT, ns,
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
            self._delete_for_cleanup(ROOT, ns)

    def test_consumer_cannot_create_pod_in_other_namespace(self):
        """
        Consumer kubeconfig: verify create/delete in other namespaces is forbidden.
        Consumer RBAC should restrict operations; creating a pod in another ns should fail.
        """
        ns = "kubeplus-test-restrict-" + str(os.getpid())
        other_ns = "kubeplus-test-other-" + str(os.getpid())
        consumer_sa = "test-consumer-restrict"
        kubeconfig_path = os.path.join(ROOT, consumer_sa + ".json")
        try:
            cfg, proc = self._create_and_get_kubeconfig(ROOT, ns, sa=consumer_sa)
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(cfg, expected_namespace=ns, expected_user_name=consumer_sa)

            # Create another namespace (as admin)
            _run_command("kubectl create namespace " + other_ns + self.kubeconfig_flag)

            # Try to create pod in other namespace using consumer kubeconfig (should fail)
            out, err = _run_command(
                "kubectl run nginx --image=nginx -n " + other_ns
                + " --kubeconfig=" + kubeconfig_path
            )
            # Expect Forbidden (authorization denial), not generic errors (DNS, image pull, etc.)
            self.assertTrue(
                "forbidden" in err.lower(),
                "Consumer should not be able to create pod in other namespace; got out=%r err=%r"
                % (out, err),
            )
        finally:
            _run_command("kubectl delete namespace " + other_ns + self.kubeconfig_flag + " 2>/dev/null")
            self._delete_for_cleanup(ROOT, ns, sa=consumer_sa)


if __name__ == "__main__":
    unittest.main()
