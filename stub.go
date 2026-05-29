package sdk

import (
	"context"
	"encoding/json"
	"fmt"
)

// --- Ownership API stubs ---

// CreateOwnershipProof creates a new ownership proof challenge.
// Real client: POST /api/ownership
func (c *PeaknodeClient) CreateOwnershipProof(ctx context.Context, blockchain, address string) (json.RawMessage, error) {
	// TODO(phase-28): Replace with generated client call
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// SubmitOwnershipSignature submits a signature for verification.
// Real client: POST /api/ownership/:id/verification
func (c *PeaknodeClient) SubmitOwnershipSignature(ctx context.Context, proofID, signature string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// GetOwnershipProof retrieves an ownership proof by ID.
// Real client: GET /api/ownership/:id
func (c *PeaknodeClient) GetOwnershipProof(ctx context.Context, proofID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// --- Funds API stubs ---

// Note: ProgressEvent and DoneEvent types are defined in sse.go.
// StreamProgress is also defined in sse.go.

// CreateWallet registers a wallet for the organization.
// Real client: POST /api/wallets
func (c *PeaknodeClient) CreateWallet(ctx context.Context, name, chainID, address string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// CreateProofOfFunds creates a new proof of funds.
// Real client: POST /api/funds
func (c *PeaknodeClient) CreateProofOfFunds(ctx context.Context, name, description string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// StartScan initiates a balance scan for a proof of funds.
// Real client: POST /api/funds/:id/scan
func (c *PeaknodeClient) StartScan(ctx context.Context, proofOfFundsID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// GetScanJob retrieves a scan job by ID (for polling fallback).
// Real client: GET /api/jobs/:id
func (c *PeaknodeClient) GetScanJob(ctx context.Context, jobID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// GetFundsAssets retrieves scanned assets for a proof of funds.
// Real client: GET /api/funds/:id/assets
func (c *PeaknodeClient) GetFundsAssets(ctx context.Context, proofOfFundsID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// --- Liabilities methods ---

// CreateProofOfLiabilities creates a new proof of liabilities.
// Real client: POST /api/liabilities
func (c *PeaknodeClient) CreateProofOfLiabilities(ctx context.Context, name, proofMode string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// UploadLiabilityData uploads liability data CSV file.
// Real client: POST /api/liabilities/:id/data
func (c *PeaknodeClient) UploadLiabilityData(ctx context.Context, proofID, filePath string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// ProcessLiabilities starts processing and returns the SSE job ID for streaming.
// Real client: POST /api/liabilities/:id/processing
func (c *PeaknodeClient) ProcessLiabilities(ctx context.Context, proofID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// GetProofOfLiabilities retrieves a proof of liabilities by ID.
// Real client: GET /api/liabilities/:id
func (c *PeaknodeClient) GetProofOfLiabilities(ctx context.Context, proofID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// --- Verify / Merkle methods ---

// GetInclusionProof retrieves a Merkle inclusion proof for a user.
// Real client: GET /verify/liabilities/:proofId/inclusion?user=X&nonce=Y
func (c *PeaknodeClient) GetInclusionProof(ctx context.Context, proofID, userIdentifier, nonce string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// CheckProof verifies a Merkle or ZKP proof.
// Real client: POST /verify/liabilities/:proofId/check
func (c *PeaknodeClient) CheckProof(ctx context.Context, proofID string, proofData json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// --- ZKP methods ---

// GetZKPArtifact retrieves the ZKP artifact for a proof of liabilities.
// Real client: GET /api/liabilities/:id/zkp-artifact
func (c *PeaknodeClient) GetZKPArtifact(ctx context.Context, proofOfLiabilitiesID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// --- Reports methods ---

// GenerateReport creates a new report (returns 202 Accepted).
// Real client: POST /api/reports
func (c *PeaknodeClient) GenerateReport(ctx context.Context, format, name string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// GetReport polls report status.
// Real client: GET /api/reports/:id
func (c *PeaknodeClient) GetReport(ctx context.Context, reportID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}

// GetReportDownloadURL retrieves a presigned download URL.
// Real client: GET /api/reports/:id/download
func (c *PeaknodeClient) GetReportDownloadURL(ctx context.Context, reportID string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented: connect to a running server")
}
