# Custom Hello World app

from flask import Flask
from prometheus_client import Counter, generate_latest, CONTENT_TYPE_LATEST

app = Flask(__name__)

# Prometheus metrics (counts requests to "/" and "/bye")
HELLO_REQUEST_COUNT = Counter("hello_requests_total", "Total requests to hello endpoint")
BYE_REQUEST_COUNT = Counter("bye_requests_total", "Total requests to bye endpoint")

@app.route("/")
def hello():
    HELLO_REQUEST_COUNT.inc()  # increment counter every time / is hit
    return "Hello World, from Kubernetes!<br>"

@app.route("/bye")
def bye():
    BYE_REQUEST_COUNT.inc()  # increment counter every time /bye is hit
    return "Bye, from Kubernetes!<br>"

# Prometheus metrics endpoint
@app.route("/metrics")
def metrics():
    return generate_latest(), 200, {"Content-Type": CONTENT_TYPE_LATEST}

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
