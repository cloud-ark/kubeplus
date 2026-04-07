"""
Tests for provider-kubeconfig.py: CLI, structure validation, integration tests
that verify non-empty fields, service account creation, and CLI flags in output.
"""
import json
import importlib.util
import os
import subprocess
import sys
import tempfile
import unittest
import uuid


def _run_command(cmd):
    """Run shell command, return (stdout, stderr)."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout or "", result.stderr or ""


def _cluster_available(kubeconfig=""):
    """True if kubectl can reach a cluster."""
    path = os.path.expanduser(kubeconfig).strip() if kubeconfig else ""
    k_opt = " --kubeconfig=" + path if path else ""
    out, err = _run_command("kubectl get ns default" + k_opt)
    return out and "default" in out and "NotFound" not in err


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
        for elem in HELP_ELEMENTS:
            self.assertIn(elem, out, f"Help should mention {elem}")

    def test_update_without_permissionfile_exits_with_error(self):
        """update action without -p exits with code 1."""
        proc = subprocess.run(
            [sys.executable, os.path.join(ROOT, SCRIPT), "update", "default"],
            capture_output=True, text=True, cwd=ROOT,
        )
        self.assertNotEqual(proc.returncode, 0)
        self.assertIn("permission", (proc.stdout or proc.stderr or "").lower())

    def test_revoke_without_permissionfile_exits_with_error(self):
        """revoke action without -p exits with code 1."""
        proc = subprocess.run(
            [sys.executable, os.path.join(ROOT, SCRIPT), "revoke", "default"],
            capture_output=True, text=True, cwd=ROOT,
        )
        self.assertNotEqual(proc.returncode, 0)
        self.assertIn("permission", (proc.stdout or proc.stderr or "").lower())


class TestPermissionFileParsing(unittest.TestCase):
    """Unit tests for permission file parsing (JSON/YAML)."""

    @classmethod
    def setUpClass(cls):
        script_path = os.path.join(ROOT, SCRIPT)
        spec = importlib.util.spec_from_file_location("provider_kubeconfig_module", script_path)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        cls.generator = module.KubeconfigGenerator()

    def _write_temp_file(self, suffix, content):
        fd, path = tempfile.mkstemp(suffix=suffix, dir=ROOT, text=True)
        os.close(fd)
        with open(path, "w", encoding="utf-8") as fp:
            fp.write(content)
        return path

    def test_load_permission_data_accepts_json(self):
        json_content = json.dumps(
            {
                "perms": {
                    "apps": [{"deployments": ["get", "create"]}],
                    "non-apigroup": [{"nonResourceURL::/metrics": ["get"]}],
                }
            }
        )
        path = self._write_temp_file(".json", json_content)
        try:
            perms = self.generator._load_permission_data(path)
            rules, resources = self.generator._parse_permission_rules(perms)
            self.assertIn("apps", perms)
            self.assertIn("deployments", resources)
            self.assertTrue(any("nonResourceURLs" in r for r in rules))
        finally:
            os.remove(path)

    def test_load_permission_data_accepts_yaml(self):
        yaml_content = """
perms:
  apps:
    - deployments:
      - get
      - update
  non-apigroup:
    - "nonResourceURL::/healthz":
      - get
"""
        path = self._write_temp_file(".yaml", yaml_content)
        try:
            perms = self.generator._load_permission_data(path)
            rules, resources = self.generator._parse_permission_rules(perms)
            self.assertIn("apps", perms)
            self.assertIn("deployments", resources)
            self.assertTrue(any("/healthz" in str(r.get("nonResourceURLs", [])) for r in rules))
        finally:
            os.remove(path)

    def test_parse_permission_rules_non_apigroup_without_marker_appends_empty_rule(self):
        """non-apigroup entries without nonResourceURL still append a rule (legacy behavior)."""
        perms = {"non-apigroup": [{"not-a-url": ["get"]}]}
        rules, resources = self.generator._parse_permission_rules(perms)
        self.assertEqual(len(rules), 1)
        self.assertEqual(rules[0], {})
        self.assertEqual(resources, ["not-a-url"])

    def test_permission_fixture_files_parse_json_and_yaml(self):
        """Fixture files in tests/permission_files should load and parse via script helpers."""
        fixture_dir = os.path.join(ROOT, "tests", "permission_files")
        fixture_files = [
            "permissions-example1.json",
            "permissions-example1.yaml",
            "permissions-example2.json",
            "permissions-example2.yaml",
            "permissions-example3.json",
            "permissions-example3.yaml",
        ]
        for name in fixture_files:
            path = os.path.join(fixture_dir, name)
            self.assertTrue(os.path.exists(path), f"Missing fixture file: {name}")
            perms = self.generator._load_permission_data(path)
            rules, resources = self.generator._parse_permission_rules(perms)
            self.assertIsInstance(perms, dict, f"{name}: perms should be dict")
            self.assertGreater(len(rules), 0, f"{name}: expected at least one parsed rule")
            self.assertGreater(len(resources), 0, f"{name}: expected at least one parsed resource")


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

    def _create_and_get_kubeconfig(self, ns, sa="kubeplus-saas-provider", extra_args=None, output_filename=None):
        """Run create, return (kubeconfig_dict, proc). Caller must delete to cleanup."""
        create_args = ["create", ns]
        if sa != "kubeplus-saas-provider":
            create_args += ["-c", sa]
        if extra_args:
            create_args += extra_args
        proc = subprocess.run(
            [sys.executable, os.path.join(ROOT, SCRIPT)] + create_args + self.kubeconfig_arg,
            cwd=ROOT, capture_output=True, text=True, timeout=120,
        )
        if proc.returncode != 0:
            return None, proc
        filename = output_filename or (sa + ".json")
        if not filename.endswith(".json"):
            filename += ".json"
        kubeconfig_path = os.path.join(ROOT, filename)
        if not os.path.exists(kubeconfig_path):
            raise AssertionError(
                f"Script exited 0 but kubeconfig not written: {kubeconfig_path}"
            )
        with open(kubeconfig_path, "r", encoding="utf-8") as fp:
            cfg = json.load(fp)
        return cfg, proc

    def _delete_for_cleanup(self, ns, sa="kubeplus-saas-provider", filename=None):
        """Delete k8s resources, local files, and test namespace."""
        delete_args = ["delete", ns]
        if sa != "kubeplus-saas-provider":
            delete_args += ["-c", sa]
        if filename:
            delete_args += ["-f", filename]
        subprocess.run(
            [sys.executable, os.path.join(ROOT, SCRIPT)] + delete_args + self.kubeconfig_arg,
            cwd=ROOT, capture_output=True, timeout=60,
        )
        _run_command(
            "kubectl delete namespace " + ns
            + " --ignore-not-found --wait=false" + self.kubeconfig_flag
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
            self.assertEqual(cluster_entry.get("name"), expected_cluster_name)
            self.assertEqual(ctx.get("cluster"), expected_cluster_name)
        if expected_namespace:
            self.assertEqual(ctx.get("namespace"), expected_namespace)

    def _sa_exists(self, sa, ns):
        out, err = _run_command("kubectl get sa " + sa + " -n " + ns + self.kubeconfig_flag)
        return out and sa in out and "NotFound" not in err

    def _current_cluster_server(self):
        """Return current cluster server URL from kubeconfig, if available."""
        cmd = "kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'"
        if self.kubeconfig_flag:
            cmd += self.kubeconfig_flag
        out, _ = _run_command(cmd)
        return (out or "").strip().strip("'")

    def _auth_can_i(self, namespace, sa, verb, resource):
        """Run auth can-i for an impersonated ServiceAccount and return normalized output."""
        out, err = _run_command(
            "kubectl auth can-i " + verb + " " + resource
            + " -n " + namespace
            + " --as=system:serviceaccount:" + namespace + ":" + sa
            + self.kubeconfig_flag
        )
        if err and ("unable to connect to the server" in err.lower() or "i/o timeout" in err.lower()):
            self.skipTest("Skipping authz assertion due to transient API connectivity issue: " + err.strip())
        return (out or "").strip().lower()

    def test_provider_kubeconfig_all_fields_nonempty(self):
        """Provider kubeconfig: every field that should exist is non-empty."""
        ns = "kubeplus-test-prov-" + uuid.uuid4().hex[:8]
        try:
            cfg, proc = self._create_and_get_kubeconfig(ns)
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(
                cfg,
                expected_namespace=ns,
                expected_user_name="kubeplus-saas-provider",
            )
            self.assertTrue(self._sa_exists("kubeplus-saas-provider", ns))
        finally:
            self._delete_for_cleanup(ns)

    def test_consumer_kubeconfig_all_fields_nonempty(self):
        """Consumer kubeconfig: every field non-empty, user name matches SA, namespace in context."""
        ns = "kubeplus-test-cons-" + uuid.uuid4().hex[:8]
        consumer_sa = "test-consumer-sa"
        try:
            cfg, proc = self._create_and_get_kubeconfig(ns, sa=consumer_sa)
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(
                cfg,
                expected_namespace=ns,
                expected_user_name=consumer_sa,
            )
            self.assertTrue(self._sa_exists(consumer_sa, ns))
        finally:
            self._delete_for_cleanup(ns, sa=consumer_sa)

    def test_flag_s_apiserverurl_reflected_in_kubeconfig(self):
        """-s/--apiserverurl sets cluster server in kubeconfig."""
        ns = "kubeplus-test-s-" + uuid.uuid4().hex[:8]
        test_server = "https://api.example.com:6443"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                ns,
                extra_args=["-s", test_server],
            )
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(cfg, expected_server=test_server, expected_namespace=ns)
        finally:
            self._delete_for_cleanup(ns)

    def test_flag_x_clustername_reflected_in_kubeconfig(self):
        """-x/--clustername sets context name and cluster name in kubeconfig."""
        ns = "kubeplus-test-x-" + uuid.uuid4().hex[:8]
        test_cluster = "my-test-cluster"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                ns,
                extra_args=["-x", test_cluster],
            )
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(
                cfg,
                expected_cluster_name=test_cluster,
                expected_namespace=ns,
            )
        finally:
            self._delete_for_cleanup(ns)

    def test_flag_f_filename_uses_custom_output_file(self):
        """-f/--filename writes kubeconfig to specified file."""
        ns = "kubeplus-test-f-" + uuid.uuid4().hex[:8]
        custom_name = "custom-provider-kubeconfig"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                ns,
                extra_args=["-f", custom_name],
                output_filename=custom_name,
            )
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(cfg, expected_namespace=ns)
            self.assertTrue(os.path.exists(os.path.join(ROOT, custom_name + ".json")))
        finally:
            self._delete_for_cleanup(ns, filename=custom_name)

    def test_flags_s_and_x_combined(self):
        """-s and -x together: both server and cluster name in kubeconfig."""
        ns = "kubeplus-test-sx-" + uuid.uuid4().hex[:8]
        test_server = "https://api.example.com:6443"
        test_cluster = "my-test-cluster"
        try:
            cfg, proc = self._create_and_get_kubeconfig(
                ns,
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
            self._delete_for_cleanup(ns)

    def test_consumer_can_create_deployment_but_not_pod(self):
        """
        Consumer kubeconfig: verify allowed and disallowed creates:
        - create deployment in own namespace succeeds
        - create pod in own namespace is forbidden
        """
        ns = "kubeplus-test-restrict-" + uuid.uuid4().hex[:8]
        consumer_sa = "test-consumer-restrict"
        kubeconfig_path = os.path.join(ROOT, consumer_sa + ".json")
        api_server = self._current_cluster_server()
        try:
            extra_args = ["-s", api_server] if api_server else None
            cfg, proc = self._create_and_get_kubeconfig(ns, sa=consumer_sa, extra_args=extra_args)
            self.assertEqual(proc.returncode, 0, proc.stderr)
            self._assert_kubeconfig_valid(cfg, expected_namespace=ns, expected_user_name=consumer_sa)

            # Verify create works in own namespace.
            own_name = "consumer-own-" + uuid.uuid4().hex[:6]
            own_out, own_err = _run_command(
                "kubectl create deployment " + own_name + " --image=nginx -n " + ns
                + " --kubeconfig=" + kubeconfig_path
            )
            own_conn_err = "unable to connect to the server" in own_err.lower() or "i/o timeout" in own_err.lower()
            if own_conn_err:
                self.skipTest("Skipping authz assertion due to transient API connectivity issue: " + own_err.strip())
            self.assertTrue(
                "created" in own_out.lower(),
                "Consumer should be able to create deployment in own namespace; got out=%r err=%r"
                % (own_out, own_err),
            )

            # Verify pod create is forbidden for consumer.
            pod_name = "consumer-pod-" + uuid.uuid4().hex[:6]
            out, err = _run_command(
                "kubectl run " + pod_name + " --image=nginx -n " + ns
                + " --kubeconfig=" + kubeconfig_path
            )
            conn_err = "unable to connect to the server" in err.lower() or "i/o timeout" in err.lower()
            if conn_err:
                self.skipTest("Skipping authz assertion due to transient API connectivity issue: " + err.strip())
            self.assertTrue(
                "forbidden" in err.lower(),
                "Consumer should not be able to create pods; got out=%r err=%r"
                % (out, err),
            )
        finally:
            _run_command("kubectl delete deployment --all -n " + ns + self.kubeconfig_flag + " 2>/dev/null")
            self._delete_for_cleanup(ns, sa=consumer_sa)

    def test_update_revoke_json_yaml_have_same_auth_can_i_effect(self):
        """
        update/revoke with equivalent JSON and YAML permission files should yield same auth outcome:
        baseline deny -> allow after update -> deny after revoke.
        """
        ns = "kubeplus-test-uprev-" + uuid.uuid4().hex[:8]
        sa = "uprev-sa"
        json_file = os.path.join("tests", "permission_files", "permissions-example1.json")
        yaml_file = os.path.join("tests", "permission_files", "permissions-example1.yaml")
        self.assertTrue(os.path.exists(os.path.join(ROOT, json_file)))
        self.assertTrue(os.path.exists(os.path.join(ROOT, yaml_file)))
        try:
            _run_command("kubectl create ns " + ns + self.kubeconfig_flag)
            _run_command("kubectl create sa " + sa + " -n " + ns + self.kubeconfig_flag)

            def run_flow(permission_file):
                baseline = self._auth_can_i(ns, sa, "create", "secrets")
                self.assertIn(baseline, ["yes", "no"])
                proc_update = subprocess.run(
                    [
                        sys.executable,
                        os.path.join(ROOT, SCRIPT),
                        "update",
                        ns,
                        "-c",
                        sa,
                        "-p",
                        permission_file,
                    ] + self.kubeconfig_arg,
                    cwd=ROOT,
                    capture_output=True,
                    text=True,
                    timeout=120,
                )
                self.assertEqual(proc_update.returncode, 0, proc_update.stderr)
                after_update = self._auth_can_i(ns, sa, "create", "secrets")
                proc_revoke = subprocess.run(
                    [
                        sys.executable,
                        os.path.join(ROOT, SCRIPT),
                        "revoke",
                        ns,
                        "-c",
                        sa,
                        "-p",
                        permission_file,
                    ] + self.kubeconfig_arg,
                    cwd=ROOT,
                    capture_output=True,
                    text=True,
                    timeout=120,
                )
                self.assertEqual(proc_revoke.returncode, 0, proc_revoke.stderr)
                after_revoke = self._auth_can_i(ns, sa, "create", "secrets")
                return baseline, after_update, after_revoke

            json_result = run_flow(json_file)
            yaml_result = run_flow(yaml_file)
            self.assertEqual(json_result, yaml_result)
            self.assertEqual(json_result[1], "yes")
            self.assertEqual(json_result[2], "no")
        finally:
            _run_command("kubectl delete clusterrole " + sa + "-update" + self.kubeconfig_flag + " 2>/dev/null")
            _run_command("kubectl delete clusterrolebinding " + sa + "-update" + self.kubeconfig_flag + " 2>/dev/null")
            _run_command("kubectl delete configmap " + sa + "-perms -n " + ns + self.kubeconfig_flag + " 2>/dev/null")
            _run_command("kubectl delete sa " + sa + " -n " + ns + self.kubeconfig_flag + " 2>/dev/null")
            _run_command("kubectl delete namespace " + ns + " --ignore-not-found --wait=false" + self.kubeconfig_flag)


if __name__ == "__main__":
    unittest.main()
