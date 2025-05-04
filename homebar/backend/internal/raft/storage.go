package raft

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// Storage interface for Raft state persistence
type Storage interface {
	// SaveState persists the current term and votedFor
	SaveState(term uint64, votedFor string) error

	// LoadState loads the saved term and votedFor
	LoadState() (term uint64, votedFor string, err error)

	// AppendLog appends entries to the log file
	AppendLog(entries []LogEntry) error

	// LoadLog loads all log entries
	LoadLog() ([]LogEntry, error)

	// Close releases any resources
	Close() error
}

// FileStorage implements the Storage interface using files
type FileStorage struct {
	mu        sync.Mutex
	stateFile string
	logFile   string
	dir       string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(nodeID string, dir string) (Storage, error) {
	// Create directory if it doesn't exist
	if dir == "" {
		dir = "raft-data"
	}

	fullDir := filepath.Join(dir, fmt.Sprintf("node-%s", nodeID))
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &FileStorage{
		stateFile: filepath.Join(fullDir, "state.json"),
		logFile:   filepath.Join(fullDir, "log.json"),
		dir:       fullDir,
	}, nil
}

// State represents the persistent Raft state
type persistentState struct {
	CurrentTerm uint64 `json:"current_term"`
	VotedFor    string `json:"voted_for"`
}

// SaveState persists the current term and votedFor
func (fs *FileStorage) SaveState(term uint64, votedFor string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state := persistentState{
		CurrentTerm: term,
		VotedFor:    votedFor,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to a temporary file first, then rename for atomicity
	tmpFile := fs.stateFile + ".tmp"
	if err := ioutil.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return os.Rename(tmpFile, fs.stateFile)
}

// LoadState loads the saved term and votedFor
func (fs *FileStorage) LoadState() (uint64, string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// If file doesn't exist, return defaults
	if _, err := os.Stat(fs.stateFile); os.IsNotExist(err) {
		return 0, "", nil
	}

	data, err := ioutil.ReadFile(fs.stateFile)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read state file: %w", err)
	}

	var state persistentState
	if err := json.Unmarshal(data, &state); err != nil {
		return 0, "", fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return state.CurrentTerm, state.VotedFor, nil
}

// AppendLog appends entries to the log file
func (fs *FileStorage) AppendLog(entries []LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Load existing log
	var log []LogEntry
	existingLog, err := fs.loadLogInternal()
	if err == nil {
		log = existingLog
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing log: %w", err)
	}

	// Append new entries
	log = append(log, entries...)

	// Write back the full log
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	tmpFile := fs.logFile + ".tmp"
	if err := ioutil.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return os.Rename(tmpFile, fs.logFile)
}

// LoadLog loads all log entries
func (fs *FileStorage) LoadLog() ([]LogEntry, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.loadLogInternal()
}

// loadLogInternal is a helper function that loads the log without locking
func (fs *FileStorage) loadLogInternal() ([]LogEntry, error) {
	// If file doesn't exist, return empty log with dummy entry
	if _, err := os.Stat(fs.logFile); os.IsNotExist(err) {
		return []LogEntry{{Term: 0, Index: 0}}, nil
	}

	data, err := ioutil.ReadFile(fs.logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	var log []LogEntry
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("failed to unmarshal log: %w", err)
	}

	// Ensure log has at least a dummy entry
	if len(log) == 0 {
		log = append(log, LogEntry{Term: 0, Index: 0})
	}

	return log, nil
}

// Close releases any resources
func (fs *FileStorage) Close() error {
	// Nothing to close for file storage
	return nil
}
