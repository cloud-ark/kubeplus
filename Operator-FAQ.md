# Kubernetes Operator FAQ

**Q. What is a Kubernetes Operator?**

A. An [Operator](https://coreos.com/operators/) is essentially a new REST API that is added to a Kubernetes cluster's control plane. Like traditional REST APIs, it has a resource definition and code that knows how to perform CRUD operations on that resource. In Kubernetes nomenclature, the resource definition is called as the Custom Resource and the code that performs CRUD operations is called Custom Controller. The official definition of an Operator is being debated and discussed
in the [CNCF sig-app-delivery group currently](https://lists.cncf.io/g/cncf-sig-app-delivery/topic/operator_definition/44377945). 

**Q. Why are Operators useful?**

A. Operators offer Kubernetes-native way to implement your platform automation by extending Kubernetes for your platform workflows. It allows Operator developers to easily share this automation with the broader community enabling composability and reuse. Ultimately this approach makes it possible to reduce in-house custom platform automation and at the same time offers a declarative way to define the platform stacks using Kubernetes YAMLs.

**Q. How are Kubernetes Operators different than Helm charts?**

A. Operators and Helm charts serve different purpose. Helm is a tool that simplifies deploying Kubernetes manifests to a cluster. It focuses on templatization, reusability, etc. of Kubernetes YAMLs. Helm does not handle CRUD operations on Custom Resources. On the other hand, Operators have knowledge about what it means to perform CRUD operations on Custom Resources. In fact, Helm is a very good mechanism to package and deploy Operators. Helm can also be used to deploy Kubernetes manifests that include Custom Resources introduced by Operators.

**Q. What is the difference between an Operator and a Custom Controller?**

A. Custom Controller which is not part of an Operator is something that understands only Kubernetes’s native abstractions such as Pod, Deployment, Service, etc. A Custom Controller that is part of an Operator also understands Custom Resource abstractions that the Operator has introduced (e.g. MysqlCluster).

**Q. Where does an Operator run?**

A. An Operator runs inside of a Kubernetes cluster.

**Q. What is needed in Kubernetes to run an Operator?**

A. To run an Operator, you need the ‘Custom Resource Definition (CRD)’ meta REST API enabled in your cluster. A CRD enables registering custom REST APIs into a Kubernetes cluster. If your Kubernetes cluster version is 1.9 or above then you already have this meta REST API enabled.

**Q. How to run an Operator in a cluster?**

A. To run an Operator, you need to first create a container image of your Operator code (essentially, the image that packages Custom Controller code). Then use a CRD object to register into your cluster’s set of APIs, the Custom Resource API that your Operator is managing. Then create a Deployment manifest with the container image that you built. Finally, use kubectl or helm to apply/install the Deployment manifest in the cluster.

**Q. What permissions are required to deploy an Operator?**

A. The CRD object is not namespaced. It is available at the Cluster scope. So in order to register CRD for your Operator, the user needs ClusterRole permission on the customresourcedefinition object. If in your setup regular users do not have Cluster scope permissions then the cluster admin will have to install the CRD object. Deploying the Operator’s container Pod requires Role level permission on the deployment object. This permission can be granted to a regular user. So the Operator Pod can be created by a regular user. It is easier if both steps - installing the CRD and installing Operator Pod - are done by the same user to simplify Operator deployment.

**Q. What permissions are required for Operator Pods?**

A. This depends on the type of actions that the Operator’s custom controller is designed to perform. If an Operator’s controller is creating Kubernetes resources that are Cluster-level then the Operator’s Pod will need corresponding ClusterRole permissions. For instance, if the controller is creating PersistentVolumes then the Pod will need Cluster-level permission for managing PersistentVolumes. You will need to create a Service Account and grant it such permissions. Then use this Service Account in your Operator Pod Spec.

**Q. Can an Operator be deployed in a particular namespace?**

A. Operator Pod can be deployed in any namespace. The CRD object is Cluster wide though.

**Q. Can one Operator inherit another Operator?**

A. No, there is no inheritance in the style of Object-Oriented languages between Operators.

**Q. Are there any standards emerging around Operators and Custom Resources?**

A. Not yet. The only standard right now is that the Custom Resource Spec definition should follow Kubernetes Spec definition format. Operator developers can define their own Custom Resource Spec properties.
We have defined an [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) 
that defines guidelines towards developing Operators that can be used in different scenarios and setups.

**Q. What is the purpose of Operators that provision Cloud Services like AWS RDS?**

A. Such Operators can be used to provision Cloud-based managed services directly using kubectl.

**Q. Will running MySQL on Kubernetes using a MySQL Operator provide same level of robustness as a managed database service like AWS RDS or Google CloudSQL?**

A. It depends on how the Operator’s code is written. You can choose from multiple available MySQL community Operators or develop your own or customize a community Operator further as per your internal requirements. 

**Q. What are all the personas involved in developing, installing, using Operators? What tools exist for each?**

A. The three personas involved are - Operator developers/Operator curators, Cluster Admins/DevOps Engineers, Application/Microservice developers. The tools available for each persona are listed in following table:

- Operator developer/Curator (Developing/Customizing Operators)
  - [sample-controller](https://github.com/kubernetes/sample-controller), [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), [Operator SDK](https://github.com/operator-framework/operator-sdk), [Helm](https://helm.sh/)
  - [Operator guidelines](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) to enable Platform-as-Code (Bringing consistency across Operators)
- Cluster Admin/ DevOps engineer (Installing Operator)
  - kubectl, Helm, [Operator Lifecycle Manager(OLM)](https://github.com/operator-framework/operator-lifecycle-manager)
- Application developer (Using Custom Resources introduced by Operators)
  - [KubePlus Custom Resource Discovery and Binding Add-on](https://github.com/cloud-ark/kubeplus) (Language and New endpoints for consuming Custom Resources introduced by Operators)
  - [Application CRD](https://github.com/kubernetes-sigs/application) (Abstracting application)

**Q. In what language are Operators written? Is there a preferred language?**

A. Operators can be written in any language. Currently there are [officially supported Kubernetes libraries](https://kubernetes.io/docs/reference/using-api/client-libraries/) for Go, Python, Java, dotnet, JavaScript. There exist [sample-controller](https://github.com/kubernetes/sample-controller), [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), [Operator SDK](https://github.com/operator-framework/operator-sdk) which help in Operator development. These are Golang based.


**Q. Are there situations where one needs multiple Operators?**

A. Often. We are seeing enterprises use multiple Operators to build their platform stacks on Kubernetes. At CloudARK, we have pioneered Platform-as-Code approach for creating multi Operator platform stacks.

**Q. What are the advantages of using multiple Operators?**

A. Development and maintenance of in-house platform automation is reduced. This approach helps you in your goal of moving towards multi-cloud application portability as this significantly simplifies creating or transferring your workloads from one cloud environment to other. 

**Q. What are the challenges when using multiple Operators together?**

A. Primary challenge is interoperability and binding between various Custom Resources supported by Operators. We have developed KubePlus API Add-on for Custom Resource Discovery and Binding that helps with this process.

**Q. We are interested in building our Kubernetes platform using Open source Operators. What should we look for when choosing an Operator?**

A. There are Operator listing and repositories such as following: https://operatorhub.io/, https://chartmuseum.com/, https://github.com/operator-framework/awesome-operators, https://kubedex.com/operators/. You can pick an Operator from any of these places. Then use the [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) that we have developed to evaluate whether the Operator that you have selected has some of the attributes mentioned in the 
Operator Maturity Model.

**Q. Are Operators production ready? Is there any analysis of open source Operators?**

A. We have done analysis of open source Operators. You can find it [here](https://medium.com/@cloudark/analysis-of-open-source-kubernetes-operators-f6be898f2340).

**Q. Are there situations where Operator pattern cannot be used? Or, what is an Operator anti-pattern?**

A. Operator pattern is not suitable if there is no need for monitoring and reconciling the cluster state based on some declarative input leveraging Custom Resource abstraction. 
Operator pattern is also not suitable if actions that need to be performed on the cluster can not be translated into declarative input and instead are imperative in nature by design. For handling such a requirement, you can either create a separate Pod and run it in the cluster, or if you need the functionality through kubectl then consider using Kubernetes Aggregated API Server.
