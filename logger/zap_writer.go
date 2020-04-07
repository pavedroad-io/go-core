// Inspired by github.com/ORBAT/krater/unsafe_writer.go

package logger

import (
	"errors"
	"sync"
	"sync/atomic"
	"syscall"
)

// ZapKafkaWriter is a zap WriteSyncer (io.Writer) that writes messages to Kafka
type ZapKafkaWriter struct {
	kp        *KafkaProducer
	closed    int32          // Nonzero if closing, must access atomically
	pendingWg sync.WaitGroup // WaitGroup for pending messages
	closeMut  sync.Mutex
}

// newZapKafkaWriter returns a kafka io.writer instance
func newZapKafkaWriter(
	kpcfg ProducerConfiguration,
	cecfg CloudEventsConfiguration) (*ZapKafkaWriter, error) {

	// create an async producer
	kp, err := newKafkaProducer(kpcfg, cecfg)
	if err != nil {
		return nil, err
	}

	zw := &ZapKafkaWriter{kp: kp}
	return zw, nil
}

// Sync satisfies zapcore.WriteSyncer interface, zapcore.AddSync works as well
func (zw *ZapKafkaWriter) Sync() error {
	// TODO can writer be closed here?
	return nil
}

// Write sends byte slices to Kafka ignoring error responses (Thread-safe)
// Write might block if the Input() channel of the AsyncProducer is full
func (zw *ZapKafkaWriter) Write(msg []byte) (int, error) {
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

// Closed returns true if the writer is closed, false otherwise (Thread-safe)
func (zw *ZapKafkaWriter) Closed() bool {
	return atomic.LoadInt32(&zw.closed) != 0
}

// Close must be called when the writer is no longer needed (Thread-safe)
func (zw *ZapKafkaWriter) Close() (err error) {
	zw.closeMut.Lock()
	defer zw.closeMut.Unlock()

	if zw.Closed() {
		return syscall.EINVAL
	}

	atomic.StoreInt32(&zw.closed, 1)

	zw.pendingWg.Wait()
	return nil
}
