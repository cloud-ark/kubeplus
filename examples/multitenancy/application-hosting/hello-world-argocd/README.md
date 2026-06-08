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

Argo CD should sync from a small examples repo rather than from the full
`kubeplus` source repo. In this layout, both the provider and consumer Argo CD
applications point at this same `kubeplus-examples` repo, but use different
subpaths:

- `multitenancy/application-hosting/hello-world-argocd/provider`
- `multitenancy/application-hosting/hello-world-argocd/tenants`

This keeps the GitOps source compact while still allowing the full example to
live in one official examples repository.
