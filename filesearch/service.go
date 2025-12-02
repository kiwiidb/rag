package filesearch

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/genai"
)

// Service provides file search operations using Gemini API
type Service struct {
	client    *genai.Client
	modelName string
}

// Config holds the configuration for the Service
type Config struct {
	APIKey    string
	ModelName string
	Backend   genai.Backend
}

// NewService creates a new file search service
func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if cfg.ModelName == "" {
		cfg.ModelName = "gemini-2.5-flash"
	}

	if cfg.Backend.String() == "" {
		cfg.Backend = genai.BackendGeminiAPI
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: cfg.Backend,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Service{
		client:    client,
		modelName: cfg.ModelName,
	}, nil
}

// Store represents a file search store
type Store struct {
	Name        string
	DisplayName string
	CreateTime  string
	UpdateTime  string
}

// Document represents a document in a store
type Document struct {
	Name           string
	DisplayName    string
	CreateTime     string
	UpdateTime     string
	CustomMetadata map[string]string
}

// CreateStore creates a new file search store
func (s *Service) CreateStore(ctx context.Context, displayName string) (*Store, error) {
	storeConfig := &genai.CreateFileSearchStoreConfig{
		DisplayName: displayName,
	}

	store, err := s.client.FileSearchStores.Create(ctx, storeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return &Store{
		Name:        store.Name,
		DisplayName: store.DisplayName,
		CreateTime:  store.CreateTime.String(),
		UpdateTime:  store.UpdateTime.String(),
	}, nil
}

// ListStores lists all file search stores
func (s *Service) ListStores(ctx context.Context) ([]*Store, error) {
	storeList, err := s.client.FileSearchStores.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list stores: %w", err)
	}

	stores := make([]*Store, 0, len(storeList.Items))
	for _, store := range storeList.Items {
		stores = append(stores, &Store{
			Name:        store.Name,
			DisplayName: store.DisplayName,
			CreateTime:  store.CreateTime.String(),
			UpdateTime:  store.UpdateTime.String(),
		})
	}

	return stores, nil
}

// GetStoreByName finds a store by its display name
func (s *Service) GetStoreByName(ctx context.Context, displayName string) (*Store, error) {
	stores, err := s.ListStores(ctx)
	if err != nil {
		return nil, err
	}

	for _, store := range stores {
		if store.DisplayName == displayName {
			return store, nil
		}
	}

	return nil, fmt.Errorf("store %q not found", displayName)
}

// ListDocuments lists all documents in a store
func (s *Service) ListDocuments(ctx context.Context, storeName string) ([]*Document, error) {
	docList, err := s.client.FileSearchStores.Documents.List(ctx, storeName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	documents := make([]*Document, 0, len(docList.Items))
	for _, doc := range docList.Items {
		// Extract custom metadata
		metadata := make(map[string]string)
		for _, cm := range doc.CustomMetadata {
			metadata[cm.Key] = cm.StringValue
		}

		documents = append(documents, &Document{
			Name:           doc.Name,
			DisplayName:    doc.DisplayName,
			CreateTime:     doc.CreateTime.String(),
			UpdateTime:     doc.UpdateTime.String(),
			CustomMetadata: metadata,
		})
	}

	return documents, nil
}

// UploadDocument uploads a document to a store using a reader
func (s *Service) UploadDocument(ctx context.Context, reader io.Reader, fileName string, storeName string) (*Document, error) {
	return s.UploadDocumentWithURL(ctx, reader, fileName, storeName, "")
}

// UploadDocumentWithURL uploads a document with an optional source URL stored in metadata
func (s *Service) UploadDocumentWithURL(ctx context.Context, reader io.Reader, fileName string, storeName string, sourceURL string) (*Document, error) {
	config := &genai.UploadToFileSearchStoreConfig{
		DisplayName: fileName,
		MIMEType:    "application/pdf",
	}

	// Add source URL as custom metadata if provided
	if sourceURL != "" {
		config.CustomMetadata = []*genai.CustomMetadata{
			{
				Key:         "source_url",
				StringValue: sourceURL,
			},
		}
	}

	_, err := s.client.FileSearchStores.UploadToFileSearchStore(ctx, reader, storeName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	return &Document{
		DisplayName: fileName,
	}, nil
}

// PromptResponse contains the response from a prompt query
type PromptResponse struct {
	Parts            []string
	Citations        []*Citation
	GroundingSupport *GroundingSupport
}

// Citation represents a citation from the file search
type Citation struct {
	StartIndex int
	EndIndex   int
	Sources    []*Source
}

// Source represents a source document
type Source struct {
	Title string
	URI   string
}

// GroundingSupport contains grounding metadata from the response
type GroundingSupport struct {
	GroundingChunks []*GroundingChunk
	WebSearchQueries []string
}

// GroundingChunk represents a chunk of content used for grounding
type GroundingChunk struct {
	Web  *WebGroundingChunk
	File *FileGroundingChunk
}

// WebGroundingChunk represents web-based grounding
type WebGroundingChunk struct {
	URI   string
	Title string
}

// FileGroundingChunk represents file-based grounding
type FileGroundingChunk struct {
	FileName string
	URI      string
}

// Prompt sends a prompt to the model with access to the specified store (without history)
func (s *Service) Prompt(ctx context.Context, prompt string, storeName string) (*PromptResponse, error) {
	tool := &genai.Tool{
		FileSearch: &genai.FileSearch{
			FileSearchStoreNames: []string{storeName},
		},
	}

	resp, err := s.client.Models.GenerateContent(ctx, s.modelName,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Tools: []*genai.Tool{tool},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return s.parseResponse(resp), nil
}

// PromptWithHistory sends a prompt to the model with conversation history and access to the specified store
func (s *Service) PromptWithHistory(ctx context.Context, prompt string, storeName string, history interface{}) (*PromptResponse, error) {
	tool := &genai.Tool{
		FileSearch: &genai.FileSearch{
			FileSearchStoreNames: []string{storeName},
		},
	}

	// Build the full prompt with conversation history
	fullPrompt := prompt
	if history != nil {
		// History is passed as []HistoryMessage from handler
		if historySlice, ok := history.([]interface{}); ok && len(historySlice) > 0 {
			contextStr := "Previous conversation:\n"
			for _, msg := range historySlice {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					role := msgMap["role"]
					content := msgMap["content"]
					contextStr += fmt.Sprintf("%s: %s\n", role, content)
				}
			}
			fullPrompt = contextStr + "\nCurrent question: " + prompt
		}
	}

	resp, err := s.client.Models.GenerateContent(ctx, s.modelName,
		genai.Text(fullPrompt),
		&genai.GenerateContentConfig{
			Tools: []*genai.Tool{tool},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return s.parseResponse(resp), nil
}

// parseResponse extracts the response data from the Gemini API response
func (s *Service) parseResponse(resp *genai.GenerateContentResponse) *PromptResponse {

	response := &PromptResponse{
		Parts:     make([]string, 0),
		Citations: make([]*Citation, 0),
	}

	for _, cand := range resp.Candidates {
		// Extract text parts
		for _, part := range cand.Content.Parts {
			if part.Text != "" {
				response.Parts = append(response.Parts, part.Text)
			}
		}

		// Extract citations from citation metadata
		if cand.CitationMetadata != nil && len(cand.CitationMetadata.Citations) > 0 {
			for _, citation := range cand.CitationMetadata.Citations {
				c := &Citation{
					StartIndex: int(citation.StartIndex),
					EndIndex:   int(citation.EndIndex),
					Sources:    make([]*Source, 0),
				}

				// Add source information if available (URI field in Citation)
				if citation.URI != "" {
					c.Sources = append(c.Sources, &Source{
						Title: citation.Title,
						URI:   citation.URI,
					})
				}

				response.Citations = append(response.Citations, c)
			}
		}

		// Extract grounding metadata
		if cand.GroundingMetadata != nil {
			response.GroundingSupport = &GroundingSupport{
				GroundingChunks:  make([]*GroundingChunk, 0),
				WebSearchQueries: cand.GroundingMetadata.WebSearchQueries,
			}

			for _, chunk := range cand.GroundingMetadata.GroundingChunks {
				gc := &GroundingChunk{}

				if chunk.Web != nil {
					gc.Web = &WebGroundingChunk{
						URI:   chunk.Web.URI,
						Title: chunk.Web.Title,
					}
				}

				if chunk.RetrievedContext != nil && chunk.RetrievedContext.URI != "" {
					gc.File = &FileGroundingChunk{
						FileName: chunk.RetrievedContext.Title,
						URI:      chunk.RetrievedContext.URI,
					}
				}

				response.GroundingSupport.GroundingChunks = append(response.GroundingSupport.GroundingChunks, gc)
			}
		}
	}

	return response
}
