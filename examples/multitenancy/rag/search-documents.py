# Source ChatGPT

# Let's say you have a new query to retrieve similar documents
query = "Where is the Eiffel Tower?"
query_embedding = model.encode([query])  # Assuming you're using Sentence-BERT

# Perform the search for the top-k most similar documents
k = 2  # Number of similar documents to retrieve
D, I = index.search(np.array(query_embedding).astype('float32'), k)

# D contains the distances (lower is more similar), and I contains the indices of the retrieved documents
print(f"Distances: {D}")
print(f"Indices of retrieved documents: {I}")

# Retrieve the documents based on the indices
retrieved_documents = [documents[i] for i in I[0]]
print(f"Retrieved documents: {retrieved_documents}")

