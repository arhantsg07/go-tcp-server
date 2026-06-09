package protocol

import (
	"encoding/json"
)

// decoder.go
func Decode(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}
