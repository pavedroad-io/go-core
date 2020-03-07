package logger

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	uuid "github.com/satori/go.uuid"
)

// ceKey provides cloudevents key type
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

// ceIDType provides cloudevents id field type
type ceIDType int8


// Types of cloudevents id fields
const (
	HMAC ceIDType = iota // for de-duplication
	UUID                 // completely unique
)

// Fixed field values for cloudevents
var ceFields = Fields{
	ceSourceKey:  "http://github.com/pavedroad-io/core/go/logger",
	ceVersionKey: "1.0",
	ceTypeKey:    "io.pavedroad.cloudevents.log",
}

// KafkaProducer adds the cloudevents id field to the message
func (kp *KafkaProducer) ceAddFields(msgMap map[string]interface{}) error {
	// Other cloudevents fields could be added here based on config
	switch kp.config.CloudeventsID {
	case UUID:
		id, err := uuid.NewV4() // RFC4112
		if err != nil {
			return err
		}
		msgMap[string(ceIDKey)] = id
	case HMAC:
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
