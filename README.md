# Go TCP Chat Server

A concurrent TCP chat server built from scratch in Go, featuring a custom JSON-framed protocol, per-connection rate limiting, live metrics over HTTP, and graceful shutdown.

---

## Features

| Feature | Detail |
|---|---|
| **Concurrent connections** | Each client runs in its own goroutine; tested with 200+ simultaneous connections |
| **Custom binary-safe protocol** | Newline-delimited JSON frames — no third-party dependencies |
| **Broadcast messaging** | `CHAT` messages fan-out to all registered clients |
| **Private messaging** | `PRIVATE` messages delivered point-to-point by username |
| **Username registry** | `JOIN` handshake with duplicate-username rejection |
| **Token-bucket rate limiting** | Burst of 10 messages, sustained at 5 msg/s per connection |
| **Lock-free metrics** | All counters use `sync/atomic` — zero-contention stats collection |
| **Per-connection heartbeat** | 30-second keep-alive; dead connections closed automatically |
| **Middleware chain** | `LoggingMiddleware` wraps every read/write with type + latency logging |
| **HTTP stats API** | Live metrics, client list, and health probe at `:8081` |
| **Graceful shutdown** | `SIGINT`/`SIGTERM` closes the listener and waits for in-flight handlers |
| **Identity enforcement** | Server overwrites the `from` field — clients cannot spoof identity |
| **Concurrent-write safety** | Per-connection `sync.Mutex` prevents garbled frames when the heartbeat goroutine and a broadcast race to write |

---

## Project Structure

```
cmd/
  server/main.go        — entrypoint: signal handling, stats logging
  benchmark/main.go     — concurrent load-testing tool

internal/
  protocol/
    message.go          — Message struct + MessageType constants
    encoder.go          — JSON marshal → newline-delimited frame
    decoder.go          — JSON unmarshal from raw bytes

  server/
    server.go           — Server struct: Register / Unregister / Broadcast / SendPrivate
    connection.go       — Connection type, full lifecycle (HandleConnection + route)
    handler.go          — ReadMessage / SendMessage + byte tracking
    middleware.go       — LoggingMiddleware (MiddlewareFunc pattern)
    heartbeat.go        — per-connection keep-alive goroutine
    metrics.go          — lock-free atomic Metrics + Stats snapshot
    ratelimit.go        — token-bucket RateLimiter (burst + sustained rate)
    stats_server.go     — HTTP /stats /clients /health
    tcp.go              — context-aware Start() + graceful shutdown
```

---

## Running

```bash
# Start the server
go run ./cmd/server

# In another terminal, connect with netcat (send raw JSON)
echo '{"type":"JOIN","from":"alice"}' | nc 127.0.0.1 8080

# Or use any JSON-capable TCP client
```

### Checking live stats

```bash
curl http://localhost:8081/stats    # full JSON metrics
curl http://localhost:8081/clients  # online user list
curl http://localhost:8081/health   # liveness probe
```

---

## Benchmark

```bash
# Start the server first
go run ./cmd/server

# Run the load test (separate terminal)
go run ./cmd/benchmark --clients=100 --messages=500

# Example output:
# Launching 100 clients → 127.0.0.1:8080  (500 messages each)
#
# ── Benchmark Results ─────────────────────────────
#   Clients         : 100
#   Messages/client : 500
#   Total sent      : 50000
#   Errors          : 0
#   Duration        : 3.2s
#   Throughput      : ~15,600 msg/s
#   Avg send latency: 64µs per message
# ─────────────────────────────────────────────────
```

Flags: `--clients`, `--messages`, `--addr`

---

## Protocol Reference

All messages are UTF-8 JSON objects terminated by a single `\n`.

| Type | Direction | Fields used | Description |
|---|---|---|---|
| `JOIN` | client → server | `from` | Register with a username |
| `CHAT` | client → server | `text` | Broadcast to all other clients |
| `PRIVATE` | client → server | `to`, `text` | Send to a specific user |
| `HEARTBEAT` | client → server | — | Echo the server's heartbeat back |
| `LIST` | client → server | — | Request a comma-separated list of online users |
| `DISCONNECT` | client → server | — | Clean disconnect with server acknowledgement |
| `HEARTBEAT` | server → client | — | Sent every 30 s; close connection if undeliverable |
| `ERROR` | server → client | `text` | Protocol violation or routing error |

### Example session

```json
// server → client (on connect)
{"type":"JOIN","from":"Server","timestamp":"...","text":"Connected. Send a JOIN message with your username to continue."}

// client → server
{"type":"JOIN","from":"alice"}

// server → client (join confirmed)
{"type":"JOIN","from":"Server","to":"alice","timestamp":"...","text":"Welcome, alice! 1 user(s) online."}

// client → server (broadcast)
{"type":"CHAT","from":"alice","text":"hello everyone"}

// client → server (private)
{"type":"PRIVATE","from":"alice","to":"bob","text":"hey bob"}

// client → server (list users)
{"type":"LIST","from":"alice"}

// server → client (list response)
{"type":"LIST","from":"Server","to":"alice","timestamp":"...","text":"alice, bob"}
```

---

## Design Notes

- **`sync.RWMutex` on the client registry**: `Broadcast` holds only a read lock while collecting targets, then releases it before the actual writes — avoiding lock contention during fan-out.
- **Lock-free metrics**: `sync/atomic` counters are used for all stats so no mutex is needed on hot paths.
- **Identity enforcement**: The `from` field is overwritten server-side in `route()` — clients can never impersonate another user.
- **Write serialisation**: `Connection.writeMu` ensures the heartbeat goroutine and a concurrent broadcast never interleave partial writes on the same socket.
