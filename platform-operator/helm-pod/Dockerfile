FROM ubuntu:20.04 
COPY kubectl /root/
COPY helm-pod /root/
RUN apt-get update && apt-get install wget curl vim python -y && mkdir /.helm && mkdir -p /.helm/repository && mkdir /.helm/repository/cache && mkdir -p /.helm/cache/archive && mkdir -p /.helm/cache/plugins && wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz && gunzip kubeplus-kubectl-plugins.tar.gz && tar -xvf kubeplus-kubectl-plugins.tar && cp -r /plugins/* bin/ && cp /root/kubectl bin/. && chmod +x /root/kubectl && chmod +x /bin/kubectl
RUN wget https://get.helm.sh/helm-v3.12.1-linux-amd64.tar.gz \
    && gunzip helm-v3.12.1-linux-amd64.tar.gz \
    && tar -xvf helm-v3.12.1-linux-amd64.tar \
    && mv linux-amd64/helm bin/.

COPY repositories.yaml /.helm/repository/
COPY cloudark-helm-charts-index.yaml /.helm/repository/cache/
ENTRYPOINT ["/root/helm-pod"]
