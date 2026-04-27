Hello World Argo CD
===================

This example captures provider-side and consumer-side GitOps inputs for the
KubePlus Hello World multi-tenancy flow.

What this example includes
--------------------------

- The Argo CD `AppProject` used for the provider sync
- A provider-side `ResourceComposition` mirror under `provider/`
- The Argo CD `Application` used for the provider sync
- The Argo CD `AppProject` used for the consumer sync
- The Argo CD `Application` used for the consumer sync
- Tenant instances of `HelloWorldService`
- A short `steps.txt` for reproducing the workflow

Important note
--------------

Argo CD should still sync from dedicated small repos:

- one provider repo for the `ResourceComposition`
- one consumer repo for the tenant instances

The files in this folder are mirrors/templates so the example can be committed
and documented inside `kubeplus`.
