FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY rag_server.py .
COPY index.pkl .

EXPOSE 8000

CMD ["python", "rag_server.py", "--http"]
