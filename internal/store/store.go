package store

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Chinzzii/redis-go/pkg/util"
)

// Entry holds a value and an optional expiration timestamp.
type Entry struct {
	Value     string
	ExpiresAt time.Time
}

// Store is the in‑memory key-value store with persistence and pub/sub.
type Store struct {
	mu              sync.RWMutex
	data            map[string]Entry
	subs            map[string]map[net.Conn]bool
	subscribedChans map[net.Conn]map[string]bool
	aofFile         *os.File
}

// NewStore initializes the store, loads AOF or snapshot, and starts expiration cleanup.
func NewStore(aofFilename string) *Store {
	s := &Store{
		data:            make(map[string]Entry),
		subs:            make(map[string]map[net.Conn]bool),
		subscribedChans: make(map[net.Conn]map[string]bool),
	}

	// Load persistence
	if _, err := os.Stat(aofFilename); err == nil {
		_ = s.loadAOF(aofFilename)
	} else if _, err := os.Stat("dump.rdb"); err == nil {
		_ = s.loadSnapshot("dump.rdb")
	}

	// Open AOF for appending
	if f, err := os.OpenFile(aofFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		s.aofFile = f
	} else {
		fmt.Fprintf(os.Stderr, "Warning: could not open AOF file: %v\n", err)
	}

	go s.startExpirationCleaner()
	return s
}

// Close shuts down persistence resources.
func (s *Store) Close() {
	if s.aofFile != nil {
		s.aofFile.Close()
	}
}

// Set stores a key-value pair and logs it to AOF.
func (s *Store) Set(key, value string) string {
	s.mu.Lock()
	s.data[key] = Entry{Value: value}
	s.mu.Unlock()
	if s.aofFile != nil {
		fmt.Fprintf(s.aofFile, "SET %s %s\r\n", key, util.EscapeNewlines(value))
	}
	return "OK"
}

// Get retrieves a value; returns (value, true) or ("", false).
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	entry, ok := s.data[key]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return "", false
	}
	return entry.Value, true
}

// Del removes a key; returns 1 if removed or 0 if not present.
func (s *Store) Del(key string) int {
	s.mu.Lock()
	_, ok := s.data[key]
	if ok {
		delete(s.data, key)
	}
	s.mu.Unlock()
	if ok && s.aofFile != nil {
		fmt.Fprintf(s.aofFile, "DEL %s\r\n", key)
	}
	if ok {
		return 1
	}
	return 0
}
