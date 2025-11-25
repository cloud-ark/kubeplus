FROM ubuntu:22.04
RUN apt-get update -y && apt-get install -y python-setuptools python3-pip
ADD . /src
RUN pip install -r /src/requirements.txt
CMD ["python3", "/src/app.py"]
