// Package memory provides persistent session memory storage and summarization.
//
// Memory allows 2M Code to retain context across sessions, so agents can
// recall what was done in previous runs. Each team has its own JSONL file
// at ~/.2mcode/memory/<team>.jsonl containing chronologically ordered entries.
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Entry represents a single memory entry from a completed session.
type Entry struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	TeamName  string    `json:"team_name"`
	Task      string    `json:"task"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}

// Store defines the interface for reading and writing memory entries.
type Store interface {
	Save(entry Entry) error
	LoadRecent(teamName string, limit int) ([]Entry, error)
	All(teamName string) ([]Entry, error)
	Clear(teamName string) error
}

// FileStore persists memory entries as JSONL files in a directory.
// Thread-safe for concurrent reads/writes.
type FileStore struct {
	dir string
	mu  sync.RWMutex
}

// NewFileStore creates a FileStore rooted at dir (usually ~/.2mcode/memory/).
// Creates the directory if it does not exist.
func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("memory: cannot create directory %s: %w", dir, err)
	}
	return &FileStore{dir: dir}, nil
}

// path returns the JSONL file path for a given team.
func (fs *FileStore) path(teamName string) string {
	safe := strings.NewReplacer(" ", "_", "/", "_", "\\", "_").Replace(teamName)
	return filepath.Join(fs.dir, safe+".jsonl")
}

// Save appends a memory entry to the team's JSONL file.
func (fs *FileStore) Save(entry Entry) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	f, err := os.OpenFile(fs.path(entry.TeamName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("memory: cannot open file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("memory: cannot marshal entry: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("memory: cannot write entry: %w", err)
	}

	return nil
}

// LoadRecent returns the most recent N entries for a team, newest first.
func (fs *FileStore) LoadRecent(teamName string, limit int) ([]Entry, error) {
	entries, err := fs.All(teamName)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.After(entries[j].CreatedAt)
	})

	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}

// All returns every entry for a team, oldest first.
func (fs *FileStore) All(teamName string) ([]Entry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.path(teamName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("memory: cannot read file: %w", err)
	}

	raw := strings.TrimSpace(string(data))
	if raw == "" {
		return nil, nil
	}

	var entries []Entry
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Clear removes all memory entries for a team.
func (fs *FileStore) Clear(teamName string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := os.Remove(fs.path(teamName)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("memory: cannot clear file: %w", err)
	}
	return nil
}
