package source

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Checkpoint persists the last-seen cursor position for a log source so that
// restarts resume from where the process left off rather than re-reading old
// events.
type Checkpoint struct {
	mu   sync.Mutex
	path string
	data checkpointData
}

type checkpointData struct {
	Cursors   map[string]string `json:"cursors"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// NewCheckpoint loads an existing checkpoint file or creates an empty one.
func NewCheckpoint(path string) (*Checkpoint, error) {
	cp := &Checkpoint{
		path: path,
		data: checkpointData{Cursors: make(map[string]string)},
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cp, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &cp.data); err != nil {
		return nil, err
	}
	if cp.data.Cursors == nil {
		cp.data.Cursors = make(map[string]string)
	}
	return cp, nil
}

// Get returns the stored cursor for the given source key, or an empty string
// if no cursor has been saved yet.
func (c *Checkpoint) Get(key string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data.Cursors[key]
}

// Set updates the cursor for key and flushes the checkpoint file to disk.
func (c *Checkpoint) Set(key, cursor string) error {
	c.mu.Lock()
	c.data.Cursors[key] = cursor
	c.data.UpdatedAt = time.Now().UTC()
	snap := c.data
	c.mu.Unlock()

	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, b, 0o644)
}
