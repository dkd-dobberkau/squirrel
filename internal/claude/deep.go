package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var todoPattern = regexp.MustCompile(`(?i)(?:TODO|FIXME|HACK):\s*(.+)`)
var checkboxPattern = regexp.MustCompile(`- \[ \]\s+(.+)`)

// ParseSessionMessages reads a session JSONL file, keeping only the last maxLines lines
// via a ring buffer to avoid loading huge files entirely into memory.
func ParseSessionMessages(path string, maxLines int) ([]SessionMessage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Ring buffer: keep only the last maxLines raw lines
	ring := make([]string, maxLines)
	idx := 0
	total := 0

	for scanner.Scan() {
		ring[idx%maxLines] = scanner.Text()
		idx++
		total++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	count := total
	if count > maxLines {
		count = maxLines
	}

	var msgs []SessionMessage
	start := 0
	if total > maxLines {
		start = idx % maxLines
	}

	for i := 0; i < count; i++ {
		line := ring[(start+i)%maxLines]
		var msg SessionMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// ExtractText extracts plain text content from a SessionMessage.
// Handles both simple string messages and structured content blocks.
func ExtractText(msg SessionMessage) string {
	if msg.Message == nil {
		return ""
	}

	// Try as simple string first
	var s string
	if err := json.Unmarshal(msg.Message, &s); err == nil {
		return s
	}

	// Try as UserMessage (content is a string)
	var userMsg UserMessage
	if err := json.Unmarshal(msg.Message, &userMsg); err == nil && userMsg.Content != "" {
		return userMsg.Content
	}

	// Try as AssistantMessage (content is []ContentBlock)
	var assistantMsg AssistantMessage
	if err := json.Unmarshal(msg.Message, &assistantMsg); err == nil {
		var parts []string
		for _, block := range assistantMsg.Content {
			if block.Type == "text" && block.Text != "" {
				parts = append(parts, block.Text)
			}
		}
		return strings.Join(parts, "\n")
	}

	return ""
}

// ExtractTodos searches session messages for TODO/FIXME/HACK markers and unchecked checkboxes.
func ExtractTodos(messages []SessionMessage, sessionID string) []TodoItem {
	seen := make(map[string]bool)
	var todos []TodoItem

	for _, msg := range messages {
		text := ExtractText(msg)
		if text == "" {
			continue
		}

		for _, match := range todoPattern.FindAllStringSubmatch(text, -1) {
			item := strings.TrimSpace(match[1])
			if !seen[item] {
				seen[item] = true
				todos = append(todos, TodoItem{
					Text:      item,
					Source:    "TODO",
					SessionID: sessionID,
					Timestamp: msg.Timestamp,
				})
			}
		}

		for _, match := range checkboxPattern.FindAllStringSubmatch(text, -1) {
			item := strings.TrimSpace(match[1])
			if !seen[item] {
				seen[item] = true
				todos = append(todos, TodoItem{
					Text:      item,
					Source:    "checkbox",
					SessionID: sessionID,
					Timestamp: msg.Timestamp,
				})
			}
		}
	}

	return todos
}

// EnrichWithTodos reads session JSONL files for a single project and extracts TODOs.
func EnrichWithTodos(project *ProjectInfo, claudeProjectsDir string) {
	dirName := projectPathToDir(project.Path)
	sessionsDir := filepath.Join(claudeProjectsDir, dirName)

	for _, session := range project.Sessions {
		jsonlPath := filepath.Join(sessionsDir, session.SessionID+".jsonl")
		msgs, err := ParseSessionMessages(jsonlPath, 200)
		if err != nil {
			continue
		}

		todos := ExtractTodos(msgs, session.SessionID)
		project.Todos = append(project.Todos, todos...)

		// Extract last few human messages for context
		var lastMsgs []string
		for i := len(msgs) - 1; i >= 0 && len(lastMsgs) < 5; i-- {
			if msgs[i].Type == "human" || msgs[i].Type == "user" {
				text := ExtractText(msgs[i])
				if text != "" {
					lastMsgs = append(lastMsgs, text)
				}
			}
		}
		project.LastMessages = append(project.LastMessages, lastMsgs...)
	}

	// Limit last messages to avoid bloat
	if len(project.LastMessages) > 10 {
		project.LastMessages = project.LastMessages[:10]
	}
}

// EnrichAllWithTodos enriches all projects with TODO data (for status --deep).
func EnrichAllWithTodos(projects []ProjectInfo, claudeProjectsDir string) {
	for i := range projects {
		EnrichWithTodos(&projects[i], claudeProjectsDir)
	}
}
