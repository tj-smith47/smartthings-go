# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

1. **Do NOT** create a public GitHub issue for security vulnerabilities
2. Email security concerns to the maintainers directly
3. Include the following in your report:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: Within 48 hours of your report
- **Status Update**: Within 7 days with our assessment
- **Resolution Timeline**: Security patches typically released within 30 days

### Scope

The following are in scope for security reports:

- Authentication/authorization bypasses
- Token/credential exposure
- Webhook signature validation bypasses
- Injection vulnerabilities
- Denial of service vulnerabilities
- Sensitive data exposure

### Out of Scope

- Issues in dependencies (report these to the respective projects)
- Issues requiring physical access to a user's machine
- Social engineering attacks
- Issues in the SmartThings API itself (report to Samsung)

## Security Best Practices

When using this library, follow these security recommendations:

### Token Storage

```go
// DO: Use secure storage for tokens
store := &SecureTokenStore{} // Your implementation with encryption

// DON'T: Store tokens in plaintext files with world-readable permissions
```

### Webhook Validation

```go
// ALWAYS validate webhook signatures
valid := smartthings.ValidateWebhookSignature(body, signature, secret)
if !valid {
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
}
```

### HTTPS Only

This library enforces HTTPS for all API communications. Do not disable TLS verification in production.

### Token Permissions

Request only the minimum required OAuth scopes for your application:

```go
config := &smartthings.OAuthConfig{
    Scopes: []string{
        "r:devices:*",  // Read devices only if you don't need write
        // Don't request "w:devices:*" unless necessary
    },
}
```

## Security Features

This library includes several security features:

1. **HMAC-SHA256 Webhook Validation**: Prevents webhook spoofing
2. **Secure Token Refresh**: Tokens are refreshed before expiry
3. **TLS Enforcement**: All API calls use HTTPS
4. **No Credential Logging**: Tokens are never logged by default

## Acknowledgments

We appreciate the security research community's efforts in responsibly disclosing vulnerabilities.
