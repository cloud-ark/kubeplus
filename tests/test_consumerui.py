import importlib.util
import os
import unittest
from unittest.mock import patch


ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
CONSUMERUI = os.path.join(ROOT, "consumerui", "consumerui.py")


def _load_consumerui():
    spec = importlib.util.spec_from_file_location("consumerui_app", CONSUMERUI)
    module = importlib.util.module_from_spec(spec)
    with patch("logging.config.dictConfig"):
        spec.loader.exec_module(module)
    return module


consumerui = _load_consumerui()


class TestConsumerUI(unittest.TestCase):

    @patch.object(consumerui.subprocess, "Popen")
    def test_resource_manpage_passes_resource_as_single_argument(self, mock_popen):
        malicious_resource = "foo;touch /tmp/kubeplus-cve-test"
        mock_popen.return_value.communicate.return_value = (b"", b"")

        with patch.dict(os.environ, {"HOME": "/root"}):
            response = consumerui.app.test_client().get(
                "/get_resource_manpage", query_string={"resource": malicious_resource}
            )

        self.assertEqual(response.status_code, 200)
        command = mock_popen.call_args.args[0]
        self.assertIsInstance(command, list)
        self.assertEqual(
            command,
            ["kubectl", "man", malicious_resource, "-k", "/root/.kube/config"],
        )
        self.assertFalse(mock_popen.call_args.kwargs.get("shell", False))


if __name__ == "__main__":
    unittest.main()
