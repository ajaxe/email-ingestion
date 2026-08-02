package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ajaxe/email-ingestion/internal/ingest"
)

type IngestClient struct {
	httpClient *http.Client
	baseURL    string
	edgeToken  string
}

func NewIngestClient(baseURL, edgeToken string) *IngestClient {
	return &IngestClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
		edgeToken:  edgeToken,
	}
}

func (c *IngestClient) StreamPayload(ctx context.Context, reader io.Reader) error {
	reqURL := fmt.Sprintf("%s%s", c.baseURL, ingest.IngestEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, reader)
	if err != nil {
		return fmt.Errorf("failed to create ingest request: %w", err)
	}

	req.Header.Set(ingest.HeaderEdgeToken, c.edgeToken)
	req.Header.Set("Content-Type", "message/rfc822")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send ingest request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ingest API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *IngestClient) ValidateAddress(ctx context.Context, address string) (bool, error) {
	reqURL := fmt.Sprintf("%s/internal/api/v1/addresses/%s", c.baseURL, address)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create address validation request: %w", err)
	}

	req.Header.Set(ingest.HeaderEdgeToken, c.edgeToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send address validation request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}

	return false, fmt.Errorf("address validation returned unexpected status %d", resp.StatusCode)
}
