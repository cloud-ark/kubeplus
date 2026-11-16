"""
KubePlus RAG MCP server.

Exposes one tool, search_kubeplus_docs, that does TF-IDF similarity
search over docs + example manifests from the KubePlus GitHub repo
(https://github.com/cloud-ark/kubeplus). Run ingest.py first to build
index.pkl.

Run:
    python rag_server.py            # stdio transport
    python rag_server.py --http     # streamable-http on :8000
"""

import pickle
import sys
from pathlib import Path

from fastmcp import FastMCP
from sklearn.metrics.pairwise import cosine_similarity

INDEX_PATH = Path(__file__).parent / "index.pkl"

with open(INDEX_PATH, "rb") as f:
    _index = pickle.load(f)

mcp = FastMCP("kubeplus-rag")


@mcp.tool
def search_kubeplus_docs(query: str, top_k: int = 3) -> list[dict]:
    """
    Search KubePlus documentation and example manifests for content
    relevant to `query`. Returns the top_k most relevant chunks with
    their source file path, for use as retrieved context in answers
    about KubePlus (the Kubernetes Operator for building custom
    Kubernetes-native platforms / SaaS Managers).
    """
    query_vec = _index["vectorizer"].transform([query])
    scores = cosine_similarity(query_vec, _index["matrix"]).flatten()
    top_idx = scores.argsort()[::-1][:top_k]

    results = []
    for i in top_idx:
        if scores[i] <= 0:
            continue
        results.append(
            {
                "source": _index["sources"][i],
                "score": round(float(scores[i]), 4),
                "text": _index["chunks"][i],
            }
        )
    return results


if __name__ == "__main__":
    if "--http" in sys.argv:
        mcp.run(transport="http", host="0.0.0.0", port=8000)
    else:
        mcp.run()  # stdio
