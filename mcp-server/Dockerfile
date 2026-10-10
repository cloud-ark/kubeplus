FROM python:3.12-slim

RUN apt-get update && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl" \
    && chmod +x kubectl && mv kubectl /usr/local/bin/

RUN curl -fsSL https://get.helm.sh/helm-v3.18.3-linux-amd64.tar.gz \
    | tar -xz \
    && mv linux-amd64/helm /usr/local/bin/helm \
    && rm -rf linux-amd64

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY server.py .

CMD ["uvicorn", "server:app", "--host", "0.0.0.0", "--port", "8000"]
