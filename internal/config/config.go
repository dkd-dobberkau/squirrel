package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// AckEntry represents a single acknowledged project.
type AckEntry struct {
	Path      string     `json:"path"`
	AckedAt   time.Time  `json:"ackedAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// Config is the top-level squirrel configuration.
type Config struct {
	Acknowledged []AckEntry `json:"acknowledged"`
}

var durationRe = regexp.MustCompile(`^(\d+)([dwm])$`)

// ParseDuration parses a human-friendly duration like "7d", "2w", "3m".
func ParseDuration(s string) (time.Duration, error) {
	matches := durationRe.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration %q (use e.g. 7d, 2w, 3m)", s)
	}

	n, _ := strconv.Atoi(matches[1])
	if n <= 0 {
		return 0, fmt.Errorf("duration must be positive: %q", s)
	}

	var days int
	switch matches[2] {
	case "d":
		days = n
	case "w":
		days = n * 7
	case "m":
		days = n * 30
	}

	return time.Duration(days) * 24 * time.Hour, nil
}

// DefaultPath returns the default config file path.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "squirrel", "config.json")
}

// Load reads the config from path. Returns empty config if file doesn't exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to path, creating parent directories as needed.
func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// IsAcknowledged checks if a project path is acknowledged and not expired.
func (c *Config) IsAcknowledged(path string) bool {
	for _, e := range c.Acknowledged {
		if e.Path == path {
			if e.ExpiresAt != nil && e.ExpiresAt.Before(time.Now()) {
				return false
			}
			return true
		}
	}
	return false
}

// Ack adds or updates an acknowledgement for a project path.
func (c *Config) Ack(path string, expiresAt *time.Time) {
	for i, e := range c.Acknowledged {
		if e.Path == path {
			c.Acknowledged[i].AckedAt = time.Now()
			c.Acknowledged[i].ExpiresAt = expiresAt
			return
		}
	}
	c.Acknowledged = append(c.Acknowledged, AckEntry{
		Path:      path,
		AckedAt:   time.Now(),
		ExpiresAt: expiresAt,
	})
}

// Unack removes the acknowledgement for a project path. Returns true if found.
func (c *Config) Unack(path string) bool {
	for i, e := range c.Acknowledged {
		if e.Path == path {
			c.Acknowledged = append(c.Acknowledged[:i], c.Acknowledged[i+1:]...)
			return true
		}
	}
	return false
}
