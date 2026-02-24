package config

import (
	"fmt"
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
