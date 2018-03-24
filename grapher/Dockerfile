FROM ubuntu:18.04
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata &&  apt-get install -y graphviz python3-pip
ADD requirements.txt /src/requirements.txt
RUN cd /src; pip3 install -r requirements.txt
ADD . /src
ENTRYPOINT ["python3", "/src/connections.py"]


#RUN pip3 install -r /root/requirements.txt
#ADD connections.py /root/
#WORKDIR /root/
#CMD ["python3", "/src/connections.py"]
#ENTRYPOINT ["python3", "/src/connections.py"]


