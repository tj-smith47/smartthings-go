package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileTokenStore stores OAuth tokens in a JSON file
type FileTokenStore struct {
	filepath string
	mu       sync.RWMutex
}

// NewFileTokenStore creates a new FileTokenStore
func NewFileTokenStore(filepath string) *FileTokenStore {
	return &FileTokenStore{
		filepath: filepath,
	}
}

// SaveTokens saves the tokens to the file
func (f *FileTokenStore) SaveTokens(ctx context.Context, tokens *TokenResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	// Ensure the directory exists
	dir := filepath.Dir(f.filepath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create token directory: %w", err)
		}
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Write to a temporary file first, then rename for atomicity
	tmpFile := f.filepath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	if err := os.Rename(tmpFile, f.filepath); err != nil {
		// Clean up temp file on failure
		os.Remove(tmpFile)
		return fmt.Errorf("failed to save token file: %w", err)
	}

	return nil
}

// LoadTokens loads the tokens from the file
func (f *FileTokenStore) LoadTokens(ctx context.Context) (*TokenResponse, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	data, err := os.ReadFile(f.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("token file not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var tokens TokenResponse
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &tokens, nil
}

// Delete removes the token file
func (f *FileTokenStore) Delete(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.Remove(f.filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}

// Exists checks if the token file exists
func (f *FileTokenStore) Exists() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, err := os.Stat(f.filepath)
	return err == nil
}

// MemoryTokenStore stores tokens in memory (useful for testing)
type MemoryTokenStore struct {
	tokens *TokenResponse
	mu     sync.RWMutex
}

// NewMemoryTokenStore creates a new in-memory token store
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{}
}

// SaveTokens saves tokens to memory
func (m *MemoryTokenStore) SaveTokens(ctx context.Context, tokens *TokenResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens = tokens
	return nil
}

// LoadTokens loads tokens from memory
func (m *MemoryTokenStore) LoadTokens(ctx context.Context) (*TokenResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tokens == nil {
		return nil, fmt.Errorf("no tokens stored")
	}
	return m.tokens, nil
}

// Clear removes stored tokens
func (m *MemoryTokenStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens = nil
}
