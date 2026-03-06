package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mane-pal/lazylinear/pkg/utils"
)

// OAuth application credentials — injected at build time via ldflags:
//
//	go build -ldflags "-X github.com/mane-pal/lazylinear/pkg/auth.ClientID=... -X github.com/mane-pal/lazylinear/pkg/auth.ClientSecret=..."
var (
	ClientID     = ""
	ClientSecret = ""
)

const (
	AuthorizeURL  = "https://linear.app/oauth/authorize"
	TokenURL      = "https://api.linear.app/oauth/token"
	RevokeURL     = "https://api.linear.app/oauth/revoke"
	DefaultScopes = "read,write,issues:create,comments:create"
)

// tokenResponse is the JSON response from Linear's token endpoint
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// Login performs the full OAuth2 PKCE browser flow
func Login() error {
	if ClientID == "" || ClientSecret == "" {
		return fmt.Errorf("OAuth credentials not configured — this binary was built without client ID/secret.\nRebuild with: go build -ldflags \"-X github.com/mane-pal/lazylinear/pkg/auth.ClientID=... -X github.com/mane-pal/lazylinear/pkg/auth.ClientSecret=...\"")
	}

	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return fmt.Errorf("generate PKCE verifier: %w", err)
	}
	challenge := GenerateCodeChallenge(verifier)

	state, err := GenerateCodeVerifier() // reuse for random state
	if err != nil {
		return fmt.Errorf("generate state: %w", err)
	}

	// Start local server on fixed port (must match registered redirect URI)
	const callbackPort = 54321
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", callbackPort))
	if err != nil {
		return fmt.Errorf("start local server on port %d (is something else using it?): %w", callbackPort, err)
	}
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", callbackPort)

	// Channel to receive the auth code
	type authResult struct {
		code string
		err  error
	}
	resultCh := make(chan authResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			resultCh <- authResult{err: fmt.Errorf("OAuth error: %s — %s", errParam, desc)}
			fmt.Fprint(w, "<html><body><h1>Authentication failed</h1><p>You can close this tab.</p></body></html>")
			return
		}

		if r.URL.Query().Get("state") != state {
			resultCh <- authResult{err: fmt.Errorf("state mismatch")}
			fmt.Fprint(w, "<html><body><h1>Authentication failed</h1><p>State mismatch. Please try again.</p></body></html>")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			resultCh <- authResult{err: fmt.Errorf("no authorization code received")}
			fmt.Fprint(w, "<html><body><h1>Authentication failed</h1><p>No code received.</p></body></html>")
			return
		}

		resultCh <- authResult{code: code}
		fmt.Fprint(w, "<html><body><h1>Authentication successful!</h1><p>You can close this tab and return to the terminal.</p></body></html>")
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			resultCh <- authResult{err: fmt.Errorf("local server: %w", err)}
		}
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Build authorization URL
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256&prompt=consent",
		AuthorizeURL,
		url.QueryEscape(ClientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(DefaultScopes),
		url.QueryEscape(state),
		url.QueryEscape(challenge),
	)

	fmt.Println("Authenticating with Linear...")
	fmt.Println("Opening browser for authentication...")
	fmt.Printf("If the browser doesn't open, visit:\n%s\n\n", authURL)

	if err := utils.OpenBrowser(authURL); err != nil {
		// Non-fatal: user can copy the URL manually
		fmt.Printf("Could not open browser: %v\n", err)
	}

	fmt.Println("Waiting for authentication...")

	// Wait for callback with timeout
	select {
	case result := <-resultCh:
		if result.err != nil {
			return result.err
		}
		// Exchange code for tokens
		token, err := exchangeCode(result.code, verifier, redirectURI)
		if err != nil {
			return err
		}
		if err := SaveToken(token); err != nil {
			return err
		}
		fmt.Printf("\nAuthentication successful!\n")
		fmt.Printf("  Scopes: %s\n", token.Scope)
		fmt.Printf("  Token expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		return nil

	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timed out (5 minutes)")
	}
}

// exchangeCode exchanges an authorization code for tokens
func exchangeCode(code, verifier, redirectURI string) (*Token, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {ClientID},
		"client_secret": {ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}

	resp, err := http.Post(TokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("token exchange failed (%d): %s — %s", resp.StatusCode, errResp.Error, errResp.Description)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	return &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
		TokenType:    tokenResp.TokenType,
	}, nil
}

// RefreshAccessToken refreshes an expired OAuth token
func RefreshAccessToken(token *Token) (*Token, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {ClientID},
		"client_secret": {ClientSecret},
		"refresh_token": {token.RefreshToken},
	}

	resp, err := http.Post(TokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed (%d): token may have been revoked, run `lazylinear auth login`", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}

	newToken := &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
		TokenType:    tokenResp.TokenType,
	}

	// Keep old refresh token if new one wasn't provided
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = token.RefreshToken
	}

	if err := SaveToken(newToken); err != nil {
		return nil, fmt.Errorf("save refreshed token: %w", err)
	}
	return newToken, nil
}

// Logout revokes the OAuth token and removes the credentials file
func Logout() error {
	token, err := LoadToken()
	if err != nil {
		// No token saved, nothing to do
		fmt.Println("Not currently authenticated via OAuth.")
		return nil
	}

	// Attempt to revoke at Linear (best-effort)
	data := url.Values{
		"client_id":     {ClientID},
		"client_secret": {ClientSecret},
		"token":         {token.AccessToken},
	}
	resp, err := http.Post(RevokeURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err == nil {
		resp.Body.Close()
	}

	if err := DeleteToken(); err != nil {
		return err
	}

	fmt.Println("Logged out successfully.")
	return nil
}

// Status prints the current authentication status
func Status() {
	token, err := LoadToken()
	if err == nil {
		fmt.Println("Authentication: OAuth (browser login)")
		fmt.Printf("  Scopes: %s\n", token.Scope)
		if token.IsExpired() {
			fmt.Println("  Status: expired (will auto-refresh)")
		} else {
			fmt.Printf("  Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		return
	}

	fmt.Println("Not authenticated.")
	fmt.Println("\nTo authenticate, run:")
	fmt.Println("  lazylinear auth login")
}
