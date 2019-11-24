package logger

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	uuid "github.com/satori/go.uuid"
)

type ceKey string

// Keys for cloudevents fields
const (
	ceIDKey      ceKey = "id"
	ceLevelKey         = "subject"
	ceMessageKey       = "data"
	ceSourceKey        = "source"
	ceTimeKey          = "time"
	ceTypeKey          = "type"
	ceVersionKey       = "specversion"
)

type ceIDType int

const (
	TypeHMAC ceIDType = iota // for de-dupliation
	TypeUUID                 // completely unique
)

// Fixed field values for cloudevents
var ceFields = Fields{
	ceSourceKey:  "http://github.com/pavedroad-io/core/go/logger",
	ceVersionKey: "1.0",
	ceTypeKey:    "io.pavedroad.cloudevents.log",
}

// Add the cloudevents ID field to the message (UUID)
// Other cloudevents fields could be added here based on config
func (kp *kafkaProducer) ceAddFields(msgMap map[string]interface{}) error {
	switch kp.config.CloudeventsID {
	case TypeUUID:
		id, err := uuid.NewV4() // RFC4112
		if err != nil {
			return err
		}
		msgMap[string(ceIDKey)] = id
	case TypeHMAC:
		fallthrough
	default:
		key := []byte("pavedroad-secret")
		h := hmac.New(sha256.New, key)
		h.Write([]byte(msgMap[string(ceMessageKey)].(string)))
		id := base64.StdEncoding.EncodeToString(h.Sum(nil))
		msgMap[string(ceIDKey)] = id
	}
	return nil
}
