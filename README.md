# redis-go

A Redis-inspired, in-memory key-value store written in Go.  
Supports concurrent clients, basic commands (SET, GET, DEL, EXPIRE, TTL), Pub/Sub, transactions (MULTI/EXEC/DISCARD), and durable persistence via RDB-style snapshots and an Append-Only File (AOF).

[Medium Article](https://medium.com/@Chinzzii/creating-a-redis-like-in-memory-key-value-store-in-go-6f52894238c2)


## 🚀 Features

- **TCP server** handling multiple clients concurrently via goroutines  
- **Basic commands**:  
  - `SET key value`  
  - `GET key`  
  - `DEL key`  
  - `EXPIRE key seconds`  
  - `TTL key`  
- **Pub/Sub**:  
  - `PUBLISH channel message`  
  - `SUBSCRIBE channel`  
  - `UNSUBSCRIBE channel`  
- **Transactions**:  
  - `MULTI`, `EXEC`, `DISCARD`  
  - Queued command execution  
- **Persistence**:  
  - **RDB-style snapshots** every configurable interval (default 60 s)  
  - **AOF** (Append-Only File) logging of every write  
  - Recovery on startup: loads AOF (preferred) or snapshot  


## 📦 Project Structure
```bash
redis-go/
├── cmd/
│   └── redis-server/
│       └── main.go         # Entrypoint: starts server
├── internal/
│   ├── store/
│   │   ├── store.go        # Core in-memory map, Set/Get/Del
│   │   ├── expiration.go   # EXPIRE, TTL, background cleaner
│   │   ├── pubsub.go       # PUBLISH/SUBSCRIBE/UNSUBSCRIBE
│   │   ├── snapshot.go     # RDB snapshot save/load + StartSnapshotting
│   │   └── aof.go          # AOF replay on startup
│   └── protocol/
│       └── handler.go      # TCP cmd parser & dispatch, transactions
├── pkg/
│   └── util/
│       └── escape.go       # Newline escape/unescape for AOF
├── examples/
│   └── telnet_example.sh   # Example interaction via telnet
├── go.mod
└── README.md
```


## 🛠 Prerequisites

- Go 1.18+  
- (Optional) Docker & Docker CLI for containerized deployment  
- A TCP client for testing: `telnet`, `nc` (netcat), or similar  


## ⚙️ Building & Running

Build and run with Docker:
```bash
docker-compose up -d

# Exec into your running container:
docker exec -it redis-go sh     

# Now you can connect
nc localhost 6379
```


## ✅ Testing / Usage Examples

#### Basic Commands and TTL:
```bash
SET foo bar
OK

GET foo
bar

EXPIRE foo 5
(integer) 1

TTL foo
(integer) 5

# wait >5 seconds
GET foo
(nil)
```

#### Delete and numeric responses:
```bash
SET count 100
OK

GET count
100

DEL count
(integer) 1

DEL count
(integer) 0

GET count
(nil)
```

#### Transactions (MULTI/EXEC):
```bash
MULTI
OK

SET a 1
QUEUED

SET b 2
QUEUED

GET a
QUEUED

EXEC
OK
OK
1
```

#### Publish/Subscribe:
Open two terminal sessions. In the first, subscribe to a channel:
```bash
# Terminal 1:
$ nc localhost 6379
SUBSCRIBE news
Subscribed to news
```
Now in the second terminal, publish a message to that channel:
```bash
# Terminal 2:
$ nc localhost 6379
PUBLISH news Hello World
(integer) 1
```
The publisher gets (integer) 1 indicating one subscriber received the message. Over in Terminal 1, the subscriber sees:
```bash
message news Hello World
```
The subscriber can subscribe to multiple channels (one at a time with our implementation) and will receive messages for each. To stop receiving, unsubscribe:
```bash
# Terminal 1:
UNSUBSCRIBE news
Unsubscribed from news
```
After this, further publishes to "news" will not reach Terminal 1.

#### Persistence:
If you run some commands, then stop the server and restart it, the data should persist:
- On first run, do e.g. SET persistent 42.
- Stop the server (Ctrl+C).
- Restart the server (go run main.go again).
- Then GET persistent should still return 42 thanks to the AOF log (or snapshot) being loaded on startup.

You should see an appendonly.aof file growing with each write command, and a dump.rdb (snapshot) being updated every 60 seconds. This confirms that data is being persisted to disk. On restart, the code loaded the AOF and restored the state.


## 🤝 Contributing
1. Fork it
2. Create feature branch (git checkout -b feature/YourFeature)
3. Commit your changes (git commit -m "Add YourFeature")
4. Push (git push origin feature/YourFeature)
5. Open a Pull Request
