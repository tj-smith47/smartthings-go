# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-12-04

### Added
- **Core Client**
  - `NewClient` with Personal Access Token authentication
  - Configurable base URL, timeout, and HTTP client
  - Automatic retry with exponential backoff
  - Rate limit detection and callbacks
  - Response caching for capabilities and device profiles
  - HTTP/2 support enabled by default

- **Batch Operations**
  - `ExecuteCommandsBatch` for concurrent command execution on multiple devices
  - `ExecuteCommandBatch` convenience wrapper for same command on multiple devices
  - `GetDeviceStatusBatch` for concurrent status fetching
  - Configurable concurrency limits and stop-on-error behavior

- **Enhanced Rate Limiting**
  - `RateLimitError` with `RetryAfter` duration from Retry-After header
  - `WaitForRateLimit` and `WaitForRateLimitErr` helpers
  - `ShouldThrottle` for preemptive throttling
  - `RateLimitThrottler` for automatic throttling in bulk operations
  - `RemainingRequests` and `RateLimitResetTime` convenience methods

- **Structured Logging (slog)**
  - `WithLogger` option for structured logging with log/slog
  - `LoggingTransport` for HTTP request/response logging
  - `NewLoggingClient` convenience constructor
  - `LogRequest`, `LogResponse`, `LogRateLimit`, `LogDeviceCommand` helpers
  - `LogWebhookEvent` for webhook event logging

- **OAuth 2.0 Authentication**
  - `OAuthClient` with automatic token refresh
  - `TokenStore` interface for persistent token storage
  - `FileTokenStore` implementation
  - Authorization URL generation and code exchange
  - Thread-safe token management

- **Device Management** (171+ API methods)
  - List, get, update, delete devices
  - Execute commands (single and batch)
  - Device status and health monitoring
  - Component-level status retrieval
  - Virtual device creation and management
  - Device event history with pagination

- **Locations & Rooms**
  - CRUD operations for locations
  - Room management within locations
  - Mode management (home, away, night, etc.)

- **Automation**
  - Scene listing and execution
  - Rule CRUD operations
  - Schedule management with cron expressions
  - Subscription/webhook management

- **Capabilities & Profiles**
  - Capability introspection and caching
  - Device profile management
  - Device preferences configuration

- **Edge Drivers & Channels**
  - Driver listing and management
  - Channel operations
  - Hub driver installation

- **Schema Apps (SmartApp Connector)**
  - Schema app CRUD operations
  - Installed schema app management
  - Device state callbacks

- **Pagination Iterators**
  - Go 1.23+ iter.Seq2 iterators for all list endpoints
  - Automatic page fetching
  - Context cancellation support
  - `Devices`, `Locations`, `Rooms`, `Scenes`, `Rules`, `Apps`, `Capabilities`, `DeviceProfiles`, `Drivers`, `Channels`, `Organizations`, and more

- **SSDP Discovery**
  - `FindHubs` for SmartThings hub discovery
  - `FindTVs` for Samsung TV discovery
  - `DiscoverAll` for combined discovery

- **TV Remote Control**
  - Power on/off
  - Volume control (up/down/mute)
  - Input source selection
  - App launching
  - Picture and sound mode configuration

- **Appliance Status Extraction**
  - Samsung washer/dryer status parsing
  - Dishwasher cycle detection
  - Range/oven temperature and mode
  - Refrigerator compartment status

- **Webhook Support**
  - HMAC signature validation
  - Confirmation challenge handling
  - Event type constants

- **Error Handling**
  - Structured API error responses
  - Sentinel errors for common conditions
  - Rate limit information extraction

- **Documentation**
  - Comprehensive README with examples
  - godoc comments on all public APIs
  - OAuth server example in `examples/oauth-server`
  - Webhook handler example in `examples/webhook-handler`
  - SECURITY.md with vulnerability reporting policy
  - Breaking change policy in CONTRIBUTING.md
  - GitHub issue templates for bugs and features

- **Testing Infrastructure**
  - Integration test mode with `//go:build integration` tag
  - Benchmark tests for JSON, cache, webhook, and iterator operations
  - Fuzz tests for webhook signature and JSON parsing
  - Comprehensive godoc examples
  - 91.8% test coverage

### Changed
- Nothing (initial release)

### Deprecated
- Nothing (initial release)

### Removed
- Nothing (initial release)

### Fixed
- Nothing (initial release)

### Security
- HMAC-SHA256 signature validation for webhooks
- Secure token storage with file permissions

## [0.1.0] - 2025-11-15

### Added
- Initial development release
- Basic device and location operations
- OAuth flow implementation

[Unreleased]: https://github.com/tj-smith47/smartthings-go/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/tj-smith47/smartthings-go/compare/v0.1.0...v1.0.0
[0.1.0]: https://github.com/tj-smith47/smartthings-go/releases/tag/v0.1.0
