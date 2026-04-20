Hello World Argo CD Consumer
============================

This example captures the consumer-side GitOps input for the KubePlus Hello
World multi-tenancy flow.

What this example includes
--------------------------

- The Argo CD `Application` used for the consumer sync
- Tenant instances of `HelloWorldService`
- A short `steps.txt` for reproducing the workflow

Important note
--------------

Argo CD should still sync from the dedicated small repo:

`https://github.com/PatrickBladeeman/hello-world-argocd-consumer`

The files in this folder mirror that repo so the example can be committed and
documented inside `kubeplus`.
