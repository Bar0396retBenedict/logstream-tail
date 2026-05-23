package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckpoint_GetReturnsEmptyWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	cp, err := NewCheckpoint(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cp.Get("key"); got != "" {
		t.Errorf("expected empty cursor, got %q", got)
	}
}

func TestCheckpoint_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	cp, _ := NewCheckpoint(path)

	if err := cp.Set("group-a", "token-1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if got := cp.Get("group-a"); got != "token-1" {
		t.Errorf("expected token-1, got %q", got)
	}
}

func TestCheckpoint_PersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	cp1, _ := NewCheckpoint(path)
	_ = cp1.Set("group-b", "cursor-xyz")

	// Reload from the same file.
	cp2, err := NewCheckpoint(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if got := cp2.Get("group-b"); got != "cursor-xyz" {
		t.Errorf("expected cursor-xyz after reload, got %q", got)
	}
}

func TestCheckpoint_MultipleKeys(t *testing.T) {
	dir := t.TempDir()
	cp, _ := NewCheckpoint(filepath.Join(dir, "state.json"))

	_ = cp.Set("k1", "v1")
	_ = cp.Set("k2", "v2")

	if cp.Get("k1") != "v1" || cp.Get("k2") != "v2" {
		t.Error("multiple keys not stored correctly")
	}
}

func TestCheckpoint_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	_ = os.WriteFile(path, []byte("not json"), 0o644)

	_, err := NewCheckpoint(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDefaultCheckpointConfig(t *testing.T) {
	cfg := DefaultCheckpointConfig()
	if cfg.FilePath != "" {
		t.Errorf("expected empty FilePath, got %q", cfg.FilePath)
	}
}

func TestNewCheckpointConfig_WithFile(t *testing.T) {
	cfg := NewCheckpointConfig(WithCheckpointFile("/tmp/state.json"))
	if cfg.FilePath != "/tmp/state.json" {
		t.Errorf("expected /tmp/state.json, got %q", cfg.FilePath)
	}
}
