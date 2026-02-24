package claude

import (
	"encoding/json"
	"time"
)

// HistoryEntry is a single line from ~/.claude/history.jsonl
type HistoryEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"`
	Project   string `json:"project"`
}

// SessionEntry is one entry from sessions-index.json
type SessionEntry struct {
	SessionID   string `json:"sessionId"`
	FullPath    string `json:"fullPath"`
	FirstPrompt string `json:"firstPrompt"`
	Summary     string `json:"summary"`
	MsgCount    int    `json:"messageCount"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
	GitBranch   string `json:"gitBranch"`
	ProjectPath string `json:"projectPath"`
	IsSidechain bool   `json:"isSidechain"`
}

// SessionsIndex is the top-level structure of sessions-index.json
type SessionsIndex struct {
	Version int            `json:"version"`
	Entries []SessionEntry `json:"entries"`
}

// SessionMessage is a single line from a session JSONL file (for deep mode)
type SessionMessage struct {
	Type      string          `json:"type"`
	Message   json.RawMessage `json:"message"`
	Timestamp string          `json:"timestamp"`
	CWD       string          `json:"cwd"`
	GitBranch string          `json:"gitBranch"`
}

// ContentBlock represents a block within a message (text, tool_use, etc.)
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// UserMessage is a parsed user message from a session JSONL
type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AssistantMessage is a parsed assistant message with structured content blocks
type AssistantMessage struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// TodoItem represents a TODO/FIXME extracted from session messages
type TodoItem struct {
	Text      string `json:"text"`
	Source    string `json:"source"`
	SessionID string `json:"sessionId"`
	Timestamp string `json:"timestamp"`
}

// ProjectInfo aggregates all data we know about a project
type ProjectInfo struct {
	Path             string         `json:"path"`
	ShortName        string         `json:"shortName"`
	PromptCount      int            `json:"promptCount"`
	LastActivity     time.Time      `json:"lastActivity"`
	FirstActivity    time.Time      `json:"firstActivity"`
	LastPrompt       string         `json:"lastPrompt"`
	Sessions         []SessionEntry `json:"sessions,omitempty"`
	LatestSummary    string         `json:"latestSummary,omitempty"`
	LatestBranch     string         `json:"latestBranch,omitempty"`
	// Populated by deep analysis
	Todos        []TodoItem `json:"todos,omitempty"`
	LastMessages []string   `json:"lastMessages,omitempty"`
	// Populated by medium/deep analysis
	GitDirty         bool    `json:"gitDirty"`
	GitBranch        string  `json:"gitBranch"`
	UncommittedFiles int     `json:"uncommittedFiles"`
	DaysSinceActive  int     `json:"daysSinceActive"`
	IsOpenWork       bool    `json:"isOpenWork"`
	Score            float64 `json:"score"`
}
