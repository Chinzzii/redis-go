package store

import (
	"fmt"
	"time"
)

// Expire sets a TTL on a key (in seconds). Returns 1 if set, 0 if key missing.
func (s *Store) Expire(key string, seconds int) int {
	s.mu.Lock()
	entry, ok := s.data[key]
	if !ok {
		s.mu.Unlock()
		return 0
	}
	if seconds <= 0 {
		delete(s.data, key)
		s.mu.Unlock()
		if s.aofFile != nil {
			fmt.Fprintf(s.aofFile, "DEL %s\r\n", key)
		}
		return 1
	}
	entry.ExpiresAt = time.Now().Add(time.Duration(seconds) * time.Second)
	s.data[key] = entry
	s.mu.Unlock()
	if s.aofFile != nil {
		fmt.Fprintf(s.aofFile, "EXPIRE %s %d\r\n", key, seconds)
	}
	return 1
}

// TTL returns remaining TTL in seconds, -2 if missing, -1 if no TTL.
func (s *Store) TTL(key string) int {
	s.mu.RLock()
	entry, ok := s.data[key]
	s.mu.RUnlock()
	if !ok {
		return -2
	}
	if entry.ExpiresAt.IsZero() {
		return -1
	}
	rem := int(time.Until(entry.ExpiresAt).Seconds())
	if rem < 0 {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return -2
	}
	return rem
}

// startExpirationCleaner periodically removes expired keys.
func (s *Store) startExpirationCleaner() {
	ticker := time.NewTicker(time.Second)
	for now := range ticker.C {
		s.mu.Lock()
		for k, e := range s.data {
			if !e.ExpiresAt.IsZero() && now.After(e.ExpiresAt) {
				delete(s.data, k)
			}
		}
		s.mu.Unlock()
	}
}
