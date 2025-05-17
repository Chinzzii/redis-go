package protocol

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Chinzzii/redis-go/internal/store"
)

// QueuedCmd holds a command queued in a MULTI block.
type QueuedCmd struct {
	Cmd  string
	Args []string
}

// HandleConnection reads lines from conn, parses commands, and invokes store methods.
func HandleConnection(conn net.Conn, s *store.Store) {
	defer conn.Close()
	reader := bufio.NewScanner(conn)

	inTx := false
	var queue []QueuedCmd

	write := func(msg string) {
		fmt.Fprintln(conn, msg)
	}

	for reader.Scan() {
		line := reader.Text()
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		cmd := strings.ToUpper(parts[0])
		args := parts[1:]

		// Queue commands if in transaction (except control ones)
		if inTx && cmd != "EXEC" && cmd != "DISCARD" && cmd != "MULTI" {
			queue = append(queue, QueuedCmd{Cmd: cmd, Args: args})
			write("QUEUED")
			continue
		}

		switch cmd {
		case "SET":
			if len(args) < 2 {
				write("(error) ERR wrong number of arguments for SET")
				continue
			}
			write(s.Set(args[0], strings.Join(args[1:], " ")))
		case "GET":
			if len(args) != 1 {
				write("(error) ERR wrong number of arguments for GET")
				continue
			}
			if val, ok := s.Get(args[0]); ok {
				write(val)
			} else {
				write("(nil)")
			}
		case "DEL":
			if len(args) != 1 {
				write("(error) ERR wrong number of arguments for DEL")
				continue
			}
			write(fmt.Sprintf("(integer) %d", s.Del(args[0])))
		case "EXPIRE":
			if len(args) != 2 {
				write("(error) ERR wrong number of arguments for EXPIRE")
				continue
			}
			sec, err := strconv.Atoi(args[1])
			if err != nil {
				write("(error) ERR invalid expire time")
				continue
			}
			write(fmt.Sprintf("(integer) %d", s.Expire(args[0], sec)))
		case "TTL":
			if len(args) != 1 {
				write("(error) ERR wrong number of arguments for TTL")
				continue
			}
			write(fmt.Sprintf("(integer) %d", s.TTL(args[0])))
		case "SUBSCRIBE":
			if len(args) != 1 {
				write("(error) ERR wrong number of arguments for SUBSCRIBE")
				continue
			}
			write(s.Subscribe(args[0], conn))
		case "UNSUBSCRIBE":
			if len(args) != 1 {
				write("(error) ERR wrong number of arguments for UNSUBSCRIBE")
				continue
			}
			write(s.Unsubscribe(args[0], conn))
		case "PUBLISH":
			if len(args) < 2 {
				write("(error) ERR wrong number of arguments for PUBLISH")
				continue
			}
			cnt := s.Publish(args[0], strings.Join(args[1:], " "))
			write(fmt.Sprintf("(integer) %d", cnt))
		case "MULTI":
			if inTx {
				write("(error) ERR MULTI calls cannot be nested")
			} else {
				inTx = true
				queue = nil
				write("OK")
			}
		case "DISCARD":
			if !inTx {
				write("(error) ERR DISCARD without MULTI")
			} else {
				inTx = false
				queue = nil
				write("OK")
			}
		case "EXEC":
			if !inTx {
				write("(error) ERR EXEC without MULTI")
				continue
			}
			inTx = false
			for _, qc := range queue {
				// Re-dispatch each queued command
				switch qc.Cmd {
				case "SET":
					write(s.Set(qc.Args[0], strings.Join(qc.Args[1:], " ")))
				case "GET":
					if v, ok := s.Get(qc.Args[0]); ok {
						write(v)
					} else {
						write("(nil)")
					}
				case "DEL":
					write(fmt.Sprintf("(integer) %d", s.Del(qc.Args[0])))
				case "EXPIRE":
					if sec, err := strconv.Atoi(qc.Args[1]); err == nil {
						write(fmt.Sprintf("(integer) %d", s.Expire(qc.Args[0], sec)))
					} else {
						write("(error) ERR invalid expire time")
					}
				case "TTL":
					write(fmt.Sprintf("(integer) %d", s.TTL(qc.Args[0])))
				case "PUBLISH":
					cnt := s.Publish(qc.Args[0], strings.Join(qc.Args[1:], " "))
					write(fmt.Sprintf("(integer) %d", cnt))
				default:
					write(fmt.Sprintf("(error) ERR unknown command '%s'", qc.Cmd))
				}
			}
			queue = nil
		case "QUIT":
			write("OK")
			return
		default:
			write(fmt.Sprintf("(error) ERR unknown command '%s'", cmd))
		}
	}

	// Cleanup pub/sub subscriptions
	s.CleanupConnSubscriptions(conn)
}
