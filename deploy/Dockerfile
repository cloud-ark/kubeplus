FROM ubuntu:20.04
ADD webhook-create-signed-cert-new.sh /
COPY kubectl /root/
COPY kubeplus-non-pod-resources.yaml /root/.
COPY mutatingwebhook.yaml /root/.
COPY webhook-patch-ca-bundle-new.sh /root/.
COPY kubeconfiggenerator.py /root/.
COPY looper.sh /root/.
RUN apt-get update && apt-get install -y openssl jq python3 python3-pip && pip3 install pyyaml
RUN cp /root/kubectl bin/. && chmod +x /root/kubectl && chmod +x bin/kubectl && chmod +x /root/looper.sh
ENTRYPOINT ["/webhook-create-signed-cert-new.sh"]
#ENTRYPOINT ["/root/looper.sh"]