package webhook

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ChallengePayload struct {
	Challenge string `json:"challenge"`
}

// PerformChallengeHandshake generates a secure token, sends it to the targetURL,
// and verifies that the endpoint echoes the token back in the response body with a 200 OK.
func PerformChallengeHandshake(ctx context.Context, client *http.Client, targetURL string) error {
	// 1. Generate a cryptographic hex challenge token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate challenge token: %w", err)
	}
	challenge := hex.EncodeToString(tokenBytes)

	// 2. Prepare the outbound POST request
	payload := ChallengePayload{Challenge: challenge}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal challenge payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Email-Ingestion-Gateway/1.0 (Webhook-Challenge)")

	// 3. Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook challenge request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook challenge rejected: expected status 200, got %d", resp.StatusCode)
	}

	// 4. Verify the response echoes the exact challenge
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*10)) // limit to 10KB
	if err != nil {
		return fmt.Errorf("failed to read challenge response: %w", err)
	}

	// Try to decode response as JSON first
	var echoPayload ChallengePayload
	if err := json.Unmarshal(respBody, &echoPayload); err == nil && echoPayload.Challenge != "" {
		if echoPayload.Challenge != challenge {
			return fmt.Errorf("webhook challenge failed: echoed JSON token did not match")
		}
		return nil
	}

	// Fallback to checking if the raw string contains the token
	if !bytes.Contains(respBody, []byte(challenge)) {
		return fmt.Errorf("webhook challenge failed: expected token not found in response")
	}

	return nil
}
