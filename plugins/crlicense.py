import argparse
import os
import subprocess

from crmetrics import CRBase

class LicenseManager(CRBase):

    def create_license(self, kind, license_file, expiry, num_of_instances, kubeconfig):
        kubeplus_ns = self.get_kubeplus_namespace(kubeconfig)
        kind_lower = kind.lower()
        license_configmap = kind_lower + "-license"
        cmd = "kubectl create configmap " + license_configmap + " --from-file=license_file=" + license_file + " -n " + kubeplus_ns + " --kubeconfig=" + kubeconfig
        out, err = self.run_command(cmd)
        if out != '':
            msg = "License for Kind {kind} created.".format(kind=kind)
            print(msg)
        if err != '':
            msg = "License for Kind {kind} already exists.".format(kind=kind)
            print(msg)

        if expiry != "":
            cmd = "kubectl annotate configmap " + license_configmap + " expiry=" + expiry + " -n " + kubeplus_ns + " --kubeconfig=" + kubeconfig
            self.run_command(cmd)

        if num_of_instances != "":
            cmd = "kubectl annotate configmap " + license_configmap + " allowed_instances=" + num_of_instances + " -n " + kubeplus_ns + " --kubeconfig=" + kubeconfig
            self.run_command(cmd)


if __name__ == '__main__':
    license_mgr = LicenseManager()

    parser = argparse.ArgumentParser()
    parser.add_argument("action", help="action to perform") 
    parser.add_argument("kind", help="Kind name")
    parser.add_argument("licensefile", help="File with license contents.")
    parser.add_argument("-k", "--kubeconfig", help="Provider kubeconfig file.")
    parser.add_argument("-e", "--expiry", help="Expiry date for the license.")
    parser.add_argument("-n", "--appinstances", help="Allowed number of app instances to create.")

    args = parser.parse_args()

    action = args.action
    kind = args.kind
    license_file = args.licensefile

    if not os.path.isfile(license_file):
        print("License file " + license_file + " does not exist.")
        exit(0)

    kubeconfig = ''
    if args.kubeconfig:
        kubeconfig = args.kubeconfig

    if not os.path.isfile(kubeconfig):
        print("Provider kubeconfig file " + kubeconfig  + " does not exist.")
        exit(0)

    expiry = ''
    if args.expiry:
        expiry = args.expiry
        parts = expiry.split("/")
        if len(parts) < 3:
            print("Required expiry date format: MM/DD/YYYY")
            exit(0)

    appinstances = ''
    if args.appinstances:
        appinstances = args.appinstances
        if int(appinstances) < 0:
            print("App instances should be > 0.")
            exit(0)

    if not license_mgr.check_kind(kind, kubeconfig):
        print("Kind " + kind + " does not exist in the cluster in the platformapi.kubeplus api group.")
        exit(0)
    
    if action == "create":
        if expiry == '' and appinstances == '':
            print("Both expiry date and number of app instances to create is empty.")
            print("Specify at least one criteria.")
            exit(0)

        license_mgr.create_license(kind, license_file, expiry, appinstances, kubeconfig)
