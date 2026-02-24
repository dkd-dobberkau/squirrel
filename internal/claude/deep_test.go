package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTextSimpleString(t *testing.T) {
	raw, _ := json.Marshal("hello world")
	msg := SessionMessage{Message: raw}

	text := ExtractText(msg)
	if text != "hello world" {
		t.Errorf("expected 'hello world', got %q", text)
	}
}

func TestExtractTextUserMessage(t *testing.T) {
	userMsg := UserMessage{Role: "user", Content: "fix the bug"}
	raw, _ := json.Marshal(userMsg)
	msg := SessionMessage{Message: raw}

	text := ExtractText(msg)
	if text != "fix the bug" {
		t.Errorf("expected 'fix the bug', got %q", text)
	}
}

func TestExtractTextAssistantMessage(t *testing.T) {
	assistantMsg := AssistantMessage{
		Role: "assistant",
		Content: []ContentBlock{
			{Type: "text", Text: "Here is the fix"},
			{Type: "tool_use", Text: ""},
			{Type: "text", Text: "Done!"},
		},
	}
	raw, _ := json.Marshal(assistantMsg)
	msg := SessionMessage{Message: raw}

	text := ExtractText(msg)
	if text != "Here is the fix\nDone!" {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestExtractTextNilMessage(t *testing.T) {
	msg := SessionMessage{}
	text := ExtractText(msg)
	if text != "" {
		t.Errorf("expected empty string for nil message, got %q", text)
	}
}

func TestExtractTodosTODO(t *testing.T) {
	raw, _ := json.Marshal("We need to TODO: implement caching here")
	msgs := []SessionMessage{{Message: raw, Timestamp: "2026-02-20T10:00:00Z"}}

	todos := ExtractTodos(msgs, "session-1")
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].Text != "implement caching here" {
		t.Errorf("unexpected todo text: %q", todos[0].Text)
	}
	if todos[0].Source != "TODO" {
		t.Errorf("expected source 'TODO', got %q", todos[0].Source)
	}
}

func TestExtractTodosFIXME(t *testing.T) {
	raw, _ := json.Marshal("FIXME: this is broken")
	msgs := []SessionMessage{{Message: raw}}

	todos := ExtractTodos(msgs, "session-1")
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].Text != "this is broken" {
		t.Errorf("unexpected todo text: %q", todos[0].Text)
	}
}

func TestExtractTodosCheckbox(t *testing.T) {
	raw, _ := json.Marshal("- [ ] Add unit tests\n- [x] Fix typo\n- [ ] Update docs")
	msgs := []SessionMessage{{Message: raw}}

	todos := ExtractTodos(msgs, "session-1")
	if len(todos) != 2 {
		t.Fatalf("expected 2 unchecked checkboxes, got %d", len(todos))
	}
	if todos[0].Text != "Add unit tests" {
		t.Errorf("unexpected first todo: %q", todos[0].Text)
	}
	if todos[1].Text != "Update docs" {
		t.Errorf("unexpected second todo: %q", todos[1].Text)
	}
}

func TestExtractTodosDeduplication(t *testing.T) {
	raw1, _ := json.Marshal("TODO: fix this")
	raw2, _ := json.Marshal("TODO: fix this")
	msgs := []SessionMessage{
		{Message: raw1},
		{Message: raw2},
	}

	todos := ExtractTodos(msgs, "session-1")
	if len(todos) != 1 {
		t.Fatalf("expected 1 deduplicated todo, got %d", len(todos))
	}
}

func TestParseSessionMessages(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")

	lines := []SessionMessage{
		{Type: "human", Message: mustMarshal("hello"), Timestamp: "2026-02-20T10:00:00Z"},
		{Type: "assistant", Message: mustMarshal("hi there"), Timestamp: "2026-02-20T10:00:01Z"},
		{Type: "human", Message: mustMarshal("fix bug"), Timestamp: "2026-02-20T10:00:02Z"},
	}

	var content []byte
	for _, l := range lines {
		b, _ := json.Marshal(l)
		content = append(content, b...)
		content = append(content, '\n')
	}
	os.WriteFile(path, content, 0644)

	msgs, err := ParseSessionMessages(path, 100)
	if err != nil {
		t.Fatalf("ParseSessionMessages failed: %v", err)
	}

	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	if ExtractText(msgs[0]) != "hello" {
		t.Errorf("unexpected first message: %q", ExtractText(msgs[0]))
	}
}

func TestParseSessionMessagesRingBuffer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")

	var content []byte
	for i := 0; i < 10; i++ {
		msg := SessionMessage{
			Type:    "human",
			Message: mustMarshal("message " + string(rune('A'+i))),
		}
		b, _ := json.Marshal(msg)
		content = append(content, b...)
		content = append(content, '\n')
	}
	os.WriteFile(path, content, 0644)

	// Only keep last 3
	msgs, err := ParseSessionMessages(path, 3)
	if err != nil {
		t.Fatalf("ParseSessionMessages failed: %v", err)
	}

	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	// Should have the last 3 messages (H, I, J)
	if ExtractText(msgs[0]) != "message H" {
		t.Errorf("expected 'message H', got %q", ExtractText(msgs[0]))
	}
}

func TestEnrichWithTodos(t *testing.T) {
	claudeDir := t.TempDir()
	projDir := filepath.Join(claudeDir, "-Users-test-myproject")
	os.MkdirAll(projDir, 0755)

	// Create a session JSONL with a TODO
	msg := SessionMessage{
		Type:      "assistant",
		Message:   mustMarshal("TODO: add error handling"),
		Timestamp: "2026-02-20T10:00:00Z",
	}
	b, _ := json.Marshal(msg)
	os.WriteFile(filepath.Join(projDir, "session-abc.jsonl"), append(b, '\n'), 0644)

	project := &ProjectInfo{
		Path:      "/Users/test/myproject",
		ShortName: "myproject",
		Sessions: []SessionEntry{
			{SessionID: "session-abc"},
		},
	}

	EnrichWithTodos(project, claudeDir)

	if len(project.Todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(project.Todos))
	}
	if project.Todos[0].Text != "add error handling" {
		t.Errorf("unexpected todo text: %q", project.Todos[0].Text)
	}
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
