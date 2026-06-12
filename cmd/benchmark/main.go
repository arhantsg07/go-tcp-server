// cmd/benchmark/main.go
//
// A concurrent load-testing tool for the TCP server.
// It spawns N goroutines each acting as a real client: connects, registers,
// floods CHAT messages, then disconnects cleanly.
//
// Usage:
//
//	go run ./cmd/benchmark --clients=100 --messages=500 --addr=127.0.0.1:8080

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/arhantsg07/go-tcp-server/internal/protocol"
)

func main() {
	clients := flag.Int("clients", 50, "number of concurrent clients")
	messages := flag.Int("messages", 200, "CHAT messages each client sends")
	addr := flag.String("addr", "127.0.0.1:8080", "server address")
	flag.Parse()

	var (
		totalSent   atomic.Int64
		totalFailed atomic.Int64
		wg          sync.WaitGroup
	)

	fmt.Printf("Launching %d clients → %s  (%d messages each)\n\n",
		*clients, *addr, *messages)

	start := time.Now()

	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			n, err := runClient(*addr, id, *messages)
			totalSent.Add(int64(n))
			if err != nil {
				totalFailed.Add(int64(*messages - n))
				log.Printf("[client %04d] error after %d msgs: %v", id, n, err)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	sent := totalSent.Load()
	failed := totalFailed.Load()
	throughput := float64(sent) / elapsed.Seconds()
	avgLatency := time.Duration(0)
	if sent > 0 {
		avgLatency = time.Duration(elapsed.Nanoseconds() / sent)
	}

	fmt.Printf("── Benchmark Results ─────────────────────────────\n")
	fmt.Printf("  Clients         : %d\n", *clients)
	fmt.Printf("  Messages/client : %d\n", *messages)
	fmt.Printf("  Total sent      : %d\n", sent)
	fmt.Printf("  Errors          : %d\n", failed)
	fmt.Printf("  Duration        : %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Throughput      : %.0f msg/s\n", throughput)
	fmt.Printf("  Avg send latency: %s per message\n", avgLatency)
	fmt.Printf("─────────────────────────────────────────────────\n")
}

// runClient simulates one client for the lifetime of the benchmark.
// Returns the number of CHAT messages successfully sent.
func runClient(addr string, id, messages int) (int, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return 0, fmt.Errorf("dial: %w", err)
	}
	defer func() {
		// Best-effort clean disconnect.
		sendMsg(conn, protocol.Message{
			Type:      protocol.MessageTypeDisconnect,
			From:      fmt.Sprintf("bench-%04d", id),
			Timestamp: time.Now(),
		})
		conn.Close()
	}()

	username := fmt.Sprintf("bench-%04d", id)

	// Drain goroutine: continuously read and discard incoming messages
	// (welcome, join-ack, broadcasts from other clients) so the server's
	// write buffer never fills up and stalls.
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			// discard incoming broadcasts and server messages
		}
		_ = scanner.Err() // connection closed is expected on teardown
	}()

	// Send JOIN and wait a brief moment for the server to register us.
	if err := sendMsg(conn, protocol.Message{
		Type:      protocol.MessageTypeJoin,
		From:      username,
		Timestamp: time.Now(),
	}); err != nil {
		return 0, fmt.Errorf("join: %w", err)
	}
	time.Sleep(20 * time.Millisecond)

	// Flood CHAT messages.
	sent := 0
	for i := 0; i < messages; i++ {
		if err := sendMsg(conn, protocol.Message{
			Type:      protocol.MessageTypeChat,
			From:      username,
			Text:      fmt.Sprintf("m%d", i),
			Timestamp: time.Now(),
		}); err != nil {
			return sent, fmt.Errorf("send msg %d: %w", i, err)
		}
		sent++
	}
	return sent, nil
}

func sendMsg(conn net.Conn, msg protocol.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.Write(data)
	return err
}
