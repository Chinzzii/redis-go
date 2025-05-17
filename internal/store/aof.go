package store

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Chinzzii/redis-go/pkg/util"
)

// loadAOF replays write commands from the AOF file.
func (s *Store) loadAOF(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		cmd := strings.ToUpper(parts[0])
		switch cmd {
		case "SET":
			if len(parts) == 3 {
				key := parts[1]
				val := util.UnescapeNewlines(parts[2])
				s.data[key] = Entry{Value: val}
			}
		case "DEL":
			if len(parts) >= 2 {
				delete(s.data, parts[1])
			}
		case "EXPIRE":
			if len(parts) == 3 {
				if sec, err := strconv.Atoi(parts[2]); err == nil {
					if e, ok := s.data[parts[1]]; ok {
						e.ExpiresAt = time.Now().Add(time.Duration(sec) * time.Second)
						s.data[parts[1]] = e
					}
				}
			}
		}
	}
	return scanner.Err()
}
