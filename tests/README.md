## Running tests

```python3 -m unittest -v tests```

For the standalone ResourcePolicy network whitelist, see the
[manual demo](test_resource_policy/TESTING.md).

## Run tests with detailed output:

```
   export KUBEPLUS_TEST_OUTPUT=yes
   python3 -m unittest -v tests
```

## Run individual test:

```python3 -m unittest -v tests.TestKubePlus.<test_name>```





