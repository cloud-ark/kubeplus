FROM ubuntu:20.04
ADD webhook-create-signed-cert-new.sh /
COPY kubectl /root/
COPY kubeplus-non-pod-resources.yaml /root/.
COPY mutatingwebhook.yaml /root/.
COPY webhook-patch-ca-bundle-new.sh /root/.
RUN apt-get update && apt-get install -y openssl jq
RUN cp /root/kubectl bin/. && chmod +x /root/kubectl && chmod +x bin/kubectl
ENTRYPOINT ["/webhook-create-signed-cert-new.sh"]
