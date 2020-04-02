package logger

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/gofrs/uuid"
)

// ceIDType provides cloudevents id field type
type ceIDType string

// Types of cloudevents id fields
const (
	HMAC ceIDType = "hmac" // for de-duplication
	UUID          = "uuid" // completely unique
)

// CloudEventsConfiguration provides cloudevents configuration type
type CloudEventsConfiguration struct {
	Source      string
	SpecVersion string
	Type        string
	ID          ceIDType
}

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

// KafkaProducer gets the cloudevents fields to add to the message
func ceGetFields(config CloudEventsConfiguration) Fields {
	// Field key values for cloudevents
	var ceFields = Fields{}
	if config.Source != "" {
		ceFields[ceSourceKey] = config.Source
	}
	if config.SpecVersion != "" {
		ceFields[ceVersionKey] = config.SpecVersion
	}
	if config.Type != "" {
		ceFields[ceTypeKey] = config.Type
	}
	return ceFields
}

// KafkaProducer adds the cloudevents id field to the message
func (kp *KafkaProducer) ceAddFields(config CloudEventsConfiguration,
	msgMap map[string]interface{}) error {
	// Other cloudevents fields could be added here based on config
	switch config.ID {
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
