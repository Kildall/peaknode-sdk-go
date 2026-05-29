package sdk

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SSEEvent represents a parsed Server-Sent Event. Exactly one field is non-nil.
type SSEEvent struct {
	Progress *ProgressEvent
	Done     *DoneEvent
	Error    error
}

// ProgressEvent matches the server's ScanProgressEvent format.
// Uses string for JobID (JSON wire format) -- caller can parse to uuid.UUID if needed.
type ProgressEvent struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

// DoneEvent matches the server's ScanCompletedEvent format.
type DoneEvent struct {
	JobID  string          `json:"job_id"`
	Status string          `json:"status"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// StreamProgress connects to the SSE endpoint for the given job and returns a
// channel of typed events. The channel closes when the stream ends (done event),
// context is cancelled, or an error occurs. Safe for use in both Bubble Tea
// (goroutine ranging over channel with p.Send) and batch mode (direct range).
func (c *PeaknodeClient) StreamProgress(ctx context.Context, jobID string) <-chan SSEEvent {
	ch := make(chan SSEEvent)
	go func() {
		defer close(ch)

		url := fmt.Sprintf("%s/api/jobs/%s/stream", strings.TrimRight(c.config.BaseURL, "/"), jobID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			ch <- SSEEvent{Error: fmt.Errorf("creating SSE request: %w", err)}
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Token))
		req.Header.Set("Accept", "text/event-stream")

		resp, err := c.http.Do(req)
		if err != nil {
			ch <- SSEEvent{Error: fmt.Errorf("connecting to SSE stream: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ch <- SSEEvent{Error: fmt.Errorf("SSE stream returned status %d", resp.StatusCode)}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		var eventType string
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event: "):
				eventType = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				data := strings.TrimPrefix(line, "data: ")
				switch eventType {
				case "progress":
					var evt ProgressEvent
					if err := json.Unmarshal([]byte(data), &evt); err != nil {
						ch <- SSEEvent{Error: fmt.Errorf("parsing progress event: %w", err)}
						return
					}
					ch <- SSEEvent{Progress: &evt}
				case "done":
					var evt DoneEvent
					if err := json.Unmarshal([]byte(data), &evt); err != nil {
						ch <- SSEEvent{Error: fmt.Errorf("parsing done event: %w", err)}
						return
					}
					ch <- SSEEvent{Done: &evt}
					return // Stream complete
				}
				eventType = "" // Reset after processing data line
			}
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			ch <- SSEEvent{Error: fmt.Errorf("reading SSE stream: %w", err)}
		}
	}()
	return ch
}
