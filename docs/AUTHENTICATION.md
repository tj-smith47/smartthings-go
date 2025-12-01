# SmartThings Authentication Options

This document outlines authentication methods for SmartThings API integration and addresses the "daily key rotation" issue.

## Current Implementation: Personal Access Token (PAT)

**How it works:**
- PAT passed to `NewClient(token string)` in `client.go`
- Token sent as `Authorization: Bearer <token>` header on every request

**Pros:**
- Simple implementation
- No OAuth flow required
- Direct API access

**Cons:**
- Reported daily expiration (see investigation below)
- Manual token generation via SmartThings dashboard
- No automatic refresh mechanism

**Current usage:**
```go
client := smartthings.NewClient(os.Getenv("SMARTTHINGS_TOKEN"))
```

---

## Investigation: Why Does PAT Expire Daily?

According to [SmartThings documentation](https://community.smartthings.com/t/changes-to-personal-access-tokens-pat/292019), PATs **should not expire** unless manually revoked. If you're experiencing daily expiration, possible causes:

### 1. Rate Limiting (Most Likely)
SmartThings API has rate limits:
- **250 requests per minute** per token
- **1000 requests per hour** per token

**Current polling:** 15-second interval = 4 requests/minute = 240 requests/hour

**Check if you're hitting limits:**
```bash
# Look for HTTP 429 errors in API logs
grep "429" /db/appdata/api/logs/*.log

# Check SmartThings API response headers
curl -H "Authorization: Bearer $TOKEN" \
  https://api.smartthings.com/v1/devices \
  -v 2>&1 | grep -i "rate\|limit"
```

### 2. Token Scope Mismatch
PAT tokens require specific scopes. Check your token has:
- `r:devices:*` (read devices)
- `x:devices:*` (execute commands)

**Recreate token with correct scopes:**
1. Go to https://account.smartthings.com/tokens
2. Delete existing token
3. Create new token with scopes: `r:devices:*, x:devices:*`

### 3. Account-Level Expiration Policy
Some SmartThings accounts (especially developer accounts) may have token expiration policies.

**Check in Developer Workspace:**
- https://smartthings.developer.samsung.com/workspace
- Look for token lifecycle settings

---

## Option 1: OAuth 2.0 (Recommended for Production)

OAuth provides long-lived refresh tokens that auto-renew access tokens, eliminating daily rotation.

### Benefits
- **Long-lived refresh tokens** (no expiration unless revoked)
- **Automatic token renewal** (access tokens refresh every ~24h automatically)
- More secure (tokens can be revoked per-app)
- Standard OAuth 2.0 flow
- Scoped permissions

### Implementation Overview

**1. Register SmartThings App:**
- Go to https://smartthings.developer.samsung.com/
- Create new "Automation for the SmartThings App"
- Configure OAuth redirect URL: `https://api.jarvispro.io/auth/smartthings/callback`
- Save `client_id` and `client_secret`

**2. OAuth Flow:**
```
User → Authorization URL → SmartThings Login → Redirect with code
→ Exchange code for tokens → Store refresh_token → Auto-refresh access_token
```

**3. Code Changes Required:**

**New file: `smartthings-go/oauth.go`**
```go
package smartthings

import (
    "context"
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
    "time"
)

type OAuthConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
}

func (c *Client) GetAuthorizationURL(cfg *OAuthConfig, state string) string {
    params := url.Values{}
    params.Set("response_type", "code")
    params.Set("client_id", cfg.ClientID)
    params.Set("redirect_uri", cfg.RedirectURL)
    params.Set("scope", "r:devices:* x:devices:*")
    params.Set("state", state)
    return "https://api.smartthings.com/oauth/authorize?" + params.Encode()
}

func (c *Client) ExchangeCodeForToken(ctx context.Context, cfg *OAuthConfig, code string) (*TokenResponse, error) {
    data := url.Values{}
    data.Set("grant_type", "authorization_code")
    data.Set("client_id", cfg.ClientID)
    data.Set("client_secret", cfg.ClientSecret)
    data.Set("redirect_uri", cfg.RedirectURL)
    data.Set("code", code)

    req, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.smartthings.com/oauth/token",
        strings.NewReader(data.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var tokens TokenResponse
    json.NewDecoder(resp.Body).Decode(&tokens)
    return &tokens, nil
}

func (c *Client) RefreshAccessToken(ctx context.Context, cfg *OAuthConfig, refreshToken string) (*TokenResponse, error) {
    data := url.Values{}
    data.Set("grant_type", "refresh_token")
    data.Set("client_id", cfg.ClientID)
    data.Set("client_secret", cfg.ClientSecret)
    data.Set("refresh_token", refreshToken)

    req, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.smartthings.com/oauth/token",
        strings.NewReader(data.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var tokens TokenResponse
    json.NewDecoder(resp.Body).Decode(&tokens)
    return &tokens, nil
}
```

**4. Modify `client.go` for auto-refresh:**
```go
type Client struct {
    httpClient  *http.Client
    token       string
    oauthConfig *OAuthConfig
    refreshToken string
    tokenExpiry time.Time
}

func (c *Client) ensureValidToken(ctx context.Context) error {
    if c.oauthConfig != nil && time.Now().After(c.tokenExpiry.Add(-5*time.Minute)) {
        tokens, err := c.RefreshAccessToken(ctx, c.oauthConfig, c.refreshToken)
        if err != nil {
            return err
        }
        c.token = tokens.AccessToken
        c.refreshToken = tokens.RefreshToken
        c.tokenExpiry = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
        // Store tokens to database or encrypted file
    }
    return nil
}

// Call ensureValidToken() before each API request in devices.go
```

**5. API Server Handler:**

Add to `/db/appdata/api/handlers/auth.go`:
```go
func HandleSmartThingsOAuthCallback(c *gin.Context) {
    code := c.Query("code")
    state := c.Query("state")

    // Verify state (CSRF protection)
    // ...

    client := smartthings.NewClient("")
    tokens, err := client.ExchangeCodeForToken(context.Background(), oauthConfig, code)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Store tokens securely in PostgreSQL
    db.Exec("UPDATE smartthings_config SET access_token = ?, refresh_token = ?, expires_at = ? WHERE id = 1",
        tokens.AccessToken, tokens.RefreshToken, time.Now().Add(time.Duration(tokens.ExpiresIn)*time.Second))

    c.JSON(200, gin.H{"message": "Authentication successful"})
}
```

### Effort Estimate
- smartthings-go library changes: **3 hours**
- API server integration: **2 hours**
- Testing: **1 hour**
- **Total: ~6 hours**

---

## Option 2: Keep PAT with Investigation

**If PATs don't actually expire**, the "daily rotation" issue is likely:
1. **Rate limiting** - Hitting API limits, not token expiration
2. **Cache issues** - Client/server not using latest token from env
3. **Token scope** - Wrong scopes causing certain requests to fail

### Recommended Actions
1. Enable verbose logging to capture actual API error responses
2. Check for HTTP 429 (rate limit) vs 401 (auth) errors
3. Verify token scopes match requirements
4. Consider increasing polling interval to 30s or 60s

---

## Recommendation

**For homelab/personal use:**
1. **First**, investigate if PAT truly expires or if it's rate limiting
2. **If rate limiting**: Reduce polling frequency (30s instead of 15s)
3. **If actual expiration**: Contact SmartThings support - PATs shouldn't expire
4. **Last resort**: Implement OAuth if PAT issues can't be resolved

**For production/multi-user:**
- Implement OAuth 2.0 for better security and UX

---

## Resources
- [SmartThings OAuth Documentation](https://smartthings.developer.samsung.com/docs/auth-and-permissions/oauth.html)
- [SmartThings API Reference](https://smartthings.developer.samsung.com/docs/api-ref/st-api.html)
- [PAT Token Changes Announcement](https://community.smartthings.com/t/changes-to-personal-access-tokens-pat/292019)
