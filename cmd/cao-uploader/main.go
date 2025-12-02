package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"rag/caoscrape"
	"rag/filesearch"

	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	// Create the file search service
	service, err := filesearch.NewService(ctx, &filesearch.Config{
		APIKey:    apiKey,
		ModelName: "gemini-2.5-flash",
		Backend:   genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Get or create store
	storeName := "cao-documents"
	var store *filesearch.Store

	fmt.Println("Checking if File Search Store exists...")
	store, err = service.GetStoreByName(ctx, storeName)
	if err != nil {
		fmt.Println("Creating File Search Store...")
		store, err = service.CreateStore(ctx, storeName)
		if err != nil {
			log.Fatalf("Failed to create store: %v", err)
		}
		fmt.Printf("Store created: %s\n", store.DisplayName)
	} else {
		fmt.Printf("Store already exists: %s\n", store.DisplayName)
	}

	// Create CAO scraper client
	scraper := caoscrape.NewClient()

	// Search for documents with specific JC number
	jc := 3180200
	fmt.Printf("\nSearching for documents with JC number %d...\n", jc)
	urls, err := scraper.Search(&jc)
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	fmt.Printf("Found %d documents\n", len(urls))

	// Get existing documents to avoid re-uploading
	existingDocs, err := service.ListDocuments(ctx, store.Name)
	if err != nil {
		log.Printf("Warning: Failed to list existing documents: %v", err)
		existingDocs = []*filesearch.Document{}
	}

	existingFiles := make(map[string]bool)
	for _, doc := range existingDocs {
		existingFiles[doc.DisplayName] = true
	}

	// Download and upload each document
	uploadedCount := 0
	skippedCount := 0

	for i, url := range urls {
		// Extract filename from URL
		fileName := filepath.Base(url)
		if fileName == "" || fileName == "." {
			fileName = fmt.Sprintf("document_%d.pdf", i+1)
		}

		// Check if already uploaded
		if existingFiles[fileName] {
			fmt.Printf("Skipping %s (already uploaded)\n", fileName)
			skippedCount++
			continue
		}

		fmt.Printf("Downloading and uploading %s...\n", fileName)

		// Download document
		reader, err := scraper.DownloadDocument(url)
		if err != nil {
			log.Printf("Warning: Failed to download %s: %v", url, err)
			continue
		}

		// Upload to file search store with source URL
		_, err = service.UploadDocumentWithURL(ctx, reader, fileName, store.Name, url)
		if err != nil {
			log.Printf("Warning: Failed to upload %s: %v", fileName, err)
			continue
		}

		uploadedCount++
	}

	fmt.Printf("\nUpload complete: %d new documents uploaded, %d documents skipped\n", uploadedCount, skippedCount)
	fmt.Printf("\nUse 'cao-querier \"your question\"' to query the uploaded documents\n")
}
