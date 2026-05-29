// Package auth provides hand-written helpers that drive multi-step OAuth
// flows the generated client cannot orchestrate. Currently exposes:
//
//	DeviceCodeLogin — RFC 8628 OAuth 2.0 Device Authorization Grant.
//
// The helpers do NOT replace the generated PeaknodeClient — once a Token
// is obtained, callers construct a PeaknodeClient with Bearer auth via
// the generated SDK as usual.
package auth
