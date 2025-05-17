package store

import (
	"fmt"
	"net"
)

// Subscribe adds conn to channel’s subscriber list.
func (s *Store) Subscribe(channel string, conn net.Conn) string {
	s.mu.Lock()
	if s.subs[channel] == nil {
		s.subs[channel] = make(map[net.Conn]bool)
	}
	s.subs[channel][conn] = true
	if s.subscribedChans[conn] == nil {
		s.subscribedChans[conn] = make(map[string]bool)
	}
	s.subscribedChans[conn][channel] = true
	s.mu.Unlock()
	return fmt.Sprintf("Subscribed to %s", channel)
}

// Unsubscribe removes conn from a channel.
func (s *Store) Unsubscribe(channel string, conn net.Conn) string {
	s.mu.Lock()
	if subs := s.subs[channel]; subs != nil {
		delete(subs, conn)
		if len(subs) == 0 {
			delete(s.subs, channel)
		}
	}
	if chans := s.subscribedChans[conn]; chans != nil {
		delete(chans, channel)
		if len(chans) == 0 {
			delete(s.subscribedChans, conn)
		}
	}
	s.mu.Unlock()
	return fmt.Sprintf("Unsubscribed from %s", channel)
}

// Publish sends message to all subscribers; returns number of recipients.
func (s *Store) Publish(channel, message string) int {
	s.mu.RLock()
	subs := s.subs[channel]
	conns := make([]net.Conn, 0, len(subs))
	for c := range subs {
		conns = append(conns, c)
	}
	s.mu.RUnlock()
	for _, c := range conns {
		fmt.Fprintf(c, "message %s %s\r\n", channel, message)
	}
	return len(conns)
}

// CleanupConnSubscriptions removes all subscriptions for a connection.
func (s *Store) CleanupConnSubscriptions(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if chans := s.subscribedChans[conn]; chans != nil {
		for ch := range chans {
			delete(s.subs[ch], conn)
			if len(s.subs[ch]) == 0 {
				delete(s.subs, ch)
			}
		}
		delete(s.subscribedChans, conn)
	}
}
