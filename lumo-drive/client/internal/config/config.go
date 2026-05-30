// Package config persists CLI configuration (server URL, auth token, the
// designated sync directory, and the logged-in user) to a JSON file under the
// user's config directory.
package config

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultAPIURL is the lumo-drive server base URL used when none is configured.
const DefaultAPIURL = "http://localhost:6000"

// defaultConfigDirName is the config subdirectory name used when $CONFIG_DIR_NAME
// is not set (e.g. via .env).
const defaultConfigDirName = "lumo"

// configDirName returns the config subdirectory name, read from
// $CONFIG_DIR_NAME (typically loaded from .env), falling back to "lumo".
func configDirName() string {
	if name := strings.TrimSpace(os.Getenv("CONFIG_DIR_NAME")); name != "" {
		return name
	}
	return defaultConfigDirName
}

// User mirrors the subset of the server's user record we care about.
type User struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Config is the on-disk client configuration.
type Config struct {
	APIURL   string `json:"api_url"`
	ClientID string `json:"client_id,omitempty"`
	Token    string `json:"token,omitempty"`
	SyncDir  string `json:"sync_dir,omitempty"`
	User     *User  `json:"user,omitempty"`
}

// Path returns the location of the config file, honoring $LUMO_CONFIG for tests.
func Path() (string, error) {
	if p := os.Getenv("LUMO_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, configDirName(), "config.json")
	fmt.Println("Config path:", path) // Debugging output
	return path, nil
}

// newClientID returns a random RFC 4122 version 4 UUID string. It identifies
// this client install so the server can distinguish devices.
func newClientID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// Load reads the config from disk. A missing file yields a default config so
// first-run commands (register/login) work without prior setup.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := &Config{APIURL: DefaultAPIURL}
			if err := cfg.ensureClientID(); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}
	if err := cfg.ensureClientID(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ensureClientID assigns a freshly generated client id and persists it when one
// is not already set, so the identifier is stable across runs.
func (c *Config) ensureClientID() error {
	if c.ClientID != "" {
		return nil
	}
	id, err := newClientID()
	if err != nil {
		return err
	}
	c.ClientID = id
	return c.Save()
}

// Save writes the config to disk, creating parent directories as needed. The
// file is written 0600 since it holds the auth token.
func (c *Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// LoggedIn reports whether an auth token is present.
func (c *Config) LoggedIn() bool {
	return c.Token != ""
}
