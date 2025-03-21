# Source ChatGPT

import faiss
import numpy as np

# Convert the document embeddings to numpy arrays (FAISS requires numpy arrays)
document_embeddings = np.array(document_embeddings).astype('float32')

# Create a FAISS index for similarity search
index = faiss.IndexFlatL2(document_embeddings.shape[1])  # Using L2 distance (Euclidean)

# Add the document embeddings to the index
index.add(document_embeddings)

# Now, the index can be used to search for similar documents

