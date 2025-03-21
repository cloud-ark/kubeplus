# Source ChatGPT

from sentence_transformers import SentenceTransformer

# Load the pre-trained model (Sentence-BERT)
model = SentenceTransformer('all-MiniLM-L6-v2')  # or use a model like DPR for better retrieval

# Example documents
documents = [
    "The capital of France is Paris.",
    "Python is a programming language.",
    "Hugging Face is a popular platform for NLP models.",
    "The Eiffel Tower is located in Paris."
]

# Convert documents to embeddings
document_embeddings = model.encode(documents)

# Print the embeddings (this will be a list of vectors)
print(document_embeddings)

