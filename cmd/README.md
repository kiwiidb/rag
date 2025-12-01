# CAO Command-Line Tools

This directory contains command-line tools for working with Belgian CAO (Collectieve Arbeidsovereenkomst) documents using Gemini's File Search capabilities.

## Tools

### cao-uploader

Searches for and uploads CAO documents to a Gemini File Search Store.

**Usage:**
```bash
go run cmd/cao-uploader/main.go
```

Or build and run:
```bash
go build -o cao-uploader cmd/cao-uploader/main.go
./cao-uploader
```

**What it does:**
1. Creates or retrieves a File Search Store named "cao-documents"
2. Searches for documents with JC number 3180200 (configurable in code)
3. Downloads documents from the Belgian CAO public search portal
4. Uploads documents to the File Search Store (idempotent - skips already uploaded files)
5. Reports upload statistics

**Environment Variables:**
- `GEMINI_API_KEY` - Required. Your Gemini API key

**Customization:**
To search for different JC numbers, modify line 53 in the code:
```go
jc := 3180200  // Change this to your desired JC number
```

---

### cao-querier

Queries the uploaded CAO documents using natural language.

**Usage:**
```bash
go run cmd/cao-querier/main.go "your question here"
```

Or build and run:
```bash
go build -o cao-querier cmd/cao-querier/main.go
./cao-querier "Wat is het minimumloon als je 17 jaar bent?"
```

**What it does:**
1. Connects to the "cao-documents" File Search Store
2. Sends your query to Gemini with access to the uploaded documents
3. Returns an answer grounded in the documents
4. Shows which documents were used as sources
5. Displays citations if available

**Environment Variables:**
- `GEMINI_API_KEY` - Required. Your Gemini API key

**Examples:**
```bash
# Query about minimum wage
./cao-querier "Wat is het minimumloon als je 17 jaar bent?"

# Query about vacation days
./cao-querier "Hoeveel vakantiedagen heb je recht op?"

# Query about working hours
./cao-querier "What are the maximum working hours per week?"
```

---

## Quick Start

1. **Set your API key:**
   ```bash
   export GEMINI_API_KEY="your-api-key-here"
   ```

2. **Upload documents:**
   ```bash
   go run cmd/cao-uploader/main.go
   ```

3. **Query documents:**
   ```bash
   go run cmd/cao-querier/main.go "Wat is het minimumloon als je 17 jaar bent?"
   ```

---

## Building Binaries

Build both tools:
```bash
go build -o cao-uploader cmd/cao-uploader/main.go
go build -o cao-querier cmd/cao-querier/main.go
```

Or build all at once:
```bash
go build -o cao-uploader ./cmd/cao-uploader
go build -o cao-querier ./cmd/cao-querier
```

---

## Store Configuration

Both tools use a store named "cao-documents". To use a different store name, modify the `storeName` constant in both files:

**cao-uploader/main.go (line 33):**
```go
storeName := "cao-documents"  // Change this
```

**cao-querier/main.go (line 44):**
```go
storeName := "cao-documents"  // Change this
```

---

## Troubleshooting

**"GEMINI_API_KEY environment variable not set"**
- Set the environment variable before running the tools

**"Store 'cao-documents' not found"**
- Run `cao-uploader` first to create the store and upload documents

**"Failed to search" or "Failed to download"**
- Check your internet connection
- Verify the Belgian CAO search portal is accessible
- The JC number might not exist or have no documents

**"Failed to upload document"**
- Check your API key is valid
- Verify you haven't exceeded API quotas
- Ensure the document format is supported by File Search
