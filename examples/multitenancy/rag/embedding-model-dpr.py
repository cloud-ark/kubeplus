# Source ChatGPT

from transformers import DPRContextEncoder, DPRTokenizer

# Load the pretrained DPR model
tokenizer = DPRTokenizer.from_pretrained("facebook/dpr-ctx_encoder-single-nq-base")
encoder = DPRContextEncoder.from_pretrained("facebook/dpr-ctx_encoder-single-nq-base")

# Example documents
documents = [
    "The capital of France is Paris.",
    "Python is a programming language.",
    "Hugging Face is a popular platform for NLP models.",
    "The Eiffel Tower is located in Paris."
]

# Tokenize and encode documents
encoded_inputs = tokenizer(documents, padding=True, truncation=True, return_tensors="pt")
document_embeddings = encoder(**encoded_inputs).pooler_output

# Print embeddings
print(document_embeddings)

