package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arhantsg07/go-tcp-server/internal/server"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	srv := server.New("127.0.0.1", "8080")

	// Start HTTP stats server on a separate port (non-blocking).
	go srv.StartStatsServer("8081")

	// Derive a context that cancels on SIGINT or SIGTERM so Start() shuts down cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Periodically log live stats to stdout.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Printf("[stats] %s", srv.Metrics().Snapshot())
			}
		}
	}()

	if err := srv.Start(ctx); err != nil {
		log.Fatal("server error:", err)
	}

	log.Printf("[final] %s", srv.Metrics().Snapshot())
}
