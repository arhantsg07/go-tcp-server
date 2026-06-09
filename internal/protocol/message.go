package protocol

import (
	"time"
)

type MessageType string

const (
	MessageTypeJoin       MessageType = "JOIN"
	MessageTypeChat       MessageType = "CHAT"
	MessageTypePrivate    MessageType = "PRIVATE"
	MessageTypeHeartbeat  MessageType = "HEARTBEAT"
	MessageTypeDisconnect MessageType = "DISCONNECT"
	MessageTypeError      MessageType = "ERROR"
)

type Message struct {
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	From      string      `json:"from"`
	To        string      `json:"to,omitempty"`
	Text      string      `json:"text,omitempty"`
}