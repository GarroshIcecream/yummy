// Package chatgpt implements "Sign in with ChatGPT": an OAuth 2.0
// authorization-code flow with PKCE against OpenAI's auth server, so a
// ChatGPT Plus/Pro subscription can be used instead of an API key.
//
// This is the same public-client flow the OpenAI Codex CLI uses. The
// endpoints are not a documented public API and may change.
package chatgpt

const (
	// ClientID is the public OAuth client registered for the Codex CLI
	// desktop flow.
	ClientID = "app_EMoamEEZ73f0CkXaXp7hrann"

	// Issuer is OpenAI's authorization server.
	Issuer = "https://auth.openai.com"

	// AuthorizeURL and TokenURL are the OAuth endpoints on Issuer.
	AuthorizeURL = Issuer + "/oauth/authorize"
	TokenURL     = Issuer + "/oauth/token"

	// Scope requests identity claims plus a refresh token.
	Scope = "openid profile email offline_access"

	// CodexBaseURL is the ChatGPT backend that serves Responses API
	// requests for ChatGPT-plan accounts.
	CodexBaseURL = "https://chatgpt.com/backend-api/codex"

	// Originator identifies this client to OpenAI.
	Originator = "yummy"

	// callbackPath is the redirect path registered for the client.
	callbackPath = "/auth/callback"

	// authClaim is the JWT claim namespace that carries the account ID.
	authClaim = "https://api.openai.com/auth"
)

// DefaultCallbackPorts are the loopback ports allowed as redirect targets,
// tried in order.
var DefaultCallbackPorts = []int{1455, 1457}
