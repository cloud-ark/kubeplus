#FROM alpine:latest
FROM golang:1.20.1-bullseye as builder
RUN mkdir /build
ADD ./ /build/
WORKDIR /build/platform-operator/helm-pod
RUN  export GO111MODULE=on; export GOOS=linux; go mod vendor; go build .

FROM ubuntu:20.04 
COPY --from=builder /build/platform-operator/helm-pod/helm-pod /root/
RUN apt-get update && apt-get install wget curl vim python -y 
RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
 wget "https://get.helm.sh/helm-v3.11.1-linux-${arch}.tar.gz" && \
 tar xvf "helm-v3.11.1-linux-${arch}.tar.gz" && \
 mv "linux-${arch}/helm" /root/ && rm  "helm-v3.11.1-linux-${arch}.tar.gz" && \
 rm -rf "linux-${arch}"/
RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && cd /root/ && curl -LO "https://dl.k8s.io/release/v1.26.0/bin/linux/${arch}/kubectl"
RUN apt-get update && apt-get install wget curl vim python -y && mkdir /.helm && mkdir -p /.helm/repository && mkdir /.helm/repository/cache && mkdir -p /.helm/cache/archive && mkdir -p /.helm/cache/plugins && chmod +x /root/helm && chmod +x /root/kubectl && wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz && gunzip kubeplus-kubectl-plugins.tar.gz && tar -xvf kubeplus-kubectl-plugins.tar && cp -r /plugins/* bin/ && cp /root/helm bin/. && cp /root/kubectl bin/.
COPY platform-operator/helm-pod/repositories.yaml /.helm/repository/
COPY platform-operator/helm-pod/cloudark-helm-charts-index.yaml /.helm/repository/cache/
ENTRYPOINT ["/root/helm-pod"]
