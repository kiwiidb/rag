package filesearch_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"rag/filesearch"

	"google.golang.org/genai"
)

func Example() {
	ctx := context.Background()

	// Create service
	service, err := filesearch.NewService(ctx, &filesearch.Config{
		APIKey:    os.Getenv("GEMINI_API_KEY"),
		ModelName: "gemini-2.5-flash",
		Backend:   genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	// List existing stores
	stores, err := service.ListStores(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d stores\n", len(stores))

	// Create or get store
	storeName := "my-documents"
	store, err := service.GetStoreByName(ctx, storeName)
	if err != nil {
		// Store doesn't exist, create it
		store, err = service.CreateStore(ctx, storeName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created store: %s\n", store.DisplayName)
	} else {
		fmt.Printf("Using existing store: %s\n", store.DisplayName)
	}

	// List documents in store
	docs, err := service.ListDocuments(ctx, store.Name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Store has %d documents\n", len(docs))

	// Upload a document from a reader
	content := "This is a sample document content."
	reader := strings.NewReader(content)
	doc, err := service.UploadDocument(ctx, reader, "sample.txt", store.Name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Uploaded document: %s\n", doc.DisplayName)

	// Query the store
	resp, err := service.Prompt(ctx, "What information is in the documents?", store.Name)
	if err != nil {
		log.Fatal(err)
	}

	// Print response parts
	for i, part := range resp.Parts {
		fmt.Printf("Response part %d: %s\n", i+1, part)
	}

	// Print citations
	for i, citation := range resp.Citations {
		fmt.Printf("Citation %d (chars %d-%d):\n", i+1, citation.StartIndex, citation.EndIndex)
		for j, source := range citation.Sources {
			fmt.Printf("  Source %d: %s (%s)\n", j+1, source.Title, source.URI)
		}
	}
}
