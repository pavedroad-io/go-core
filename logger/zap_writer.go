// Based on github.com/ORBAT/krater/unsafe_writer.go

package logger

import (
	"errors"
	"sync"
	"sync/atomic"
	"syscall"
)

// ZapWriter is an io.Writer that writes messages to Kafka, ignoring any responses
// The AsyncProducer passed to newZapWriter must have:
//    Config.Return.Successes == false
//    Config.Return.Errors == false
type ZapWriter struct {
	kp        *KafkaProducer
	closed    int32          // Nonzero if closing started, must be accessed atomically
	pendingWg sync.WaitGroup // WaitGroup for pending messages
	closeMut  sync.Mutex
}

// newZapWriter returns a kafka io.writer instance
func newZapWriter(
	kpcfg ProducerConfiguration,
	cecfg CloudEventsConfiguration) (*ZapWriter, error) {

	// create an async producer
	kp, err := newKafkaProducer(kpcfg, cecfg)
	if err != nil {
		return nil, err
	}

	zw := &ZapWriter{kp: kp}
	return zw, nil
}

// Sync does nothing for now
func (zw *ZapWriter) Sync() error {
	return nil
}

// Write writes byte slices to Kafka ignoring error responses (Thread-safe)
// Write might block if the Input() channel of the underlying AsyncProducer is full
func (zw *ZapWriter) Write(msg []byte) (int, error) {
	if zw.Closed() {
		return 0, syscall.EINVAL
	}

	if zw.kp.producer == nil {
		return 0, errors.New("No producer defined")
	}

	zw.pendingWg.Add(1)
	defer zw.pendingWg.Done()

	err := zw.kp.sendMessage(msg)
	return len(msg), err
}

// Closed returns true if ZapWriter is closed, false otherwise (Thread-safe)
func (zw *ZapWriter) Closed() bool {
	return atomic.LoadInt32(&zw.closed) != 0
}

// Close must be called when the writer is no longer needed (Thread-safe)
func (zw *ZapWriter) Close() (err error) {
	zw.closeMut.Lock()
	defer zw.closeMut.Unlock()

	if zw.Closed() {
		return syscall.EINVAL
	}

	atomic.StoreInt32(&zw.closed, 1)

	zw.pendingWg.Wait()
	return nil
}
