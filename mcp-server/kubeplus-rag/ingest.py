"""
Build a small local RAG index from the KubePlus GitHub repo.

Run once (or whenever the repo changes) to (re)build index.pkl:
    python ingest.py

It clones https://github.com/cloud-ark/kubeplus, indexes the repository
README, documentation and example manifests, chunks them, and fits a TF-IDF vector index.
No external embedding API / model download is required, so this works
fine inside a minikube/offline classroom environment.
"""

import pickle
import subprocess
from pathlib import Path

from sklearn.feature_extraction.text import TfidfVectorizer

REPO_URL = "https://github.com/cloud-ark/kubeplus.git"
CLONE_DIR = Path("./kubeplus")
INDEX_PATH = Path("./index.pkl")
CHUNK_WORDS = 250
CHUNK_OVERLAP = 50


def clone_repo() -> None:
    if CLONE_DIR.exists():
        print(f"Repo already present at {CLONE_DIR}, skipping clone.")
        return
    subprocess.run(
        ["git", "clone", "--depth", "1", REPO_URL, str(CLONE_DIR)], check=True
    )


def iter_source_files():
    """Docs and examples used by the KubePlus documentation RAG index."""
    roots = [
        CLONE_DIR / "docs",
        CLONE_DIR / "examples",
    ]

    # Include the repository README as the main top-level documentation.
    if (CLONE_DIR / "README.md").is_file():
        yield CLONE_DIR / "README.md"

    for root in roots:
        if not root.exists():
            continue

        for ext in ("*.md", "*.rst", "*.txt", "*.yaml", "*.yml"):
            for path in root.rglob(ext):
                if "vendor" in path.parts or ".git" in path.parts:
                    continue
                if "html" in path.relative_to(CLONE_DIR).parts:
                    continue
                yield path


def chunk_text(text: str, size=CHUNK_WORDS, overlap=CHUNK_OVERLAP):
    words = text.split()
    if not words:
        return
    step = max(size - overlap, 1)
    for start in range(0, len(words), step):
        chunk = words[start : start + size]
        if chunk:
            yield " ".join(chunk)
        if start + size >= len(words):
            break


def build_index():
    clone_repo()

    chunks, sources = [], []
    for path in iter_source_files():
        try:
            text = path.read_text(encoding="utf-8", errors="ignore")
        except OSError:
            continue
        rel = str(path.relative_to(CLONE_DIR))
        for chunk in chunk_text(text):
            chunks.append(chunk)
            sources.append(rel)

    print(f"Collected {len(chunks)} chunks from {len(set(sources))} files.")

    vectorizer = TfidfVectorizer(stop_words="english", max_features=20000)
    matrix = vectorizer.fit_transform(chunks)

    with open(INDEX_PATH, "wb") as f:
        pickle.dump(
            {
                "vectorizer": vectorizer,
                "matrix": matrix,
                "chunks": chunks,
                "sources": sources,
            },
            f,
        )
    print(f"Saved index to {INDEX_PATH.resolve()}")


if __name__ == "__main__":
    build_index()
