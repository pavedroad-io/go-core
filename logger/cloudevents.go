package logger

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/gofrs/uuid"
)

// ceSetIDType provides cloudevents id field type
type ceSetIDType string

// Types of cloudevents id fields
const (
	ceHMAC    ceSetIDType = "hmac"  // for de-duplication
	ceUUID    ceSetIDType = "uuid"  // completely unique
	ceIncrID  ceSetIDType = "incr"  // incremental
	ceFuncID  ceSetIDType = "func"  // set by WithFields or FilterFunc
	ceFixedID ceSetIDType = "fixed" // use config ID
)

// ceSSetubjectType provides cloudevents subject field type
type ceSetSubjectType string

// Types of cloudevents subject fields
const (
	ceLevelSubject ceSetSubjectType = "level" // insert log level
	ceFixedSubject ceSetSubjectType = "fixed" // use config subject
	ceSkipSubject  ceSetSubjectType = "skip"  // skip subject field
)

// CloudEventsConfiguration provides cloudevents configuration type
type CloudEventsConfiguration struct {
	ID          string
	Source      string
	SpecVersion string
	Type        string
	Subject     string
	SetID       ceSetIDType
	SetSubject  ceSetSubjectType
}

// Keys for cloudevents fields, values must be non-empty strings
const (
	ceIDKey           = "id"              // Required - unique per producer
	ceSourceKey       = "source"          // Required - URI-reference
	ceSpecVersionKey  = "specversion"     // Required - current spec is "1.0"
	ceTypeKey         = "type"            // Required - reverse-DNS name prefix
	ceDataContentType = "datacontenttype" // Optional - adheres to RFC2046
	ceDataSchemaKey   = "dataschema"      // Optional - URI-reference
	ceSubjectKey      = "subject"         // Optional - possibly pass log level
	ceTimeKey         = "time"            // Optional - adheres to RFC3339
	ceDataKey         = "data"            // Optional - no specific format
)

// KafkaProducer gets the cloudevents fields to add to the message
func ceGetFields(config CloudEventsConfiguration) LogFields {
	// Possibly override default field values for cloudevents
	ceFields := LogFields{}
	defaultCfg := DefaultCloudEventsCfg()

	// cloudevents fields must contain non-empty strings
	if config.SetID == ceFixedID {
		if config.ID == "" {
			ceFields[ceIDKey] = defaultCfg.ID
		} else {
			ceFields[ceIDKey] = config.ID
		}
	}
	if config.Source == "" {
		ceFields[ceSourceKey] = defaultCfg.Source
	} else {
		ceFields[ceSourceKey] = config.Source
	}
	if config.SpecVersion == "" {
		ceFields[ceSpecVersionKey] = defaultCfg.SpecVersion
	} else {
		ceFields[ceSpecVersionKey] = config.SpecVersion
	}
	if config.Type == "" {
		ceFields[ceTypeKey] = defaultCfg.Type
	} else {
		ceFields[ceTypeKey] = config.Type
	}
	if config.SetSubject == ceFixedSubject {
		if config.Subject == "" {
			ceFields[ceSubjectKey] = defaultCfg.Subject
		} else {
			ceFields[ceSubjectKey] = config.Subject
		}
	}
	return ceFields
}

func incrementalID() func() string {
	i := 0
	return func() string {
		i++
		return fmt.Sprintf("%d", i)
	}
}

// KafkaProducer adds the cloudevents id field to the message
func (kp *KafkaProducer) ceAddFields(config CloudEventsConfiguration,
	msgMap map[string]interface{}) error {
	// Other cloudevents fields could be added here based on config
	switch config.SetID {
	case ceFuncID:
		// set by FilterFn or WithFields
		break
	case ceIncrID:
		msgMap[string(ceIDKey)] = incrementalID()
	case ceUUID:
		id, err := uuid.NewV4() // RFC4112
		if err != nil {
			return err
		}
		msgMap[string(ceIDKey)] = id
	case ceFixedID:
		// set by configuration only for console and logfiles
		fallthrough
	case ceHMAC:
		fallthrough
	default:
		key := []byte("pavedroad-secret")
		h := hmac.New(sha256.New, key)
		h.Write([]byte(msgMap[string(ceDataKey)].(string)))
		id := base64.StdEncoding.EncodeToString(h.Sum(nil))
		msgMap[string(ceIDKey)] = id
	}
	return nil
}
