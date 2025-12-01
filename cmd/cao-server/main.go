package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"rag/filesearch"

	"google.golang.org/genai"
)

func main() {
	// Get configuration from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create the file search service
	ctx := context.Background()
	service, err := filesearch.NewService(ctx, &filesearch.Config{
		APIKey:    apiKey,
		ModelName: "gemini-2.5-flash",
		Backend:   genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create handler
	handler := filesearch.NewHandler(service)

	// Register routes
	http.HandleFunc("/query", handler.Query)
	http.HandleFunc("/stores", handler.ListStoresHandler)
	http.HandleFunc("/documents", handler.ListDocumentsHandler)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Serve interactive chat interface at root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "cmd/cao-server/templates/chat.html")
	})

	// Start server
	addr := ":" + port
	log.Printf("Starting CAO Query Server on %s", addr)
	log.Printf("Visit http://localhost%s for the chat interface", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
