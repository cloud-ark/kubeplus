FROM ubuntu:20.04 
COPY helm /root/
COPY kubectl /root/
COPY helm-pod /root/
RUN apt-get update && apt-get install wget curl vim python -y && mkdir /.helm && mkdir -p /.helm/repository && mkdir /.helm/repository/cache && mkdir -p /.helm/cache/archive && mkdir -p /.helm/cache/plugins && chmod +x /root/helm && chmod +x /root/kubectl && wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz && gunzip kubeplus-kubectl-plugins.tar.gz && tar -xvf kubeplus-kubectl-plugins.tar && cp /plugins/* bin/ && cp /root/helm bin/. && cp /root/kubectl bin/.
COPY repositories.yaml /.helm/repository/
COPY cloudark-helm-charts-index.yaml /.helm/repository/cache/
ENTRYPOINT ["/root/helm-pod"]
