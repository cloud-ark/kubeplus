# Provider Kubeconfig Refactor

## Summary

Refactors `provider-kubeconfig.py` to use unified RBAC rule lists, fixes several bugs present on master, and adds tests. The script generates provider and consumer kubeconfigs with appropriate RBAC for the KubePlus platform.

---

## Bug Fixes (master had these bugs)

### nonResourceURL parsing (`_update_rbac`)

Master used `parts[0]` after splitting on `nonResourceURL::`; the correct value is in `parts[1]`. For `"foo/nonResourceURL::bar"`: master→`"foo/"`, PR→`"bar"`.

### Namespace creation (create action)

Master had `create_ns = "kubectl get ns " + ...` and ran that same GET command twice when the namespace was not found. The namespace was never created. PR runs `kubectl create ns <namespace>` when the namespace does not exist.

### Token extraction (`_extract_kubeconfig`)

Master used `if 'token' in line`, which matches any line containing "token" (e.g. `Type: kubernetes.io/service-account-token`, `Labels: kubernetes.io/legacy-token-last-used`). PR only treats lines where the key is exactly `"token"` as token lines.

### Sleep placement (token extraction)

Master had `time.sleep(2)` inside the line-parsing loop, so it ran once per line when the token wasn't found. PR moves it outside the loop so it runs once per retry.

### Default kubeconfig path

Master used `os.getenv("HOME") + "/.kube/config"`, which can fail if `HOME` is unset. PR uses `os.path.expanduser("~")` for robustness.

---

## Main Changes (master → PR)

### RBAC refactor

- Added a temporary shadow-compare path:
  - keep legacy grouped rules ("pink") and unified `rule_list` rules ("green") side by side
  - optional parity assertion via `KUBEPLUS_RBAC_EQ_CHECK=1`
  - runtime still uses legacy grouped rules in this PR to minimize risk
- Same permissions target: consumer (read + apps + impersonate + portforward), provider (full platform operator).
- `-perms` ConfigMap behavior:
  - consumer `all_resources` includes wildcard entries as before
  - provider `all_resources` excludes `"*"` to match master behavior (master effectively omitted wildcard groups from provider perms inventory)

### Code cleanup

- Use module-level `run_command` instead of `self.run_command`.
- `create_role_rolebinding`: use `with open(..., encoding="utf-8")`, `os.path.join`.
- `run_command`: use context manager, handle `None` from `communicate()`.
- `sorted(list(set(x)))` → `sorted(set(x))`.

---

## Tests (new; master has none)

New file `tests/test_provider_kubeconfig.py`:

- **CLI:** `--help` shows actions, flags, namespace; `update` without `-p` exits with error.
- **Integration:** Provider and consumer kubeconfigs have non-empty fields, correct namespace, SA creation; flags `-s`, `-x`, `-f` reflected in output.
- **Consumer RBAC:** `test_consumer_cannot_create_pod_in_other_namespace` verifies create in another namespace is forbidden.

Integration tests skip when no cluster is reachable (`KUBECONFIG` unset).

---

## How to test

```bash
# CLI tests (no cluster needed)
python -m unittest tests.test_provider_kubeconfig.TestCli -v

# Full suite (requires cluster)
KUBECONFIG=/path/to/kubeconfig python -m unittest tests.test_provider_kubeconfig -v
```
