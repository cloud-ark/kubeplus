# Kubernetes Operator Maturity Model for multi-Operator Stacks

We have developed Kubernetes Operator Maturity Model for Kubernetes native stacks with two broad goals:

1) Help Kubernetes Operator Development teams in developing Enterprise-ready
   Kubernetes Operators that are good citizens in multi-Operator world.

2) Help DevOps teams in selecting the Operators for running their workloads 
   on Kubernetes.

A Kubernetes Operator extends Kubernetes control plane to perform domain-specific workflow actions using declarative resource definitions. Operators are being built today for a variety of softwares such as MySQL, Postgres, Cassandra, Airflow, Kafka, Prometheus, Moodle, Wordpress, etc. Increasingly enterprises are running diverse workloads like SaaS, AI, Analytics, ML, Edge networking, CI/CD etc. on Kubernetes. It is becoming common to use more than one Operator in order to build Kubernetes native stacks. The typical challenges that enterprises face today include figuring out how to use various Custom and Kubernetes's built-in resources to create the workflows required for running their workloads on Kubernetes, ensuring that the workflows are portable across Kubernetes distributions and cloud providers, etc.

There is a need for enterprises to understand the enterprise readiness of Operators and how to use them to build the workflows required for running diverse workloads on Kubernetes. The Kubernetes Operator maturity model addresses this need. This model has emerged from our experience of working alongside Operator authors as well as enterprises who are adopting Kubernetes and are using Operators. The model consists of set of guidelines related to consumability, configurability, security, robustness, debuggability and portability of Operators. The guidelines are available [here](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md). Operators that satisfy these properties are easy to consume in multi-Operator setups, support multi-tenant workloads and are portable across cloud providers.

![](./docs/Maturity-Model.jpg)

If you are an Operator author, use this model as a guiding framework when developing your Operator to fit real-life multi-Operator stacks. If you are a Platform Engineer/DevOps Engineer, use this model for evaluating Operators for your platform needs. 




