import os
import sys
import unittest
from unittest.mock import Mock, patch


ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
PLUGINS = os.path.join(ROOT, "plugins")
sys.path.insert(0, PLUGINS)
try:
    import appresources
    import crmetrics
finally:
    sys.path.remove(PLUGINS)


class TestPluginSubprocessCommands(unittest.TestCase):

    @patch.object(crmetrics.subprocess, "Popen")
    def test_resource_value_is_one_argument_and_shell_is_not_enabled(self, mock_popen):
        malicious_resource = "foo;touch /tmp/kubeplus-plugin-test"
        malicious_instance = "bar;touch /tmp/kubeplus-instance-test"
        mock_popen.return_value.communicate.return_value = (b'{"kind":"Foo"}', b"")

        crmetrics.CRMetrics()._get_identity(malicious_resource, malicious_instance, "default")

        command = mock_popen.call_args.args[0]
        self.assertEqual(
            command,
            ["kubectl", "get", malicious_resource, malicious_instance, "-n", "default", "-o", "json"],
        )
        self.assertIsInstance(command, list)
        self.assertNotIn("shell", mock_popen.call_args.kwargs)

    @patch.object(crmetrics.subprocess, "Popen")
    def test_kubeplus_namespace_pipeline_is_parsed_without_a_shell(self, mock_popen):
        mock_popen.return_value.communicate.return_value = (
            b"NAMESPACE NAME READY STATUS RESTARTS AGE\n"
            b"platform-system kubeplus-deployment-123 1/1 Running 0 2d\n",
            b"",
        )

        kubeconfig = "--kubeconfig=/tmp/provider kubeconfig;touch /tmp/kubeplus-kubeconfig-test"
        namespace = crmetrics.CRBase().get_kubeplus_namespace(kubeconfig)

        self.assertEqual(namespace, "platform-system")
        self.assertEqual(
            mock_popen.call_args.args[0],
            ["kubectl", "get", "pods", "-A", kubeconfig],
        )
        self.assertNotIn("shell", mock_popen.call_args.kwargs)

    @patch.object(crmetrics.subprocess, "Popen")
    def test_top_metrics_pipeline_ignores_header_and_preserves_pod_arguments(self, mock_popen):
        malicious_pod = "pod;touch-/tmp/kubeplus-pod-test"
        mock_popen.return_value.communicate.return_value = (
            b"NAME CPU(cores) MEMORY(bytes)\n"
            + malicious_pod.encode("utf-8")
            + b" 25m 64Mi\n",
            b"",
        )

        result = crmetrics.CRMetrics()._get_cpu_memory_usage_kubectl_top(
            [{"Name": malicious_pod, "Namespace": "team-a"}]
        )

        self.assertEqual(result, (25, 64))
        self.assertEqual(
            mock_popen.call_args.args[0],
            ["kubectl", "top", "pods", malicious_pod, "-n", "team-a"],
        )
        self.assertNotIn("shell", mock_popen.call_args.kwargs)

    @patch.object(crmetrics.CRMetrics, "_get_cpu_memory_usage_rootres", return_value=(2, 3, 1))
    @patch.object(crmetrics.subprocess, "Popen")
    def test_crd_pipeline_parses_crds_and_namespaced_instances(self, mock_popen, mock_metrics):
        mock_popen.side_effect = [
            Mock(communicate=Mock(return_value=(b"NAME CREATED AT\nwidgets.example.com 1m\n", b""))),
            Mock(communicate=Mock(return_value=(
                b"NAMESPACE NAME AGE\nteam-a widget-one 1m\n", b""
            ))),
        ]

        result = crmetrics.CRMetrics()._get_metrics_cr_instances("account")

        self.assertEqual(result, (2, 3, 1))
        self.assertEqual(
            [call.args[0] for call in mock_popen.call_args_list],
            [
                ["kubectl", "get", "crds"],
                ["kubectl", "get", "widgets.example.com", "--all-namespaces"],
            ],
        )
        self.assertEqual(
            mock_metrics.call_args.args[1],
            ["kubectl", "get", "widgets.example.com", "widget-one", "-n", "team-a", "-o", "json"],
        )
        for call in mock_popen.call_args_list:
            self.assertNotIn("shell", call.kwargs)

    @patch.object(crmetrics.CRMetrics, "_get_pod", side_effect=[
        {"metadata": {"annotations": {"accountidentity": "account-a"}}},
        {"metadata": {"annotations": {"accountidentity": "account-b"}}},
    ])
    @patch.object(crmetrics.subprocess, "Popen")
    def test_account_pods_pipeline_parses_columns_and_filters_account(self, mock_popen, mock_get_pod):
        mock_popen.return_value.communicate.return_value = (
            b"NAMESPACE NAME READY STATUS\n"
            b"team-a  pod-one  1/1  Running\n"
            b"team-b  pod-two  1/1  Running\n",
            b"",
        )

        pods = crmetrics.CRMetrics()._get_pods_for_account("account-a")

        self.assertEqual(pods, [{"Namespace": "team-a", "Name": "pod-one"}])
        self.assertEqual(mock_get_pod.call_args_list[0].args[0], {"Namespace": "team-a", "Name": "pod-one"})
        self.assertEqual(mock_get_pod.call_args_list[1].args[0], {"Namespace": "team-b", "Name": "pod-two"})
        self.assertEqual(
            mock_popen.call_args.args[0],
            ["kubectl", "get", "pods", "--all-namespaces"],
        )
        self.assertNotIn("shell", mock_popen.call_args.kwargs)

    @patch.object(crmetrics.CRMetrics, "_get_cpu_memory_usage_rootres", return_value=(1, 2, 1))
    @patch.object(crmetrics.subprocess, "Popen")
    def test_kind_listing_pipeline_filters_headers_and_splits_columns(self, mock_popen, mock_metrics):
        kind = "pods;touch /tmp/kubeplus-plugin-test"
        mock_popen.return_value.communicate.return_value = (
            b"NAMESPACE NAME READY STATUS RESTARTS AGE\n"
            b"team-a  pod-one  1/1  Running  0  2d\n"
            b"team-b  pod-two  1/1  Running  0  2d\n",
            b"",
        )

        result = crmetrics.CRMetrics()._get_metrics_kind(kind, "account")

        self.assertEqual(result, (2, 4, 2))
        self.assertEqual(
            mock_popen.call_args.args[0],
            ["kubectl", "get", kind, "--all-namespaces"],
        )
        self.assertEqual(
            [call.args[2:4] for call in mock_metrics.call_args_list],
            [("pod-one", "team-a"), ("pod-two", "team-b")],
        )
        self.assertNotIn("shell", mock_popen.call_args.kwargs)


if __name__ == "__main__":
    unittest.main()
