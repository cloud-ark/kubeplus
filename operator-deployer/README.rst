===================
Operator Deployer
===================

Helper utility to deploy Operator Helm charts.

Note
=====

Earlier versions of KubePlus used operator-deployer to deploy Operator Helm charts.

Current versions of KubePlus does not use operator-deployer. It relies on Helm for Operator deployment.

If you want to use operator-deployer build it using following commands:

dep ensure

./build-local-deploy-artifacts.sh

Then uncomment ../kubeplus/deploy/rc.yaml to include operator-deployer container as part of KubePlus rc manifest.