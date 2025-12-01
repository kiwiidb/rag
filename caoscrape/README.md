# CAO Scraper Package

This package provides functionality to search and download documents from the Belgian CAO (Collectieve Arbeidsovereenkomst / Convention Collective de Travail) public search portal.

## Features

- **Search**: Search for CAO documents by JC (Joint Committee) number
- **Download**: Download documents as `io.Reader` for further processing

## Usage

### Basic Search

```go
import "rag/caoscrape"

client := caoscrape.NewClient()

// Search by JC number
jc := 3180200
urls, err := client.Search(&jc)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d documents\n", len(urls))
for _, url := range urls {
    fmt.Println(url)
}
```

### Search All Documents

```go
// Pass nil to search all documents
urls, err := client.Search(nil)
```

### Download Document

```go
// Download a document
reader, err := client.DownloadDocument(url)
if err != nil {
    log.Fatal(err)
}

// Use the reader (e.g., upload to file search)
service.UploadDocument(ctx, reader, "document.pdf", storeName)
```

## API Reference

### Client

```go
type Client struct {
    // internal fields
}

func NewClient() *Client
```

### Methods

#### Search

```go
func (c *Client) Search(jc *int) ([]string, error)
```

Searches for documents by JC number. Pass `nil` to search all documents.

**Parameters:**
- `jc`: Joint Committee number (optional, can be nil)

**Returns:**
- List of full document URLs
- Error if the request fails

#### DownloadDocument

```go
func (c *Client) DownloadDocument(url string) (io.Reader, error)
```

Downloads a document from the given URL and returns it as an `io.Reader`.

**Parameters:**
- `url`: Full URL of the document to download

**Returns:**
- `io.Reader` containing the document data
- Error if the download fails

## Integration Example

See [cmd/cao-uploader/main.go](../cmd/cao-uploader/main.go) for a complete example that:
1. Searches for CAO documents
2. Downloads them
3. Uploads them to a Gemini File Search Store
4. Queries the documents with natural language

Run it with:
```bash
go run cmd/cao-uploader/main.go
```
