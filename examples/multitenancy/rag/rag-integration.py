# Source ChatGPT

from transformers import RagTokenizer, RagSequenceForGeneration

# Initialize the tokenizer, retriever, and RAG model
tokenizer = RagTokenizer.from_pretrained("facebook/rag-sequence-nq")
retriever = RagRetriever.from_pretrained("facebook/rag-sequence-nq")
model = RagSequenceForGeneration.from_pretrained("facebook/rag-sequence-nq")

# Example query
query = "Where is the Eiffel Tower located?"

# Tokenize the input query
inputs = tokenizer(query, return_tensors="pt")

# Retrieve relevant documents from your FAISS index
query_embedding = model.encode([query])  # Use the same model for the query as before
D, I = index.search(np.array(query_embedding).astype('float32'), k=5)

# Use the indices to get the top-k relevant documents
retrieved_docs = [documents[i] for i in I[0]]

# Convert the retrieved documents into the appropriate format for the RAG model
retrieved_docs_input = tokenizer(retrieved_docs, padding=True, truncation=True, return_tensors="pt")

# Generate the response using RAG with the retrieved context
generated_output = model.generate(input_ids=inputs["input_ids"], context_input_ids=retrieved_docs_input["input_ids"])

# Decode and print the response
answer = tokenizer.decode(generated_output[0], skip_special_tokens=True)
print(answer)

