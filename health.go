package sdk

import (
	"context"
	"fmt"
	"net/http"
)

// HealthCheck performs a lightweight GET /health request to verify server
// reachability.
func (c *PeaknodeClient) HealthCheck(ctx context.Context) error {
	url := c.config.BaseURL + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating health request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Token))

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("health check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	return nil
}

// BaseURL returns the configured server base URL.
func (c *PeaknodeClient) BaseURL() string {
	return c.config.BaseURL
}
