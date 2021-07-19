FROM ubuntu:20.04

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata && apt-get install -y python3-pip python-setuptools  curl wget tar sudo apt-transport-https ca-certificates socat python-yaml vim

RUN curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg && echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list && apt-get update -y && apt-get install -y kubectl

ADD requirements.txt /root/requirements.txt
ADD consumerui.py /root/consumerui.py
ADD templates /root/templates
RUN cd /root; pip install -r requirements.txt
RUN mkdir /root/.kube/

EXPOSE 5000
CMD ["python3", "/root/consumerui.py"]
