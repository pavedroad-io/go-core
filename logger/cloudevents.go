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
	CEHMAC   ceSetIDType = "hmac" // message signature
	CEUUID   ceSetIDType = "uuid" // completely unique
	CEIncrID ceSetIDType = "incr" // incremental
	CEFuncID ceSetIDType = "func" // set by WithFields or FilterFunc
)

// CloudEventsConfiguration provides cloudevents configuration type
type CloudEventsConfiguration struct {
	SetID           ceSetIDType
	HMACKey         string
	Source          string
	SpecVersion     string
	Type            string
	SetSubjectLevel bool
}

// Keys for cloudevents fields, values must be non-empty strings
const (
	CEIDKey           = "id"              // Required - unique per producer
	CESourceKey       = "source"          // Required - URI-reference
	CESpecVersionKey  = "specversion"     // Required - current spec is "1.0"
	CETypeKey         = "type"            // Required - reverse-DNS name prefix
	CEDataContentType = "datacontenttype" // Optional - adheres to RFC2046
	CEDataSchemaKey   = "dataschema"      // Optional - URI-reference
	CESubjectKey      = "subject"         // Optional - possibly pass log level
	CETimeKey         = "time"            // Optional - adheres to RFC3339
	CEDataKey         = "data"            // Optional - no specific format
)

type incrementalFn func() string

// CloudEvents provides the cloudevents object type
type CloudEvents struct {
	config           CloudEventsConfiguration
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
func newCloudEvents(cfg CloudEventsConfiguration) *CloudEvents {
	// Possibly override default field values for cloudevents
	config := DefaultCloudEventsCfg()

	// cloudevents fields must contain non-empty strings
	config.SetID = cfg.SetID
	if cfg.HMACKey != "" {
		config.HMACKey = cfg.HMACKey
	}
	if cfg.Source != "" {
		config.Source = cfg.Source
	}
	if cfg.SpecVersion != "" {
		config.SpecVersion = cfg.SpecVersion
	}
	if cfg.Type != "" {
		config.Type = cfg.Type
	}
	config.SetSubjectLevel = cfg.SetSubjectLevel

	fields := LogFields{}
	fields[CESourceKey] = config.Source
	fields[CESpecVersionKey] = config.SpecVersion
	fields[CETypeKey] = config.Type

	cloudEvents := CloudEvents{
		config: config,
		fields: fields,
	}

	switch config.SetID {
	case CEIncrID:
		cloudEvents.genIncrementalID = incrementalID()
	case CEHMAC:
		fallthrough
	default:
		key := []byte(config.HMACKey)
		cloudEvents.hmacHash = hmac.New(sha256.New, key)
	}
	return &cloudEvents
}

// ceGetID returns the cloudevents id field for the message
func (ce *CloudEvents) ceGetID(msgMap map[string]interface{}) (string, error) {
	switch ce.config.SetID {
	case CEFuncID:
		// set when using FilterFn or WithFields to supply id
		return "", nil
	case CEIncrID:
		return ce.genIncrementalID(), nil
	case CEUUID:
		id, err := uuid.NewV4() // RFC4112
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s", id), nil
	case CEHMAC:
		fallthrough
	default:
		ce.hmacHash.Write([]byte(msgMap[string(CEDataKey)].(string)))
		id := base64.StdEncoding.EncodeToString(ce.hmacHash.Sum(nil))
		return id, nil
	}
}

// ceAddFields adds the cloudevents id field to the message
func (ce *CloudEvents) ceAddFields(msgMap map[string]interface{}) error {
	// Other cloudevents fields could be added here based on config

	id, err := ce.ceGetID(msgMap)
	if err != nil {
		return err
	}
	msgMap[string(CEIDKey)] = id
	return nil
}
