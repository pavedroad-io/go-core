// Credit to github.com/ORBAT/krater/unsafe_writer.go

package logger

import (
	"sync"
	"sync/atomic"
	"syscall"
)

// ZapWriter is an io.Writer that writes messages to Kafka, ignoring any responses
// The AsyncProducer passed to NewZapWriter must have:
//    Config.Return.Successes == false
//    Config.Return.Errors == false
type ZapWriter struct {
	cfg       ProducerConfiguration
	kp        KafkaProducer
	closed    int32          // nonzero if the writer has started closing. Must be accessed atomically
	pendingWg sync.WaitGroup // WaitGroup for pending messages
	closeMut  sync.Mutex
}

// NewZapWriter returns a kafka io.writer instance
func NewZapWriter(cfg ProducerConfiguration, kp KafkaProducer) *ZapWriter {
	zw := &ZapWriter{cfg: cfg, kp: kp}
	return zw
}

// Sync does nothing for now
func (zw *ZapWriter) Sync() error {
	return nil
}

// Write writes byte slices to Kafka ignoring error responses. (Thread-safe.)
// Write might block if the Input() channel of the underlying AsyncProducer is full.
func (zw *ZapWriter) Write(msg []byte) (int, error) {
	if zw.Closed() {
		return 0, syscall.EINVAL
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
