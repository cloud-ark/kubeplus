```
::KubePlus kubectl commands::
---

1. kubectl connections - Get relationships of a Kubernetes resource with other resources.

kubectl connections <Kind> <Instance> <Namespace> [--kubeconfig=<Absolute path to kubeconfig>] [-o json|png] (default value='flat') [--ignore=<Kind1:Instance1,Kind1:Instance1>]

---

2. kubectl man <Custom Resource Kind> - Get Man page information about a Custom Resource

---

3. kubectl metrics cr - Get CPU/Memory/Storage consumption of a Custom Resource instance

4. kubectl metrics service - Get CPU/Memory/Storage consumption of all the Pods reachable from a Service instance

5. kubectl metrics account - Get CPU/Memory/Storage consumption of all the Pods that are created by the given account

6. kubectl metrics helmrelease - Get CPU/Memory/Storyage consumption of all the Pods that are part of a Helm release

---

7. kubectl grouplogs cr composition - Get logs of all the Pod/containers that are children of a Custom Resource instance

8. kubectl grouplogs cr connections - Get logs of all the Pod/containers that are related to a Custom Resource instance

9. kubectl grouplogs service - Get logs of all the Pods/containers that are related to a Service instance

10. kubectl grouplogs helmrelease - Get logs of all the Pods/containers that are part of a Helm release

---

```