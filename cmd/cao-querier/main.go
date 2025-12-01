package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"rag/filesearch"
	"strings"

	"google.golang.org/genai"
)

func main() {
	// Check if query is provided
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s \"your question here\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s \"Wat is het minimumloon als je 17 jaar bent?\"\n", os.Args[0])
		os.Exit(1)
	}

	// Get query from command line arguments (join all args in case user didn't quote)
	query := strings.Join(os.Args[1:], " ")

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

	// Get the store
	storeName := "cao-documents"
	store, err := service.GetStoreByName(ctx, storeName)
	if err != nil {
		log.Fatalf("Store '%s' not found. Please run cao-uploader first to upload documents.\n", storeName)
	}

	// Query the documents
	fmt.Printf("Querying: %s\n\n", query)

	resp, err := service.Prompt(ctx, query, store.Name)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	// Print response
	fmt.Println("=== Answer ===")
	for _, part := range resp.Parts {
		fmt.Println(part)
	}

	// Print grounding metadata
	if resp.GroundingSupport != nil && len(resp.GroundingSupport.GroundingChunks) > 0 {
		fmt.Printf("\n=== Sources (%d) ===\n", len(resp.GroundingSupport.GroundingChunks))

		// Deduplicate sources by filename
		seenFiles := make(map[string]bool)
		sourceCount := 0

		for _, chunk := range resp.GroundingSupport.GroundingChunks {
			if chunk.File != nil && !seenFiles[chunk.File.FileName] {
				sourceCount++
				seenFiles[chunk.File.FileName] = true
				fmt.Printf("%d. %s\n", sourceCount, chunk.File.FileName)
			}
		}
	}

	// Print citations if available
	if len(resp.Citations) > 0 {
		fmt.Printf("\n=== Citations (%d) ===\n", len(resp.Citations))
		for i, citation := range resp.Citations {
			fmt.Printf("%d. Characters %d-%d", i+1, citation.StartIndex, citation.EndIndex)
			if len(citation.Sources) > 0 {
				fmt.Printf(" - %s", citation.Sources[0].Title)
			}
			fmt.Println()
		}
	}
}
