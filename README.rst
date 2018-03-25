=========
KubePlus
=========

Purpose-built application platforms on Kubernetes.

KubePlus Purpose-built Platforms extend Kubernetes with custom resources to embed customer-specific platform workflows directly in Kubernetes.  
Example of such extensions can be Postgres, MySQL, Fluentd, Prometheus etc. 


**Guidelines for consistency across all custom resources**

Based on our study of existing Kubernetes custom controllers/extensions, we have developed guidelines that need to be followed by
any custom controller to be part of KubePlus. This brings consistency and quality even when we use any other existing open-source controller for your needs.


**Improved usability of custom resources**

KubePlus installs an additional software component KubeArk on to your Kubernetes cluster that assists users of KubePlus 
to easily consume new custom resources directly from your Kubernetes cluster. KubeArk serves following purposes: 

- Provides information about resource specific configurable parameters exposed by the controller (e.g. MySQL configurable parameters)

- Provides information on workflow actions that can be performed during the lifecycle of custom resource

- Provides information of composition of custom resources in terms on native Kubernetes resources (e.g. pods, services etc.)
