package logger

import (
	"bytes"

	uuid "github.com/satori/go.uuid"
)

type ceKey string

// Keys for cloudevents fields
const (
	ceIdKey      ceKey = "id"
	ceLevelKey         = "subject"
	ceMessageKey       = "data"
	ceSourceKey        = "source"
	ceTimeKey          = "time"
	ceTypeKey          = "type"
	ceVersionKey       = "specversion"
)

// Fixed field values for cloudevents
var ceFields = Fields{
	ceSourceKey:  "http://github.com/pavedroad-io/core/go/logger",
	ceVersionKey: "1.0",
	ceTypeKey:    "io.pavedroad.cloudevents.log",
}

// Add the cloudevents ID field to the message (UUID)
func ceAddIdField(msg []byte) ([]byte, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	parts := [][]byte{[]byte("{\"id\":\""), []byte(id.String()), []byte("\",")}
	json := bytes.Join(parts, []byte(""))
	return bytes.Replace(msg, []byte("{"), json, 1), nil
}
