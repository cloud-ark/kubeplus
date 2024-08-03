FROM ubuntu:20.04

RUN apt-get update -y && DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata && apt-get install -y python3-pip python-setuptools  curl wget tar sudo apt-transport-https ca-certificates socat python-yaml vim graphviz

RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
curl -LO https://dl.k8s.io/release/v1.26.0/bin/linux/${arch}/kubectl && \
install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
cp /usr/bin/python3.8 /usr/bin/python

RUN wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz && gunzip kubeplus-kubectl-plugins.tar.gz && tar -xvf kubeplus-kubectl-plugins.tar && cp -r /plugins/* bin/

ENV KUBEPLUS_HOME /

ADD requirements.txt /root/requirements.txt
ADD consumerui.py /root/consumerui.py
ADD grapher.py /root/grapher.py
ADD templates /root/templates
ADD static /root/static
RUN cd /root; pip install -r requirements.txt
RUN mkdir /root/.kube/

EXPOSE 5000
CMD ["python3", "/root/consumerui.py"]
