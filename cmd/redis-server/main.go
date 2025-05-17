package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Chinzzii/redis-go/internal/protocol"
	"github.com/Chinzzii/redis-go/internal/store"
)

const (
	DefaultAddr      = ":6379"
	SnapshotFile     = "dump.rdb"
	AppendOnlyFile   = "appendonly.aof"
	SnapshotInterval = 60 * time.Second
)

func main() {
	// Initialize store (loads AOF or snapshot, starts expiration cleaner)
	s := store.NewStore(AppendOnlyFile)
	defer s.Close()

	// Periodic RDB‑style snapshots
	go s.StartSnapshotting(SnapshotInterval, SnapshotFile)

	// Start listening for clients
	ln, err := net.Listen("tcp", DefaultAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listening on %s: %v\n", DefaultAddr, err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Server listening on", DefaultAddr)

	// Accept connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Accept error: %v\n", err)
			continue
		}
		go protocol.HandleConnection(conn, s)
	}
}
