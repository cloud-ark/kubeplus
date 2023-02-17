FROM ubuntu:20.04
RUN apt-get update && apt-get install -y openssl curl jq python3 python3-pip && pip3 install pyyaml
ADD deploy/webhook-create-self-signed-ca-cert.sh /
RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && cd /root/ && curl -LO "https://dl.k8s.io/release/v1.26.0/bin/linux/${arch}/kubectl"
COPY deploy/kubeplus-non-pod-resources.yaml /root/.
COPY deploy/mutatingwebhook.yaml /root/.
#COPY webhook-patch-ca-bundle-new.sh /root/.
COPY deploy/kubeconfiggenerator.py /root/.
RUN cp /root/kubectl bin/. && chmod +x /root/kubectl && chmod +x bin/kubectl 
ENTRYPOINT ["/webhook-create-self-signed-ca-cert.sh"]
