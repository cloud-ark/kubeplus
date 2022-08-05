FROM ubuntu:20.04
RUN apt-get update && apt-get install -y openssl jq python3 python3-pip && pip3 install pyyaml
ADD webhook-create-self-signed-ca-cert.sh /
COPY kubectl /root/
COPY kubeplus-non-pod-resources.yaml /root/.
COPY mutatingwebhook.yaml /root/.
#COPY webhook-patch-ca-bundle-new.sh /root/.
COPY kubeconfiggenerator.py /root/.
RUN cp /root/kubectl bin/. && chmod +x /root/kubectl && chmod +x bin/kubectl 
ENTRYPOINT ["/webhook-create-self-signed-ca-cert.sh"]
