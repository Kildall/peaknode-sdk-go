package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// DeviceCodeConfig configures DeviceCodeLogin. BaseURL and ClientID are
// required; the remaining fields have sensible defaults.
type DeviceCodeConfig struct {
	BaseURL     string                              // required, e.g. https://api.peaknode.ar
	ClientID    string                              // required: OAuth client ID to authenticate as
	Scopes      []string                            // space-joined for the form body
	Resource    string                              // optional RFC 8707 indicator
	OpenBrowser bool                                // when true, attempts to open the verification URI via os/exec
	OnUserCode  func(code, uri, uriComplete string) // display hook invoked once after initiate
	HTTPClient  *http.Client                        // optional override; defaults to 30s timeout
}

// Token is the successful return from DeviceCodeLogin.
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"` // seconds until expiry from issuance
	ExpiresAt    time.Time `json:"-"`          // derived; not from wire
	Scope        string    `json:"scope"`      // space-separated granted scopes
}

// ScopesSlice returns Token.Scope split on whitespace.
func (t *Token) ScopesSlice() []string {
	if t.Scope == "" {
		return nil
	}
	return strings.Fields(t.Scope)
}

// Typed errors for callers.
var (
	// ErrDeviceCodeExpired is returned when the device code expires before
	// the user approves it (RFC 8628 §3.5 expired_token, or local deadline).
	ErrDeviceCodeExpired = errors.New("device code expired before approval")
	// ErrDeviceCodeDenied is returned when the user (or org policy) denies
	// the request (RFC 8628 §3.5 access_denied).
	ErrDeviceCodeDenied = errors.New("device code authorization denied")
)

// deviceAuthResponse is the RFC 8628 §3.2 initiate response.
type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// tokenErrorResponse is the RFC 8628 §3.5 polling error response shape.
type tokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Internal sentinel errors for polling-loop control flow. Not exported —
// callers see authorization_pending + slow_down only as continued polling.
var (
	errAuthorizationPending = errors.New("authorization_pending")
	errSlowDown             = errors.New("slow_down")
)

// DeviceCodeLogin executes the full RFC 8628 OAuth Device Authorization
// Grant. Blocks until token is issued, denied, expires, or ctx is cancelled.
//
// Flow:
//  1. POST /oauth/device_authorization → device_code, user_code, verification_uri.
//  2. Invoke OnUserCode hook with the codes (display) and optionally open
//     the verification URI in the user's browser (OpenBrowser=true).
//  3. Poll POST /oauth/token at the server-supplied interval. Honor
//     slow_down by bumping the local interval by +5s per RFC 8628 §3.5.
//  4. Return a Token on success; ErrDeviceCodeExpired / ErrDeviceCodeDenied
//     on terminal errors.
func DeviceCodeLogin(ctx context.Context, cfg DeviceCodeConfig) (*Token, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("DeviceCodeLogin: BaseURL is required")
	}
	if cfg.ClientID == "" {
		return nil, errors.New("DeviceCodeLogin: ClientID is required")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	// 1) Initiate
	authResp, err := initiate(ctx, httpClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("device-code initiate: %w", err)
	}

	// 2) Notify caller + optionally open browser. Browser-open is best-effort;
	//    failure falls through silently and the user can still navigate to
	//    VerificationURI/VerificationURIComplete printed by OnUserCode.
	if cfg.OnUserCode != nil {
		cfg.OnUserCode(authResp.UserCode, authResp.VerificationURI, authResp.VerificationURIComplete)
	}
	if cfg.OpenBrowser {
		_ = openBrowser(authResp.VerificationURIComplete)
	}

	// 3) Poll /oauth/token honoring interval + slow_down.
	interval := time.Duration(authResp.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second // RFC 8628 §3.5 default
	}
	// PATTERNS §M: deadline = expires_in + 30s buffer so we always observe
	// the server's terminal expired_token rather than racing it client-side.
	deadline := time.Now().
		Add(time.Duration(authResp.ExpiresIn) * time.Second).
		Add(30 * time.Second)

	for {
		if time.Now().After(deadline) {
			return nil, ErrDeviceCodeExpired
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		tok, retryErr := pollToken(ctx, httpClient, cfg, authResp.DeviceCode)
		if retryErr == nil {
			return tok, nil
		}
		switch {
		case errors.Is(retryErr, errAuthorizationPending):
			continue
		case errors.Is(retryErr, errSlowDown):
			interval += 5 * time.Second // RFC 8628 §3.5
			continue
		case errors.Is(retryErr, ErrDeviceCodeExpired):
			return nil, ErrDeviceCodeExpired
		case errors.Is(retryErr, ErrDeviceCodeDenied):
			return nil, ErrDeviceCodeDenied
		default:
			return nil, fmt.Errorf("device-code poll: %w", retryErr)
		}
	}
}

func initiate(ctx context.Context, c *http.Client, cfg DeviceCodeConfig) (*deviceAuthResponse, error) {
	form := url.Values{
		"client_id": {cfg.ClientID},
	}
	if len(cfg.Scopes) > 0 {
		form.Set("scope", strings.Join(cfg.Scopes, " "))
	}
	if cfg.Resource != "" {
		form.Set("resource", cfg.Resource)
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		cfg.BaseURL+"/oauth/device_authorization",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var e tokenErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return nil, fmt.Errorf("device_authorization: %d %s: %s", resp.StatusCode, e.Error, e.ErrorDescription)
	}
	var out deviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func pollToken(ctx context.Context, c *http.Client, cfg DeviceCodeConfig, deviceCode string) (*Token, error) {
	form := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": {deviceCode},
		"client_id":   {cfg.ClientID},
	}
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		cfg.BaseURL+"/oauth/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var tok Token
		if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
			return nil, err
		}
		tok.ExpiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
		return &tok, nil
	}

	// Error case — read RFC 8628 §3.5 error_field
	var e tokenErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
		return nil, fmt.Errorf("oauth_token: decode error response: %w", err)
	}
	switch e.Error {
	case "authorization_pending":
		return nil, errAuthorizationPending
	case "slow_down":
		return nil, errSlowDown
	case "expired_token":
		return nil, ErrDeviceCodeExpired
	case "access_denied":
		return nil, ErrDeviceCodeDenied
	default:
		return nil, fmt.Errorf("oauth_token: %d %s: %s", resp.StatusCode, e.Error, e.ErrorDescription)
	}
}

// openBrowser launches the platform's default browser pointed at uri using
// os/exec. No new module dependency is introduced. Errors are returned but
// callers treat the failure as best-effort (the URL is also surfaced via
// the OnUserCode hook).
//
// THREAT T-EXEC-01: uri originates from the server's device_authorization
// response (a trusted JWT-issuing endpoint) and is passed as a single argv
// element to a fixed command — no shell interpolation occurs.
func openBrowser(uri string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", uri)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", uri)
	default:
		cmd = exec.Command("xdg-open", uri)
	}
	return cmd.Start()
}
