FROM ubuntu:20.04
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata &&  apt-get install -y graphviz python3-pip
WORKDIR /root/
ADD . /root/
RUN pip3 install -r requirements.txt
ENTRYPOINT ["python3", "connections.py"]


