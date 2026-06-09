package protocol

import "encoding/json"

func Encode(msg Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	return append(data, '\n'), err // newline = frame delimiter
}
