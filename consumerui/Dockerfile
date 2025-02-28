# Use a minimal base image
FROM ubuntu:20.04

# Set environment variables for non-interactive installs and paths
ENV DEBIAN_FRONTEND=noninteractive \
    KUBEPLUS_HOME=/root

# Update and install only required packages in a single RUN to minimize layers
RUN apt-get update -y && apt-get install -y --no-install-recommends \
    tzdata \
    python3-pip \
    python3-setuptools \
    curl \
    wget \
    tar \
    sudo \
    apt-transport-https \
    ca-certificates \
    socat \
    python3-yaml \
    vim \
    graphviz && \
    # Clean up to reduce image size
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

# Create symbolic link for Python
RUN ln -s /usr/bin/python3 /usr/bin/python

# Install KubePlus plugins
RUN wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz && \
    tar -xzf kubeplus-kubectl-plugins.tar.gz && \
    mv plugins/* /usr/local/bin/ && \
    rm -rf kubeplus-kubectl-plugins.tar.gz kubeplus-kubectl-plugins.tar

# Add application source files
WORKDIR /root
ADD requirements.txt /root/
ADD consumerui.py /root/
ADD grapher.py /root/
ADD templates /root/templates/
ADD static /root/static/

# Install Python dependencies
RUN pip3 install --no-cache-dir -r requirements.txt

# Create necessary directories
RUN mkdir -p /root/.kube/

# Expose the application port
EXPOSE 5000

# Set the entrypoint
CMD ["python3", "/root/consumerui.py"]
