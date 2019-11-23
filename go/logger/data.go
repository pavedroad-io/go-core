package logger

import (
	"bytes"

	uuid "github.com/satori/go.uuid"
)

// prefix the message with cloudevents ID field
func prefixID(msg []byte) ([]byte, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	parts := [][]byte{[]byte("{\"id\":\""), []byte(id.String()), []byte("\",")}
	json := bytes.Join(parts, []byte(""))
	return bytes.Replace(msg, []byte("{"), json, 1), nil
}
