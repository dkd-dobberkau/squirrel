package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"7d", 7 * 24 * time.Hour, false},
		{"30d", 30 * 24 * time.Hour, false},
		{"2w", 14 * 24 * time.Hour, false},
		{"3m", 90 * 24 * time.Hour, false},
		{"1m", 30 * 24 * time.Hour, false},
		{"", 0, true},
		{"abc", 0, true},
		{"0d", 0, true},
		{"-5d", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLoadSaveRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{}
	expires := time.Now().Add(30 * 24 * time.Hour)
	cfg.Ack("/projects/foo", nil)
	cfg.Ack("/projects/bar", &expires)

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Acknowledged) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded.Acknowledged))
	}

	if loaded.Acknowledged[0].Path != "/projects/foo" {
		t.Errorf("expected path /projects/foo, got %s", loaded.Acknowledged[0].Path)
	}
}

func TestLoadNonexistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load of nonexistent file should not error: %v", err)
	}
	if len(cfg.Acknowledged) != 0 {
		t.Errorf("expected empty config, got %d entries", len(cfg.Acknowledged))
	}
}

func TestIsAcknowledged(t *testing.T) {
	cfg := &Config{}
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	cfg.Ack("/projects/permanent", nil)
	cfg.Ack("/projects/valid", &future)
	cfg.Ack("/projects/expired", &past)

	if !cfg.IsAcknowledged("/projects/permanent") {
		t.Error("permanent ack should be acknowledged")
	}
	if !cfg.IsAcknowledged("/projects/valid") {
		t.Error("non-expired ack should be acknowledged")
	}
	if cfg.IsAcknowledged("/projects/expired") {
		t.Error("expired ack should NOT be acknowledged")
	}
	if cfg.IsAcknowledged("/projects/unknown") {
		t.Error("unknown project should NOT be acknowledged")
	}
}

func TestAckIdempotent(t *testing.T) {
	cfg := &Config{}
	cfg.Ack("/projects/foo", nil)
	cfg.Ack("/projects/foo", nil)

	if len(cfg.Acknowledged) != 1 {
		t.Errorf("double ack should not create duplicates, got %d", len(cfg.Acknowledged))
	}
}

func TestUnack(t *testing.T) {
	cfg := &Config{}
	cfg.Ack("/projects/foo", nil)
	cfg.Ack("/projects/bar", nil)

	if !cfg.Unack("/projects/foo") {
		t.Error("Unack should return true for existing entry")
	}
	if len(cfg.Acknowledged) != 1 {
		t.Errorf("expected 1 entry after unack, got %d", len(cfg.Acknowledged))
	}
	if cfg.Unack("/projects/unknown") {
		t.Error("Unack should return false for nonexistent entry")
	}
}
