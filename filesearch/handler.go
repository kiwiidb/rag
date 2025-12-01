package filesearch

import (
	"encoding/json"
	"net/http"
)

// HistoryMessage represents a single message in the conversation history
type HistoryMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // The message content
}

// QueryRequest represents the incoming query request
type QueryRequest struct {
	Query     string            `json:"query"`
	StoreName string            `json:"storeName"`
	History   []HistoryMessage  `json:"history,omitempty"` // Optional conversation history
}

// SourceDocument represents a source document with its URI
type SourceDocument struct {
	FileName string `json:"fileName"`
	URI      string `json:"uri"`
}

// QueryResponse represents the response to a query
type QueryResponse struct {
	Answer           string            `json:"answer"`
	Sources          []*SourceDocument `json:"sources"`
	Citations        []*Citation       `json:"citations,omitempty"`
	GroundingSupport *GroundingSupport `json:"groundingSupport,omitempty"`
	Error            string            `json:"error,omitempty"`
}

// Handler provides HTTP handlers for the file search service
type Handler struct {
	service *Service
}

// NewHandler creates a new HTTP handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Query handles POST requests to query documents
// POST /query
// Body: {"query": "your question", "storeName": "store-name"}
func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(QueryResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate request
	if req.Query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(QueryResponse{
			Error: "Query is required",
		})
		return
	}

	if req.StoreName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(QueryResponse{
			Error: "StoreName is required",
		})
		return
	}

	// Get the store by display name to get the actual store name
	store, err := h.service.GetStoreByName(r.Context(), req.StoreName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(QueryResponse{
			Error: "Store not found: " + err.Error(),
		})
		return
	}

	// Execute query with the actual store name (not display name) and conversation history
	resp, err := h.service.PromptWithHistory(r.Context(), req.Query, store.Name, req.History)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(QueryResponse{
			Error: "Failed to execute query: " + err.Error(),
		})
		return
	}

	// Build response
	response := QueryResponse{
		Citations:        resp.Citations,
		GroundingSupport: resp.GroundingSupport,
	}

	// Combine answer parts
	for _, part := range resp.Parts {
		response.Answer += part
	}

	// Extract unique source file names with URIs
	seenSources := make(map[string]bool)
	if resp.GroundingSupport != nil {
		for _, chunk := range resp.GroundingSupport.GroundingChunks {
			if chunk.File != nil && !seenSources[chunk.File.FileName] {
				seenSources[chunk.File.FileName] = true
				response.Sources = append(response.Sources, &SourceDocument{
					FileName: chunk.File.FileName,
					URI:      chunk.File.URI,
				})
			}
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListStoresHandler handles GET requests to list all stores
// GET /stores
func (h *Handler) ListStoresHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stores, err := h.service.ListStores(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to list stores: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stores)
}

// ListDocumentsHandler handles GET requests to list documents in a store
// GET /stores/{storeName}/documents
func (h *Handler) ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	storeName := r.URL.Query().Get("storeName")
	if storeName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "storeName query parameter is required",
		})
		return
	}

	docs, err := h.service.ListDocuments(r.Context(), storeName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to list documents: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}
