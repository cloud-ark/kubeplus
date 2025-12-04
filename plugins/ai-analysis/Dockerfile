FROM ollama/ollama:latest

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    python3 python3-venv python3-pip python3-full ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

RUN python3 -m venv /opt/venv
ENV PATH="/opt/venv/bin:${PATH}"

RUN pip install --no-cache-dir flask ollama requests

WORKDIR /app
COPY app.py /app/
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENV MODEL_NAME=llama3
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
