package logger

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"

	"github.com/gofrs/uuid"
)

// ceSetIDType provides cloudevents id field type
type ceSetIDType string

// Types of cloudevents id fields
const (
	ceHMAC   ceSetIDType = "hmac" // message signature
	ceUUID   ceSetIDType = "uuid" // completely unique
	ceIncrID ceSetIDType = "incr" // incremental
	ceFuncID ceSetIDType = "func" // set by WithFields or FilterFunc
)

// CloudEventsConfiguration provides cloudevents configuration type
type CloudEventsConfiguration struct {
	SetID           ceSetIDType
	HmacKey         string
	Source          string
	SpecVersion     string
	Type            string
	SetSubjectLevel bool
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

type incrementalFn func() string

// CloudEvents provides the cloudevents object type
type CloudEvents struct {
	fields           LogFields
	genIncrementalID incrementalFn
	hmacHash         hash.Hash
}

// incrementalID returns function that returns IDs starting with zero
func incrementalID() func() string {
	var i uint64
	return func() string {
		i++
		return fmt.Sprintf("%020d", i)
	}
}

// newCloudEvents returns a cloudevents instance
func newCloudEvents(config CloudEventsConfiguration) *CloudEvents {
	// Possibly override default field values for cloudevents
	ceFields := LogFields{}
	defaultCfg := DefaultCloudEventsCfg()

	// cloudevents fields must contain non-empty strings
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
	cloudEvents := CloudEvents{
		fields: ceFields,
	}

	switch config.SetID {
	case ceIncrID:
		cloudEvents.genIncrementalID = incrementalID()
	case ceHMAC:
		key := []byte(config.HmacKey)
		cloudEvents.hmacHash = hmac.New(sha256.New, key)
	}
	return &cloudEvents
}

// ceGetID returns the cloudevents id field for the message
func (ce *CloudEvents) ceGetID(config CloudEventsConfiguration,
	msgMap map[string]interface{}) (string, error) {

	switch config.SetID {
	case ceFuncID:
		// set when using FilterFn or WithFields to supply id
		return "", nil
	case ceIncrID:
		return ce.genIncrementalID(), nil
	case ceUUID:
		id, err := uuid.NewV4() // RFC4112
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s", id), nil
	case ceHMAC:
		fallthrough
	default:
		ce.hmacHash.Write([]byte(msgMap[string(ceDataKey)].(string)))
		id := base64.StdEncoding.EncodeToString(ce.hmacHash.Sum(nil))
		return id, nil
	}
}

// ceAddFields adds the cloudevents id field to the message
func (ce *CloudEvents) ceAddFields(config CloudEventsConfiguration,
	msgMap map[string]interface{}) error {
	// Other cloudevents fields could be added here based on config

	id, err := ce.ceGetID(config, msgMap)
	if err != nil {
		return err
	}
	msgMap[string(ceIDKey)] = id
	return nil
}
