// from github.com/ORBAT/krater/unsafe_writer.go

package logger

import (
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/Shopify/sarama"
)

// type KeyerFn represents any function that can turn a message into a key for that particular message
type KeyerFn func(*sarama.ProducerMessage) sarama.Encoder

// ZapWriter is an io.Writer that writes messages to Kafka, ignoring any error responses sent by the brokers.
// Parallel calls to Write / ReadFrom are safe.
//
// The AsyncProducer passed to NewZapWriter must have Config.Return.Successes == false and Config.Return.Errors == false
//
// Close() must be called when the writer is no longer needed.
type ZapWriter struct {
	kp        sarama.AsyncProducer
	topic     string
	closed    int32          // nonzero if the writer has started closing. Must be accessed atomically
	pendingWg sync.WaitGroup // WaitGroup for pending messages
	closeMut  sync.Mutex
}

func NewZapWriter(topic string, kp sarama.AsyncProducer) *ZapWriter {
	uw := &ZapWriter{kp: kp, topic: topic}
	return uw
}

// Sync does nothing for now
func (uw *ZapWriter) Sync() error {
	return nil
}

// Write writes byte slices to Kafka without checking for error responses.
// Trying to Write to a closed writer will return syscall.EINVAL. Thread-safe.
//
// Write might block if the Input() channel of the underlying sarama.AsyncProducer is full.
func (uw *ZapWriter) Write(p []byte) (n int, err error) {
	if uw.Closed() {
		return 0, syscall.EINVAL
	}
	// prefix message with id field
	p, err = prefixID(p)
	if err != nil {
		return 0, err
	}

	uw.pendingWg.Add(1)
	defer uw.pendingWg.Done()

	n = len(p)
	msg := &sarama.ProducerMessage{
		Topic: uw.topic,
		Key:   nil,
		Value: sarama.ByteEncoder(p),
	}
	uw.kp.Input() <- msg
	return
}

// Closed returns true if the ZapWriter has been closed, false otherwise. Thread-safe.
func (uw *ZapWriter) Closed() bool {
	return atomic.LoadInt32(&uw.closed) != 0
}

// Close closes the writer.
// If the writer has already been closed, Close will return syscall.EINVAL. Thread-safe.
func (uw *ZapWriter) Close() (err error) {
	uw.closeMut.Lock()
	defer uw.closeMut.Unlock()

	if uw.Closed() {
		return syscall.EINVAL
	}

	atomic.StoreInt32(&uw.closed, 1)

	uw.pendingWg.Wait()
	return nil
}
