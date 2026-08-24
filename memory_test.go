package main

import (
	"os"
	"testing"
	"time"
)

func TestAppendAndLoadMemories(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("OLLAMA_MEM_DIR", tmpDir)

	m1 := Memory{Text: "first memory", Embedding: []float64{0.1, 0.2}, CreatedAt: time.Now()}
	m2 := Memory{Text: "second memory", Embedding: []float64{0.3, 0.4}, CreatedAt: time.Now()}

	if err := appendMemory(m1); err != nil {
		t.Fatalf("appendMemory(m1) failed: %v", err)
	}
	if err := appendMemory(m2); err != nil {
		t.Fatalf("appendMemory(m2) failed: %v", err)
	}

	mems, err := loadMemories()
	if err != nil {
		t.Fatalf("loadMemories failed: %v", err)
	}
	if len(mems) != 2 {
		t.Fatalf("loadMemories returned %d entries, want 2", len(mems))
	}
	if mems[0].Text != "first memory" || mems[1].Text != "second memory" {
		t.Errorf("loaded memories don't match: got %+v", mems)
	}
}

func TestLoadMemoriesEmptyWhenFileMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("OLLAMA_MEM_DIR", tmpDir)

	mems, err := loadMemories()
	if err != nil {
		t.Fatalf("loadMemories on missing file should not error, got: %v", err)
	}
	if len(mems) != 0 {
		t.Errorf("loadMemories on missing file = %d entries, want 0", len(mems))
	}
}

func TestLoadMemoriesSkipsMalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("OLLAMA_MEM_DIR", tmpDir)

	path := memoryPath()
	content := `{"text":"valid one","embedding":[0.1],"created_at":"2024-01-01T00:00:00Z"}
not valid json at all
{"text":"valid two","embedding":[0.2],"created_at":"2024-01-02T00:00:00Z"}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	mems, err := loadMemories()
	if err != nil {
		t.Fatalf("loadMemories failed: %v", err)
	}
	if len(mems) != 2 {
		t.Fatalf("loadMemories returned %d entries, want 2 (malformed line should be skipped)", len(mems))
	}
}

func TestClearMemoriesRemovesFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("OLLAMA_MEM_DIR", tmpDir)

	if err := appendMemory(Memory{Text: "to be cleared", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("appendMemory failed: %v", err)
	}

	if err := clearMemories(); err != nil {
		t.Fatalf("clearMemories failed: %v", err)
	}

	mems, err := loadMemories()
	if err != nil {
		t.Fatalf("loadMemories after clear failed: %v", err)
	}
	if len(mems) != 0 {
		t.Errorf("expected 0 memories after clear, got %d", len(mems))
	}
}

func TestClearMemoriesOnMissingFileDoesNotError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("OLLAMA_MEM_DIR", tmpDir)

	if err := clearMemories(); err != nil {
		t.Errorf("clearMemories on missing file should not error, got: %v", err)
	}
}
