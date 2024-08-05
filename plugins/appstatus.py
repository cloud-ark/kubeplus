import subprocess
import sys
import json
from crmetrics import CRBase


class AppStatusFinder(CRBase):

    def get_app_instance_status(self, kind, instance, kubeconfig):
        cmd = 'kubectl get %s %s -o json %s' % (kind, instance, kubeconfig)
        out, err = self.run_command(cmd)
        if err != "":
            print("Something went wrong while getting app instance status.")
            print(err)
            exit(1)

        response = json.loads(out)
        if 'status' in response:
            if 'helmrelease' in response['status']:
                helm_release = response['status']['helmrelease'].strip('\n')
                ns, name = helm_release.split(':')
                return name, ns, True
            else:
                # an error has occurred
                status = response['status']
                return status, None, False
            
        else:
            return 'Application not deployed properly', None, False


    def get_app_pods(self, namespace, kubeconfig):
        cmd = 'kubectl get pods -n %s %s -o json' % (namespace, kubeconfig)
        out, err = self.run_command(cmd)
        # format?
        response = json.loads(out)
        pods = []
        for pod in response['items']:
            name = pod['metadata']['name']
            typ = pod['kind']
            ns = pod['metadata']['namespace']
            phase = pod['status']['phase']
            pods.append((name, typ, ns, phase))
        return pods


if __name__ == '__main__':
    appStatusFinder = AppStatusFinder()
    kind = sys.argv[1]
    instance = sys.argv[2]
    kubeconfig = sys.argv[3]

    valid_consumer_api = appStatusFinder.verify_kind_is_consumerapi(kind, kubeconfig)
    if not valid_consumer_api:
        print(("{} is not a valid Consumer API.").format(kind))
        exit(0)

    res_exists, ns, err = appStatusFinder.check_res_exists(kind, instance, kubeconfig)
    if not res_exists:
        print(err)
        exit(0)
    
    working, error = appStatusFinder.validate_kind_and_instance(kind, instance, ns)
    if working == False:
        print(err)
        exit(1)
    
    release_name_or_status, release_ns, deployed = appStatusFinder.get_app_instance_status(kind, instance, kubeconfig)
    
    if deployed:
        deploy_str = 'Deployed'
    else:
        print(release_name_or_status)
        exit(1)

    pods = appStatusFinder.get_app_pods(instance, kubeconfig)

    print("{:<55} {:<55} {:<55} {:<55}".format("NAME", "TYPE", "NAMESPACE", "STATUS"))
    print("{:<55} {:<55} {:<55} {:<55}".format(release_name_or_status, 'helmrelease', release_ns, deploy_str))
    for pod_name, typ, pod_ns, phase in pods:
        print("{:<55} {:<55} {:<55} {:<55}".format(pod_name, typ, pod_ns, phase))