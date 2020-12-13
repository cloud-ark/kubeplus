Platform-as-Code Resource Annotations on Operator CRDs
------------------------------------------------------


|Sr. No. | Operator      | Github url    | CRD           | Resource Annotations                  |
|--------|:-------------:|:-------------:|:-------------:|:-------------------------------------:
| 1.     | Cassandra (DataStax) | https://github.com/datastax/cass-operator | CassandraDatacenter | resource/annotation-relationship="on:Secret, key:cassandra.datastax.com/watched-by, value:contains(INSTANCE.metadata.name)" 
| | | | | resource/composition="StatefulSet, Service, PodDisruptionBudget"
| | | | | resource/label-relationship="on:PersistentVolumeClaim, key:cassandra.datastax.com/datacenter, value:INSTANCE.metadata.name"
| 2.     | CertManager (JetStack) | https://github.com/jetstack/cert-manager | ClusterIssuer | resource/annotation-relationship="on:Ingress, key:cert-manager.io/cluster-issuer, value:INSTANCE.metadata.name"
| 3.     | MySQL (PressLabs) | https://github.com/presslabs/mysql-operator | MysqlCluster | resource/composition="StatefulSet, Service, ConfigMap, Secret, PodDisruptionBudget"
| 4.     | Multus (Intel) | https://github.com/intel/multus-cni | NetworkAttachmentDefinition | resource/annotation-relationship="on:Pod, key:k8s.v1.cni.cncf.io/networks, value:INSTANCE.metadata.name"
| 5.     | Elasticsearch (Elastic) | https://github.com/elastic/cloud-on-k8s | Elasticsearch | resource/composition="Pod, Service, Secret, ConfigMap"
| 6.     | Redis (RedisLabs) | https://github.com/RedisLabs/redis-enterprise-k8s-docs | RedisEnterpriseCluster | resource/composition="StatefulSet, Service, PodDisruptionBudget"

Here is an example CRD annotation commands:

`
kubectl annotate crd redisenterpriseclusters.app.redislabs.com resource/composition="StatefulSet, Service, PodDisruptionBudget"
`