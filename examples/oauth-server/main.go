// Example OAuth server demonstrating SmartThings OAuth 2.0 flow
//
// Usage:
//   export SMARTTHINGS_CLIENT_ID=your-client-id
//   export SMARTTHINGS_CLIENT_SECRET=your-client-secret
//   export SMARTTHINGS_REDIRECT_URL=http://localhost:8080/callback
//   go run main.go
//
// Then visit http://localhost:8080 to start the OAuth flow.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	st "github.com/tj-smith47/smartthings-go"
)

var (
	client      *st.OAuthClient
	stateStore  = make(map[string]bool)
	stateMu     sync.RWMutex
)

func main() {
	// Load configuration from environment
	clientID := os.Getenv("SMARTTHINGS_CLIENT_ID")
	clientSecret := os.Getenv("SMARTTHINGS_CLIENT_SECRET")
	redirectURL := os.Getenv("SMARTTHINGS_REDIRECT_URL")

	if clientID == "" || clientSecret == "" {
		log.Fatal("SMARTTHINGS_CLIENT_ID and SMARTTHINGS_CLIENT_SECRET are required")
	}
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/callback"
	}

	// Configure OAuth
	config := &st.OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       st.DefaultScopes(),
	}

	// Create token store (file-based for persistence)
	store := st.NewFileTokenStore("tokens.json")

	// Create OAuth client
	var err error
	client, err = st.NewOAuthClient(config, store)
	if err != nil {
		log.Fatalf("Failed to create OAuth client: %v", err)
	}

	// Set up routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/devices", handleDevices)
	http.HandleFunc("/logout", handleLogout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on http://localhost:%s", port)
	log.Printf("Redirect URL configured as: %s", redirectURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	authenticated := client.IsAuthenticated()

	html := `<!DOCTYPE html>
<html>
<head><title>SmartThings OAuth Example</title></head>
<body>
<h1>SmartThings OAuth Example</h1>`

	if authenticated {
		html += `
<p>Status: <strong style="color: green;">Authenticated</strong></p>
<ul>
  <li><a href="/devices">View Devices</a></li>
  <li><a href="/logout">Logout</a></li>
</ul>`
	} else {
		html += `
<p>Status: <strong style="color: red;">Not Authenticated</strong></p>
<p><a href="/login">Login with SmartThings</a></p>`
	}

	html += `</body></html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate secure random state
	state, err := generateState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state for validation
	stateMu.Lock()
	stateStore[state] = true
	stateMu.Unlock()

	// Redirect to SmartThings authorization
	authURL := client.GetAuthorizationURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Check for errors
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		http.Error(w, fmt.Sprintf("OAuth error: %s - %s", errParam, errDesc), http.StatusBadRequest)
		return
	}

	// Validate state
	state := r.URL.Query().Get("state")
	stateMu.Lock()
	valid := stateStore[state]
	delete(stateStore, state)
	stateMu.Unlock()

	if !valid {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := client.ExchangeCode(ctx, code); err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func handleDevices(w http.ResponseWriter, r *http.Request) {
	if !client.IsAuthenticated() {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	ctx := context.Background()
	devices, err := client.ListDevices(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list devices: %v", err), http.StatusInternalServerError)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head><title>SmartThings Devices</title></head>
<body>
<h1>Your SmartThings Devices</h1>
<p><a href="/">Back to Home</a></p>
<table border="1" cellpadding="10">
<tr><th>Name</th><th>Type</th><th>Device ID</th></tr>`

	for _, device := range devices {
		html += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td><code>%s</code></td></tr>",
			device.Label, device.Type, device.DeviceID)
	}

	html += `</table></body></html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if err := client.ClearTokens(ctx); err != nil {
		log.Printf("Failed to clear tokens: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
