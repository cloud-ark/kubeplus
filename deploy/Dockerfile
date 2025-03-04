FROM ubuntu:22.04
USER root
RUN apt-get update && apt-get upgrade && apt-get install -y curl openssl jq python3 python3-pip && pip3 install pyyaml
ADD webhook-create-self-signed-ca-cert.sh /
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN install -o root -g root -m 0755 kubectl bin/kubectl
RUN cp bin/kubectl /root/kubectl
COPY kubeplus-non-pod-resources.yaml /root/.
COPY mutatingwebhook.yaml /root/.
#COPY webhook-patch-ca-bundle-new.sh /root/.
COPY kubeconfiggenerator.py /root/.
#COPY kubectl /root/
#RUN cp /root/kubectl bin/.  
RUN chmod +x /root/kubectl && chmod +x bin/kubectl 
ENTRYPOINT ["/webhook-create-self-signed-ca-cert.sh"]
