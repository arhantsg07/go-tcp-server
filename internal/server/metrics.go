package server

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Metrics uses lock-free atomic counters for zero-contention stats collection.
// All fields are safe for concurrent reads and writes without a mutex.
type Metrics struct {
	TotalConnections  atomic.Int64
	ActiveConnections atomic.Int64
	MessagesReceived  atomic.Int64
	MessagesSent      atomic.Int64
	BytesReceived     atomic.Int64
	BytesSent         atomic.Int64
	BroadcastsSent    atomic.Int64
	PrivatesSent      atomic.Int64
	startTime         time.Time
}

func newMetrics() *Metrics {
	return &Metrics{startTime: time.Now()}
}

// Stats is a point-in-time snapshot of Metrics, safe to read without locks.
type Stats struct {
	TotalConnections  int64   `json:"total_connections"`
	ActiveConnections int64   `json:"active_connections"`
	MessagesReceived  int64   `json:"messages_received"`
	MessagesSent      int64   `json:"messages_sent"`
	BytesReceived     int64   `json:"bytes_received"`
	BytesSent         int64   `json:"bytes_sent"`
	BroadcastsSent    int64   `json:"broadcasts_sent"`
	PrivatesSent      int64   `json:"privates_sent"`
	UptimeSeconds     float64 `json:"uptime_seconds"`
	MessagesPerSecond float64 `json:"messages_per_second"`
}

// Snapshot returns a consistent read of all counters at this moment.
func (m *Metrics) Snapshot() Stats {
	uptime := time.Since(m.startTime).Seconds()
	if uptime < 1 {
		uptime = 1
	}
	received := m.MessagesReceived.Load()
	return Stats{
		TotalConnections:  m.TotalConnections.Load(),
		ActiveConnections: m.ActiveConnections.Load(),
		MessagesReceived:  received,
		MessagesSent:      m.MessagesSent.Load(),
		BytesReceived:     m.BytesReceived.Load(),
		BytesSent:         m.BytesSent.Load(),
		BroadcastsSent:    m.BroadcastsSent.Load(),
		PrivatesSent:      m.PrivatesSent.Load(),
		UptimeSeconds:     uptime,
		MessagesPerSecond: float64(received) / uptime,
	}
}

func (s Stats) String() string {
	return fmt.Sprintf(
		"uptime=%.1fs  active=%d  total=%d  recv=%d  sent=%d  throughput=%.1f msg/s  bytes_in=%d  bytes_out=%d",
		s.UptimeSeconds, s.ActiveConnections, s.TotalConnections,
		s.MessagesReceived, s.MessagesSent, s.MessagesPerSecond,
		s.BytesReceived, s.BytesSent,
	)
}
