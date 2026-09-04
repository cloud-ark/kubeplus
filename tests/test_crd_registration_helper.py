import importlib.util
import os
import select
import unittest
import uuid
from unittest.mock import patch

from kubernetes import config
from kubernetes.client import Configuration
from kubernetes.client.api import core_v1_api
from kubernetes.stream import portforward


ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
HELPER = os.path.join(ROOT, "deploy", "crd_registration_helper.py")


def _load_helper():
    spec = importlib.util.spec_from_file_location(
        "crd_registration_helper", HELPER
    )
    module = importlib.util.module_from_spec(spec)
    with patch("logging.config.dictConfig"):
        spec.loader.exec_module(module)

    return module


crd_registration_helper = _load_helper()


class TestCRDRegistrationHelper(unittest.TestCase):

    @patch.object(crd_registration_helper, "run_command")
    def test_download_and_untar_chart_uses_argument_lists(self, mock_run_command):
        """Commands in download_and_untar_chart are passed as argument lists."""
        malicious_chart_name = "$(touch /tmp/kubeplus-cve-test)"
        mock_run_command.return_value = ("", "")
        crd_registration_helper.download_and_untar_chart("https://example.com/chart.tgz", malicious_chart_name)
        commands = [item.args[0] for item in mock_run_command.call_args_list]

        self.assertEqual(commands[0], ["wget", "-O", "/" + malicious_chart_name + ".tgz", "--no-check-certificate", "https://example.com/chart.tgz",],)
        self.assertEqual(commands[1], ["rm", "-rf", "/" + malicious_chart_name,],)
        self.assertEqual(commands[2], ["tar", "-xvzf", "/" + malicious_chart_name + ".tgz",],)

        for command in commands:
            self.assertIsInstance(command, list)

    def test_registercrd_chart_name_command_injection(self):
        """A malicious chartName must not execute a shell command."""
        marker = "/tmp/kubeplus-cve-" + uuid.uuid4().hex
        config.load_kube_config()
        configuration = Configuration.get_default_copy()
        configuration.assert_hostname = False
        Configuration.set_default(configuration)
        api_instance = core_v1_api.CoreV1Api()

        pods = api_instance.list_namespaced_pod(namespace="default", label_selector="app=kubeplus",)
        kubeplus_pod = None
        for pod in pods.items:
            if pod.status.phase == "Running":
                kubeplus_pod = pod.metadata.name
                break
        self.assertIsNotNone(kubeplus_pod, "A running KubePlus pod is required for this integration test",)

        pf = portforward(api_instance.connect_get_namespaced_pod_portforward, kubeplus_pod, "default", ports="5005",)
        http = pf.socket(5005)
        http.setblocking(True)

        try:
            request = (
                "GET /registercrd"
                "?kind=InjectionTest"
                "&version=v1"
                "&group=inject.test.io"
                "&plural=injectiontests"
                "&chartURL=file:///tmp/hello-world-chart-0.0.3.tgz"
                "&chartName=$(touch%20" + marker + ")"
                " HTTP/1.1\r\n"
                "Host: 127.0.0.1\r\n"
                "Connection: close\r\n"
                "Accept: */*\r\n"
                "\r\n"
            )
            http.sendall(request.encode("utf-8"))
            response = b""
            while True:
                select.select([http], [], [])
                data = http.recv(1024)
                if not data:
                    break
                response += data
        finally:
            http.close()

        self.assertFalse(os.path.exists(marker), "Security check failed: chartName was executed as a shell command",)
        if os.path.exists(marker):
            os.remove(marker)


if __name__ == "__main__":
    unittest.main()