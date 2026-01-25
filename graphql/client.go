// Package graphql provides a Go SDK for the Sui GraphQL API.
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Default endpoints for Sui networks
const (
	MainnetEndpoint = "https://graphql.mainnet.sui.io/graphql"
	TestnetEndpoint = "https://graphql.testnet.sui.io/graphql"
	DevnetEndpoint  = "https://graphql.devnet.sui.io/graphql"
)

// Client is a GraphQL client for the Sui blockchain.
type Client struct {
	endpoint   string
	httpClient *http.Client
	headers    map[string]string
	maxRetries int
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithEndpoint sets the GraphQL endpoint URL.
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithHeader adds a custom header to all requests.
func WithHeader(key, value string) ClientOption {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// WithRetries sets the maximum number of retries for failed requests.
func WithRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// NewClient creates a new Sui GraphQL client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		endpoint: MainnetEndpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers:    make(map[string]string),
		maxRetries: 3,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// graphqlRequest represents a GraphQL request payload.
type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []any                  `json:"path,omitempty"`
	Extensions map[string]any         `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents the location of a GraphQL error.
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Error implements the error interface.
func (e GraphQLError) Error() string {
	return e.Message
}

// GraphQLErrors is a collection of GraphQL errors.
type GraphQLErrors []GraphQLError

// Error implements the error interface.
func (e GraphQLErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Message
	}
	return fmt.Sprintf("%s (and %d more errors)", e[0].Message, len(e)-1)
}

// Execute sends a GraphQL query and unmarshals the response.
func (c *Client) Execute(ctx context.Context, query string, variables map[string]any, result any) error {
	return c.executeWithRetry(ctx, query, variables, result, 0)
}

// executeWithRetry executes a GraphQL query with exponential backoff retry logic.
func (c *Client) executeWithRetry(ctx context.Context, query string, variables map[string]any, result any, attempt int) error {
	reqBody := graphqlRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if attempt < c.maxRetries {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond) // Exponential backoff
			return c.executeWithRetry(ctx, query, variables, result, attempt+1)
		}
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 500 && attempt < c.maxRetries {
		time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
		return c.executeWithRetry(ctx, query, variables, result, attempt+1)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse into a temporary structure to check for errors
	var rawResp struct {
		Data   json.RawMessage `json:"data"`
		Errors []GraphQLError  `json:"errors,omitempty"`
	}

	if err := json.Unmarshal(body, &rawResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(rawResp.Errors) > 0 {
		return GraphQLErrors(rawResp.Errors)
	}

	if result != nil && len(rawResp.Data) > 0 {
		if err := json.Unmarshal(rawResp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

// Ptr returns a pointer to the given value. Useful for optional parameters.
func Ptr[T any](v T) *T {
	return &v
}
