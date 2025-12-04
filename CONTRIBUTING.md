# Contributing to smartthings-go

Thank you for your interest in contributing to smartthings-go! This document provides guidelines and information for contributors.

## Prerequisites

- Go 1.23 or later (uses `iter.Seq` for pagination iterators)
- A SmartThings developer account for testing (optional but recommended)
- Git

## Development Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/smartthings-go.git
   cd smartthings-go
   ```

2. **Verify dependencies:**
   ```bash
   go mod verify
   ```

3. **Run tests:**
   ```bash
   go test -v -race ./...
   ```

4. **Run tests with coverage:**
   ```bash
   go test -v -race -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out  # View coverage report
   ```

## Code Style

This project follows standard Go conventions:

- **Formatting:** Use `gofmt` or `goimports`
- **Linting:** Run `go vet ./...` before submitting
- **Error handling:** Use the `fmt.Errorf("FunctionName: operation: %w", err)` pattern
- **Documentation:** All exported functions must have godoc comments
- **Tests:** Aim for 80%+ coverage on new code

### Error Message Format

```go
// Good
return fmt.Errorf("GetDevice: request failed: %w", err)
return fmt.Errorf("CreateLocation: invalid name: %w", ErrInvalidName)

// Bad
return fmt.Errorf("failed to get device: %w", err)
return errors.New("invalid name")
```

### Test Style

Use table-driven tests:

```go
func TestGetString(t *testing.T) {
    tests := []struct {
        name     string
        data     map[string]any
        path     []string
        want     string
        wantOK   bool
    }{
        {
            name:   "nested path",
            data:   map[string]any{"a": map[string]any{"b": "value"}},
            path:   []string{"a", "b"},
            want:   "value",
            wantOK: true,
        },
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, ok := GetString(tt.data, tt.path...)
            if ok != tt.wantOK || got != tt.want {
                t.Errorf("GetString() = %q, %v; want %q, %v", got, ok, tt.want, tt.wantOK)
            }
        })
    }
}
```

## Making Changes

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation changes
- `test/description` - Test improvements

### Commit Messages

Follow conventional commits:

```
feat: add webhook signature validation
fix: handle nil device status
docs: update OAuth example
test: add iterator error path tests
refactor: consolidate helper functions
```

### Pull Request Process

1. **Create a feature branch** from `main`
2. **Write tests** for new functionality
3. **Ensure all tests pass:** `go test -race ./...`
4. **Run go vet:** `go vet ./...`
5. **Update documentation** if API changes
6. **Submit PR** with clear description

### PR Description Template

```markdown
## Summary
Brief description of changes.

## Changes
- Added X
- Fixed Y
- Updated Z

## Testing
- [ ] Unit tests added/updated
- [ ] Manual testing performed
- [ ] Coverage maintained at 80%+

## Related Issues
Fixes #123
```

## Testing Guidelines

### Unit Tests

- Mock HTTP responses using `httptest.Server`
- Use the `testClient` helper for consistent test setup
- Test both success and error paths

### What Not to Unit Test

Some functions require live network operations and cannot be unit tested:

- `FindHubs()`, `FindTVs()`, `DiscoverAll()` - SSDP multicast
- These are tested via integration tests (see below)

### Integration Tests

For testing against the live SmartThings API:

```bash
# Run integration tests (requires SMARTTHINGS_TOKEN env var)
SMARTTHINGS_TOKEN=your-token go test -v -tags=integration ./...
```

Integration tests are skipped by default.

## API Coverage

The library aims for complete SmartThings API v1 coverage. When adding new endpoints:

1. Add the method to `interface.go`
2. Implement in the appropriate file (e.g., `devices.go`, `locations.go`)
3. Add tests with mocked HTTP responses
4. Update README.md if it's a major feature

## Versioning and Breaking Changes

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### What Is a Breaking Change

1. **Removed** public functions, types, or constants
2. **Changed** function signatures (parameters, return types)
3. **Changed** struct field types or removed public fields
4. **Changed** behavior in a way that could cause existing code to fail

### What Is NOT a Breaking Change

- Adding new public functions, types, or methods
- Adding new optional parameters via functional options
- Adding new fields to structs (unless they affect JSON marshaling)
- Performance improvements
- Bug fixes (even if they change behavior to match documentation)
- Internal/unexported changes

### Deprecation Process

When deprecating functionality:

1. Add a `// Deprecated:` comment with the deprecation reason and alternative
2. Keep deprecated functionality for at least one MINOR version
3. Remove deprecated functionality only in the next MAJOR version
4. Document all deprecations in the CHANGELOG

Example:
```go
// Deprecated: Use NewClientWithOptions instead. This function will be
// removed in v2.0.0.
func OldFunction() { ... }
```

## Release Process

Releases are automated via GitHub Actions when a tag is pushed:

```bash
git tag v1.x.x
git push origin v1.x.x
```

GoReleaser creates the GitHub release automatically.

## Getting Help

- **Issues:** Open a GitHub issue for bugs or feature requests
- **Discussions:** Use GitHub Discussions for questions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
