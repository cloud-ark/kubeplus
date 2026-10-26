import json
from pathlib import Path
import runpy
import sys
import unittest
from unittest.mock import patch


main = runpy.run_path(str(Path(__file__).resolve().parents[1] / "plugins/kubectl-resource_policy"))["main"]


class TestResourcePolicyCLI(unittest.TestCase):
    @patch("subprocess.run")
    def test_allow_uses_the_same_policy_for_reversed_pairs(self, run):
        run.return_value.returncode = 0
        for pair in [("chatbot-a", "rag-a"), ("rag-a", "chatbot-a")]:
            with patch.object(sys, "argv", ["resource-policy", "allow-traffic", *pair,
                                           "-n", "policies", "-k", "test.kubeconfig"]):
                self.assertEqual(main(), 0)
            self.assertEqual(run.call_args.args[0], [
                "kubectl", "--namespace", "policies", "--kubeconfig", "test.kubeconfig", "apply", "-f", "-"
            ])
            self.assertEqual(json.loads(run.call_args.kwargs["input"]), {
                "apiVersion": "workflows.kubeplus/v1alpha1",
                "kind": "ResourcePolicy",
                "metadata": {"name": "network-access.chatbot-a.rag-a"},
                "spec": {"policy": {"network": {"allowedPairs": [
                    {"namespace1": "chatbot-a", "namespace2": "rag-a"}
                ]}}},
            })

    @patch("subprocess.run")
    def test_deny_removes_only_the_pair_policy_and_propagates_errors(self, run):
        run.return_value.returncode = 1
        with patch.object(sys, "argv", ["resource-policy", "deny-traffic", "rag-a", "chatbot-a"]):
            self.assertEqual(main(), 1)
        self.assertEqual(run.call_args.args[0], [
            "kubectl", "delete", "resourcepolicy", "network-access.chatbot-a.rag-a", "--ignore-not-found"
        ])


if __name__ == "__main__":
    unittest.main()
