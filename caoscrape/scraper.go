package caoscrape

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL       = "https://public-search.werk.belgie.be"
	apiPrefix     = "/website-service/joint-work-convention"
	searchURL     = baseURL + apiPrefix + "/search"
	documentBase  = baseURL + apiPrefix
)

// SearchRequest represents the search parameters
type SearchRequest struct {
	Lang                      string       `json:"lang"`
	JC                        *int         `json:"jc"`
	Title                     *string      `json:"title"`
	SuperTheme                string       `json:"superTheme"`
	Theme                     *string      `json:"theme"`
	TextSearchTerms           *string      `json:"textSearchTerms"`
	SignatureDate             DateRange    `json:"signatureDate"`
	DepositNumber             NumberRange  `json:"depositNumber"`
	NoticeMBDepositDate       DateRange    `json:"noticeDepositMBDate"`
	Enforced                  *bool        `json:"enforced"`
	RoyalDecreeDate           DateRange    `json:"royalDecreeDate"`
	PublicationRoyalDecreeDate DateRange   `json:"publicationRoyalDecreeDate"`
	RecordDate                DateRange    `json:"recordDate"`
	CorrectedDate             DateRange    `json:"correctedDate"`
	DepositDate               DateRange    `json:"depositDate"`
	AdvancedSearch            bool         `json:"advancedSearch"`
}

// DateRange represents a date range with start and end
type DateRange struct {
	Start *string `json:"start"`
	End   *string `json:"end"`
}

// NumberRange represents a number range with start and end
type NumberRange struct {
	Start *int `json:"start"`
	End   *int `json:"end"`
}

// SearchResult represents a single search result
type SearchResult struct {
	DocumentLink string `json:"documentLink"`
	// Add other fields as needed
}

// Client handles requests to the CAO search API
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new CAO scraper client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

// Search searches for documents by JC number
// jc parameter can be nil to search all documents
func (c *Client) Search(jc *int) ([]string, error) {
	// Build search request
	req := SearchRequest{
		Lang:           "nl",
		JC:             jc,
		Title:          nil,
		SuperTheme:     "",
		Theme:          nil,
		TextSearchTerms: nil,
		SignatureDate: DateRange{
			Start: nil,
			End:   nil,
		},
		DepositNumber: NumberRange{
			Start: nil,
			End:   nil,
		},
		NoticeMBDepositDate: DateRange{
			Start: nil,
			End:   nil,
		},
		Enforced: nil,
		RoyalDecreeDate: DateRange{
			Start: nil,
			End:   nil,
		},
		PublicationRoyalDecreeDate: DateRange{
			Start: nil,
			End:   nil,
		},
		RecordDate: DateRange{
			Start: nil,
			End:   nil,
		},
		CorrectedDate: DateRange{
			Start: nil,
			End:   nil,
		},
		DepositDate: DateRange{
			Start: nil,
			End:   nil,
		},
		AdvancedSearch: false,
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", searchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var results []SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract and prefix document links
	documentURLs := make([]string, 0, len(results))
	for _, result := range results {
		if result.DocumentLink != "" {
			// Prefix with base URL and API prefix
			// Add leading slash if the documentLink doesn't start with one
			link := result.DocumentLink
			if link[0] != '/' {
				link = "/" + link
			}
			fullURL := documentBase + link
			documentURLs = append(documentURLs, fullURL)
		}
	}

	return documentURLs, nil
}

// DownloadDocument downloads a document from the given URL and returns it as an io.Reader
func (c *Client) DownloadDocument(url string) (io.Reader, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the entire response body into memory
	// This allows us to close the response body and return a reader
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return bytes.NewReader(data), nil
}
