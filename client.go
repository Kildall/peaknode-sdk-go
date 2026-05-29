package sdk

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kildall/peaknode-sdk-go/gen"
)

// Config holds client configuration. Maps to Cobra CLI flags:
// --server -> BaseURL, --token -> Token.
type Config struct {
	BaseURL    string        // Required: API server URL (e.g., "http://localhost:8080")
	Token      string        // Required: Bearer token for authentication
	HTTPClient *http.Client  // Optional: defaults to &http.Client{Timeout: Timeout}
	Timeout    time.Duration // Optional: defaults to 30s when HTTPClient is nil
}

// PeaknodeClient wraps the generated oapi-codegen client with Bearer auth injection.
type PeaknodeClient struct {
	*gen.ClientWithResponses
	config Config
	http   *http.Client
}

// NewClient creates a PeaknodeClient that injects Bearer token auth on every request.
func NewClient(cfg Config) (*PeaknodeClient, error) {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	bearerAuth := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Token))
		return nil
	}

	client, err := gen.NewClientWithResponses(
		cfg.BaseURL,
		gen.WithHTTPClient(httpClient),
		gen.WithRequestEditorFn(bearerAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("creating API client: %w", err)
	}

	return &PeaknodeClient{
		ClientWithResponses: client,
		config:              cfg,
		http:                httpClient,
	}, nil
}
