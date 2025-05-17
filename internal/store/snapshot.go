package store

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

// StartSnapshotting runs saveSnapshot at the given interval.
func (s *Store) StartSnapshotting(interval time.Duration, filename string) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		if err := s.saveSnapshot(filename); err != nil {
			fmt.Fprintf(os.Stderr, "Snapshot error: %v\n", err)
		}
	}
}

// saveSnapshot writes the entire data map to a file.
func (s *Store) saveSnapshot(filename string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(s.data)
}

// loadSnapshot loads data from a snapshot file.
func (s *Store) loadSnapshot(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	var data map[string]Entry
	if err := gob.NewDecoder(f).Decode(&data); err != nil {
		return err
	}
	s.mu.Lock()
	s.data = data
	s.mu.Unlock()
	return nil
}
