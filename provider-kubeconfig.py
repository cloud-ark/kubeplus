import json
import subprocess
import os
import yaml
import argparse
from kubernetes import client, config, watch
from kubernetes.client.rest import ApiException
from urllib.parse import urlparse, urlunparse
import base64
import socket


from logging.config import dictConfig

dictConfig({
    'version': 1,
    'formatters': {'default': {
        'format': '[%(asctime)s] %(levelname)s in %(module)s: %(message)s',
    }},
    'handlers': {
     'file.handler': {
            'class': 'logging.handlers.RotatingFileHandler',
            'filename': 'provider-kubeconfig.log',
            'maxBytes': 10000000,
            'backupCount': 5,
            'level': 'DEBUG',
        },
    },
    'root': {
        'level': 'INFO',
        'handlers': ['file.handler']
    }
})

DEFAULT_KUBECONFIG_PATH = "~/.kube/config"
PROVIDER_KUBECONFIG = "kubeplus-saas-provider.json"

class KubeconfigGenerator(object):

    def run_command(self, cmd):
            #print("Inside run_command")
            #print(cmd)
            cmdOut = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).communicate()
            out = cmdOut[0]
            err = cmdOut[1]
            #print(out)
            #print("---")
            #print(err)
            return out, err

    def _create_kubecfg_file(self, sa, namespace, filename, token, server):
            # Initialize the kubeconfig dictionary
            kubeconfig = {
                "apiVersion": "v1",
                "kind": "Config",
                "clusters": [
                    {
                        "name": sa,
                        "cluster": {
                            #"certificate-authority-data": ca,
                            "server": server,
                            "insecure-skip-tls-verify": True
                        }
                    }
                ],
                "users": [
                    {
                        "name": sa,
                        "user": {
                            "token": token,
                        }
                    }
                ],
                "contexts": [
                    {
                        "name": sa,
                        "context": {
                            "cluster": sa,
                            "user": sa,
                            "namespace": namespace, 
                        }
                    }
                ],
                "current-context": sa,
            }

            # Write the kubeconfig to a file
            kubeconfig_json = json.dumps(kubeconfig)
            with open(filename, 'w') as f:
                f.write(kubeconfig_json)
            
            print(f"Kubeconfig file '{filename}' created successfully.")

            # Create ConfigMap
            configmap_body = client.V1ConfigMap(
                api_version = "v1",
                kind = "ConfigMap",
                metadata = {"name": sa},
                data = {filename: kubeconfig_json},
            )

            try:
                corev1.create_namespaced_config_map(namespace=namespace, body=configmap_body)
            except ApiException as e:
                print(f"Exception when creating ConfigMap: {e}\n")

    def _apply_consumer_rbac(self, sa, namespace):

        # Cluster role

        rbac_v1 = client.RbacAuthorizationV1Api()

        metadata = client.V1ObjectMeta(name=sa)

        rules = [
            client.V1PolicyRule(api_groups=["*", ""], resources=["*"], verbs=["get", "watch", "list"]),
            client.V1PolicyRule(api_groups=[""], resources=["users", "groups", "serviceaccounts"], verbs=["impersonate"]),
            client.V1PolicyRule(api_groups=[""], resources=["pods/portforward"], verbs=["create", "get"])
        ]

        # Create the ClusterRole object
        cluster_role = client.V1ClusterRole(
            api_version="rbac.authorization.k8s.io/v1",
            kind="ClusterRole",
            metadata=metadata,
            rules=rules
        )

        try:
            rbac_v1.create_cluster_role(body=cluster_role)
            print(f"ClusterRole '{sa}' created successfully.")
        except client.exceptions.ApiException as e:
            print(f"Exception when creating ClusterRole: {e}\n")

        # Cluster role binding

        subject = client.V1Subject(
            kind="ServiceAccount",
            name=sa,
            namespace=namespace
        )

        role_ref = client.V1RoleRef(
            kind="ClusterRole",
            name=sa,
            api_group="rbac.authorization.k8s.io"
        )

        cluster_role_binding = client.V1ClusterRoleBinding(
            api_version="rbac.authorization.k8s.io/v1",
            kind="ClusterRoleBinding",
            metadata=metadata,
            subjects=[subject],
            role_ref=role_ref
        )
        try:
            rbac_v1.create_cluster_role_binding(body=cluster_role_binding)
            print(f"ClusterRoleBinding '{sa}' created successfully.")
        except client.exceptions.ApiException as e:
            print(f"Exception when creating ClusterRoleBinding: {e}\n")

    def _apply_provider_rbac(self, sa, namespace):
        rbac_v1 = client.RbacAuthorizationV1Api()

        api_version="rbac.authorization.k8s.io/v1"
        kind = "ClusterRole"
        roleName = sa
        metadata = client.V1ObjectMeta(name=roleName)

        rule_groups = [
            {
                "api_groups": ["*", ""],
                "resource_groups": ["*"],
                "verbs": ["get", "watch", "list"]
            },
            {
                "api_groups": ["workflows.kubeplus"],
                "resource_groups": ["resourcecompositions", "resourcemonitors", "resourcepolicies", "resourceevents"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"]
            },
            {
                "api_groups": ["rbac.authorization.k8s.io"],
                "resource_groups": ["clusterroles", "clusterrolebindings", "roles", "rolebindings"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"]
            },
            {
                "api_groups": [""],
                "resource_groups": ["pods/portforward"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"]
            },
            {
                "api_groups": ["platformapi.kubeplus"],
                "resource_groups": ["*"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"]
            },
            {
                "api_groups": [""],
                "resource_groups": ["secrets", "serviceaccounts", "configmaps", "events", "persistentvolumeclaims", "serviceaccounts/token", "services", "services/proxy", "endpoints"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"]
            },
            {
                "api_groups": [""],
                "resource_groups": ["namespaces"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch"]
            },
            {
                "api_groups": ["apps"],
                "resource_groups": ["deployments", "daemonsets", "deployments/rollback", "deployments/scale", "replicasets", "replicasets/scale", "statefulsets", "statefulsets/scale"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"]
            },
            {
                "api_groups": [""],
                "resource_groups": ["users", "groups", "serviceaccounts"],
                "verbs": ["impersonate"]
            },
            {
                "api_groups": [""],
                "resource_groups": ["pods", "pods/attach", "pods/exec", "pods/portforward", "pods/proxy", "pods/eviction", "replicationcontrollers", "replicationcontrollers/scale"],
                "verbs": ["get", "list", "create", "update", "delete", "watch", "patch", "deletecollection"]
            },
            {
                "api_groups": ["admissionregistration.k8s.io"],
                "resource_groups": ["mutatingwebhookconfigurations"],
                "verbs": ["get", "create", "delete", "update"]
            },
            {
                "api_groups": ["apiextensions.k8s.io"],
                "resource_groups": ["customresourcedefinitions"],
                "verbs": ["get", "create", "delete", "update", "patch"]
            },
            {
                "api_groups": ["certificates.k8s.io"],
                "resource_groups": ["signers"],
                "resource_names": ["kubernetes.io/legacy-unknown", "kubernetes.io/kubelet-serving", "kubernetes.io/kube-apiserver-client", "cloudark.io/kubeplus"],
                "verbs": ["get", "create", "delete", "update", "patch", "approve"]
            },
            {
                "api_groups": ["*"],
                "resource_groups": ["*"],
                "verbs": ["get"]
            },
            {
                "api_groups": ["certificates.k8s.io"],
                "resource_groups": ["certificatesigningrequests", "certificatesigningrequests/approval"],
                "verbs": ["create", "delete", "update", "patch"]
            },
            {
                "api_groups": ["extensions"],
                "resource_groups": ["deployments", "daemonsets", "deployments/rollback", "deployments/scale", "replicasets", "replicasets/scale", "replicationcontrollers/scale", "ingresses", "networkpolicies"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"]
            },
            {
                "api_groups": ["networking.k8s.io"],
                "resource_groups": ["ingresses", "networkpolicies"],
                "verbs": ["get", "watch", "list", "create", "delete", "update", "patch", "deletecollection"]
            },
            {
                "api_groups": ["authorization.k8s.io"],
                "resource_groups": ["localsubjectaccessreviews"],
                "verbs": ["create"]
            },
            {
                "api_groups": ["autoscaling"],
                "resource_groups": ["horizontalpodautoscalers"],
                "verbs": ["create", "delete", "deletecollection", "patch", "update"]
            },
            {
                "api_groups": ["batch"],
                "resource_groups": ["cronjobs", "jobs"],
                "verbs": ["create", "delete", "deletecollection", "patch", "update"]
            },
            {
                "api_groups": ["policy"],
                "resource_groups": ["poddisruptionbudgets"],
                "verbs": ["create", "delete", "deletecollection", "patch", "update"]
            },
            {
                "api_groups": [""],
                "resource_groups": ["resourcequotas"],
                "verbs": ["create", "delete", "deletecollection", "patch", "update"]
            }
        ]

        rules = []
        all_resources = set()

        for rule_group in rule_groups:
            resources = rule_group["resource_groups"]
            rule = client.V1PolicyRule(
                api_groups=rule_group["api_groups"],
                resources=resources,
                verbs=rule_group["verbs"]
            )
            if "*" not in resources:
                for resource in resources:
                    all_resources.add(resource)

            if "resource_names" in rule_group:
                rule.resource_names = rule_group["resource_names"]
            rules.append(rule)

        cluster_role_body = client.V1ClusterRole(
            api_version=api_version,
            kind=kind,
            metadata=metadata,
            rules=rules
        )

        try:
            rbac_v1.replace_cluster_role(name=sa, body=cluster_role_body)
            print(f"ClusterRole '{sa}' replaced successfully.")
        except client.exceptions.ApiException as e:
            print(f"Error replacing ClusterRole: {e}")

        role_binding = client.V1ClusterRoleBinding(
            api_version="rbac.authorization.k8s.io/v1",
            kind="ClusterRoleBinding",
            metadata=client.V1ObjectMeta(name=sa),
            subjects=[{
                "kind": "ServiceAccount",
                "name": sa,
                "apiGroup": "",
                "namespace": namespace
            }],
            role_ref={
                "kind": "ClusterRole",
                "name": sa,
                "apiGroup": "rbac.authorization.k8s.io"
            }
        )
        try:
            rbac_v1.replace_cluster_role_binding(name=sa, body=role_binding)
            print(f"ClusterRole '{sa}' replaced successfully.")
        except client.exceptions.ApiException as e:
            print(f"Error replacing ClusterRole: {e}")

        config_map_body = client.V1ConfigMap(
            metadata=client.V1ObjectMeta(name="kubeplus-saas-provider-perms"), 
            data={"kubeplus-saas-provider-perms.txt": "\n".join(all_resources)}
        )

        corev1.create_namespaced_config_map(
            namespace=namespace,
            body=config_map_body
        )

    def _update_rbac(self, permissionfile, sa, namespace, kubeconfig):
        rbac_v1 = client.RbacAuthorizationV1Api()
        role = {}
        role["apiVersion"] = "rbac.authorization.k8s.io/v1"
        role["kind"] = "ClusterRole"
        metadata = {}
        metadata["name"] = sa + "-update"
        role["metadata"] = metadata
        
        ruleList = []
        ruleGroup = {}

        fp = open(permissionfile, "r")
        data = fp.read()
        perms_data = json.loads(data)
        perms = perms_data["perms"]
        new_resources_set = set()
        for apiGroup, res_actions in perms.items():
            for res in res_actions:
                for resource, verbs in res.items():
                    print(apiGroup + " " + resource + " " + str(verbs))
                    if resource not in new_resources_set:
                        new_resources_set.add(resource.strip())
                    ruleGroup = {}
                    if apiGroup == "non-apigroup":
                        if 'nonResourceURL' in resource:
                            parts = resource.split("nonResourceURL::")
                            nonRes = parts[0].strip()
                            ruleGroup['nonResourceURLs'] = [nonRes]
                            ruleGroup['verbs'] = verbs
                    else:
                        ruleGroup["apiGroups"] = [apiGroup]
                        ruleGroup["verbs"] = verbs
                        if 'resourceName' in resource:
                            parts = resource.split("/resourceName::")
                            resNameParent = parts[0].strip()
                            resName = parts[1].strip()
                            ruleGroup["resources"] = [resNameParent]
                            ruleGroup["resourceNames"] = [resName]
                        else:
                            ruleGroup["resources"] = [resource]

            
                    ruleList.append(ruleGroup)

        role["rules"] = ruleList

        roleName = sa + "-update-role.yaml"
        filePath = os.getcwd() + "/" + roleName
        fp = open(filePath, "w")
        yaml_content = yaml.dump(role)
        fp.write(yaml_content)
        fp.close()
        cmd = " kubectl apply -f " + filePath + '--kubeconfig ' + kubeconfig
        self.run_command(cmd)

        role_binding = client.V1ClusterRoleBinding(
            api_version="rbac.authorization.k8s.io/v1",
            kind="ClusterRoleBinding",
            metadata=client.V1ObjectMeta(name=sa + "-update"),
            subjects=[{
                "kind": "ServiceAccount",
                "name": sa,
                "apiGroup": "",
                "namespace": namespace
            }],
            role_ref={
                "kind": "ClusterRole",
                "name": sa + "-update",
                "apiGroup": "rbac.authorization.k8s.io"
            }
        )
        
        rbac_v1.create_cluster_role_binding(body=role_binding)

        # Read configmap to get earlier permissions; delete it and create it with all new permissions:
        cmd = "kubectl get configmap kubeplus-saas-provider-perms -o json -n " + namespace
        out1, err1 = self.run_command(cmd)
        print("Original Perms Out:" + str(out1))
        print("Perms Err:" + str(err1))
        kubeplus_perms = set()
        if out1 != '':
            json_op = json.loads(out1)
            perms = json_op['data']['kubeplus-saas-provider-perms.txt']
            print(perms)
            k_perms = perms.split(",")
            for p in k_perms:
                p = p.replace("'","")
                p = p.replace("[","")
                p = p.replace("]","")
                p = p.strip()
                kubeplus_perms.add(p)

        new_resources_set.update(kubeplus_perms)

        print("New perms:" + str(list(new_resources_set)))

        try:
            corev1.delete_namespaced_config_map(name="kubeplus-saas-provider-perms", namespace=namespace)
            print("ConfigMap deleted successfully.")
        except Exception as e:
            print(f"Error deleting ConfigMap: {e}")

        # create configmap to store all resources
        config_map_body = client.V1ConfigMap(
            metadata=client.V1ObjectMeta(name="kubeplus-saas-provider-perms"), 
            data={"kubeplus-saas-provider-perms.txt": "\n".join(new_resources_set)}
        )

        corev1.create_namespaced_config_map(
            namespace=namespace,
            body=config_map_body
        )


    def _apply_rbac(self, sa, namespace, entity=''):
            if entity == 'provider':
                    self._apply_provider_rbac(sa, namespace)
            if entity == 'consumer':
                    self._apply_consumer_rbac(sa, namespace)

    def _create_secret(self, sa, namespace):

        # Define annotations
        annotations = {
            'kubernetes.io/service-account.name': sa
        }

        # Define metadata
        metadata = client.V1ObjectMeta(
            name=sa,
            namespace=namespace,
            annotations=annotations
        )

        # Define secret object
        secret = client.V1Secret(
            api_version="v1",
            kind="Secret",
            metadata=metadata,
            type='kubernetes.io/service-account-token'
        )

        # Apply secret to the Kubernetes cluster
        try:
            corev1.create_namespaced_secret(namespace, secret)
            # Watch for the creation of the secret
            w = watch.Watch()
            for event in w.stream(corev1.list_namespaced_secret, namespace=namespace, watch=False):
                if event['type'] == 'ADDED' and event['object'].metadata.name == sa:
                    break
            print(f"Secret '{sa}' created successfully.")
        except ApiException as e:
            return e
        return None

    def _extract_kubeconfig(self, sa, namespace):
        # Retrieve the secret
        try:
            secret = corev1.read_namespaced_secret(name=sa, namespace=namespace)
        except ApiException as e:
            print("Exception when calling CoreV1Api->read_namespaced_secret: %s\n" % e)

        # Extract the token from the secret
        token = base64.b64decode(secret.data.get('token')).decode('utf-8')
        if not token:
            print(f'Token not found in the secret: {sa}.')
            exit(1)

        ca_cert = secret.data.get('ca.crt')
        if not ca_cert:
            print(f'CA cert not found in the secret: {sa}.')
            exit(1)
        return token, ca_cert

    def _generate_kubeconfig(self, sa, namespace, apiserver_ip):
            body = client.V1ServiceAccount(metadata={"name": sa})

            try:
                # Create ServiceAccount in the namespace
                corev1.create_namespaced_service_account(namespace, body)
                print(f"ServiceAccount '{sa}' created successfully.")
            except ApiException as e:
                print(f"Error: Failed to create ServiceAccount '{sa}': {e.reason}")

            err = self._create_secret(sa, namespace)
            if err is None:
                token, ca_cert = self._extract_kubeconfig(sa, namespace)
            else:
                print(f"Error: Failed to create Secret '{sa}': {err.reason}")
                exit(1)
            
            self._create_kubecfg_file(sa, namespace, filename, token, apiserver_ip)

    def delete_sa(self, sa_name, namespace):
        try:
            corev1.delete_namespaced_service_account(name=sa_name, namespace=namespace)
            print(f"ServiceAccount {sa_name} deleted successfully.")
        except ApiException as e:
            print(f"Error deleting ServiceAccount {sa_name}: {e}")

    def delete_config_map(self, cm_name, namespace):
        try:
            corev1.delete_namespaced_config_map(name=cm_name, namespace=namespace)
            print(f"ConfigMap {cm_name} deleted successfully.")
        except ApiException as e:
            print(f"Error deleting ConfigMap {cm_name}: {e}")

    def delete_cluster_role(self, role_name):
        try:
            rbacv1.delete_cluster_role(name=role_name)
            print(f"ClusterRole {role_name} deleted successfully.")
        except ApiException as e:
            print(f"Error deleting ClusterRole {role_name}: {e}")

    def delete_cluster_role_binding(self, binding_name):
        try:
            rbacv1.delete_cluster_role_binding(name=binding_name)
            print(f"ClusterRoleBinding {binding_name} deleted successfully.")
        except ApiException as e:
            print(f"Error deleting ClusterRoleBinding {binding_name}: {e}")

def check_and_create_namespace(namespace):
    try:
        # Check if the namespace exists
        corev1.read_namespace(namespace)
        print(f'Namespace: {namespace} already exisit')
    except ApiException as e:
        if e.status == 404:
            # Namespace doesn't exist, create it
            body = client.V1Namespace(metadata=client.V1ObjectMeta(name=namespace))
            try:
                corev1.create_namespace(body)
                print(f"Namespace '{namespace}' created successfully.")
            except ApiException as ex:
                print(f"Error: Failed to create namespace '{namespace}': {ex.reason}")
        else:
            print(f"Error: Failed to check namespace '{namespace}': {e.reason}")

def label_namespace(namespace):
    try:
        # Get current labels of the namespace
        namespace_obj = corev1.read_namespace(namespace)

        # Add/overwrite the 'managedby=kubeplus' label
        if not namespace_obj.metadata.labels:
            namespace_obj.metadata.labels = {}
        namespace_obj.metadata.labels['managedby'] = 'kubeplus'

        # Update the namespace with the new label
        corev1.patch_namespace(namespace, namespace_obj)
        print(f"Namespace '{namespace}' labeled successfully.")
    except ApiException as e:
        print(f"Error: Failed to label namespace '{namespace}': {e.reason}")


class ValidatePermissionFiles(argparse.Action):
    def __call__(self, parser, namespace, values, option_string=None):
        setattr(namespace, self.dest, values)
        if getattr(namespace, 'action', None) != 'update':
            parser.error(f"Permission file is required when using 'update' action: --permission-file\
                         {permission_help}")

class ValidateKubeconfigFiles(argparse.Action):
    def __call__(self, parser, namespace, values, option_string=None):
        setattr(namespace, self.dest, values)
        if getattr(namespace, 'action', None) != 'delete':
            parser.error(f"Please provide the kubeconfig file path with 'delete' argument.")

def ensure_https_scheme(apiserver_ip):
    # Ensures the given URL has an 'https' scheme. If the URL does not have a scheme, 'https' is added as the scheme.
    parsed_url = urlparse(apiserver_ip)
    if not parsed_url.scheme:
        parsed_url = parsed_url._replace(scheme="https")
    return urlunparse(parsed_url)

def delete_file(file_path):
    """Safely remove a file if it exists."""
    try:
        if os.path.exists(file_path):
            os.remove(file_path)
            print(f"Removed {file_path}")
        else:
            print(f"{file_path} does not exist.")
    except Exception as e:
        print(f"Error removing {file_path}: {e}")

def check_ip_port(url):
    parsed_url = urlparse(url)
    ip = parsed_url.hostname
    port = parsed_url.port
    try:
        # Create a new socket
        with socket.create_connection((ip,port), timeout=10) as sock:
            return True
    except (socket.timeout, socket.error):
        return False

if __name__ == '__main__':

    parser = argparse.ArgumentParser()

    parser.add_argument("action", help="command", choices=['create', 'delete', 'update', 'extract'])
    
    parser.add_argument("namespace", help="namespace in which KubePlus will be installed.")
    
    parser.add_argument("-k", "--kubeconfig", default=DEFAULT_KUBECONFIG_PATH, help='This flag is used to specify the path\
                        of the kubeconfig file that should be used for executing steps in provider-kubeconfig.\
                        (default: %(default)s)', action=ValidateKubeconfigFiles)
    
    parser.add_argument("-s", "--apiserver-url", dest="apiserverurl", default='', help='This flag is to be used to pass the API Server URL of the\
                        API server on which KubePlus is installed. This API Server URL will be used in constructing the server endpoint in\
                        the provider kubeconfig. Use the command `kubectl config view --minify -o jsonpath="{.clusters[0].cluster.server}"`\
                        to retrieve the API Server URL.')
    
    parser.add_argument("-f", "--filename", default=PROVIDER_KUBECONFIG, help='This flag is used to specify the output file name\
                        in which generated provider kubeconfig will be store\
                        (default: %(default)s)')
    
    permission_help = '''
permissions file (Only for 'update' command). Should be a JSON file with the following structure:
{
"perms": {
    "<apiGroup1>": [
        {
            "resource1|resource/resourceName::<resourceName>": [
            "verb1",
            "verb2",
            "..."
            ]
        },
        {
            "resource2": [
            "..."
            ]
        }
    ],
    "<apiGroup2>": [
        {
            "resource3": [
            "..."
            ]
        }
    ]
}
}

'''
    parser.add_argument("-p", "--permission-file", dest="permissionfile", default='', type=str, help=permission_help, action=ValidatePermissionFiles)
    
    args = parser.parse_args()
    action = args.action
    namespace = args.namespace
    kubeconfigString = args.kubeconfig
    apiserver_ip = args.apiserverurl
    permission_file = args.permissionfile
    filename = args.filename

    config.load_kube_config(config_file=kubeconfigString)
    corev1 = client.CoreV1Api()
    rbacv1 = client.RbacAuthorizationV1Api()


    if not apiserver_ip:
        # Get the API server IP
        configuration = client.Configuration().get_default_copy()
        apiserver_ip = configuration.host

    
    apiserver_ip = ensure_https_scheme(apiserver_ip)
    print(f'API Server IP: {apiserver_ip}')


    if not check_ip_port(apiserver_ip):
        print(f"The API Server: {apiserver_ip} is not accessible.")
        exit(1)

    kubeconfigGenerator = KubeconfigGenerator()
    sa = 'kubeplus-saas-provider'

    if not filename.endswith(".json"):
        filename += ".json"

    if action == "create":
            check_and_create_namespace(namespace)
            label_namespace(namespace)

            # 1. Generate Provider kubeconfig
            kubeconfigGenerator._generate_kubeconfig(sa, namespace, apiserver_ip)
            kubeconfigGenerator._apply_rbac(sa, namespace, entity='provider')
            print("Provider kubeconfig created: " + filename)

    if action == "extract":
            token, ca_cert = kubeconfigGenerator._extract_kubeconfig(sa, namespace)
            kubeconfigGenerator._create_kubecfg_file(sa, namespace, filename, token, apiserver_ip)
            print("Provider kubeconfig created: " + filename)

    if action == "update":
            kubeconfigGenerator._update_rbac(permission_file, sa, namespace, kubeconfigString)
            print("Provider kubeconfig permissions updated: " + filename)

    if action == "delete":
            kubeconfigGenerator.delete_sa(sa, namespace)
            
            # Delete ConfigMap
            kubeconfigGenerator.delete_config_map(sa, namespace)
            
            # Delete ClusterRole
            kubeconfigGenerator.delete_cluster_role(sa)
            
            # Delete ClusterRoleBinding
            kubeconfigGenerator.delete_cluster_role_binding(sa)
            
            # Delete Update ClusterRole
            kubeconfigGenerator.delete_cluster_role(sa + "-update")
            
            # Delete Update ClusterRoleBinding
            kubeconfigGenerator.delete_cluster_role_binding(sa + "-update")
            
            # Delete kubeplus-saas-provider-perms ConfigMap
            kubeconfigGenerator.delete_config_map("kubeplus-saas-provider-perms", namespace)

            cwd = os.getcwd()
            files_to_delete = [
                "kubeplus-saas-provider-secret.yaml",
                filename,
                "kubeplus-saas-provider-role.yaml",
                "kubeplus-saas-provider-update-role.yaml",
                "kubeplus-saas-provider-rolebinding.yaml",
                "kubeplus-saas-provider-update-rolebinding.yaml",
                "kubeplus-saas-provider-perms.txt",
                "kubeplus-saas-provider-perms-update.txt"
            ]

            for file_name in files_to_delete:
                delete_file(cwd + "/" + file_name)