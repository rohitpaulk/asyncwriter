package asyncwriter

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

type AsyncWriter struct {
	acceptedWritesCount atomic.Int32
	buffer              chan string
	flushCondition      sync.Cond
	flushError          error
	flushedWritesCount  atomic.Int32
	isClosed            bool
	writer              io.Writer
}

func New(w io.Writer) *AsyncWriter {
	return NewWithSize(w, 1000)
}

func NewWithSize(w io.Writer, bufferSize int) *AsyncWriter {
	writer := &AsyncWriter{
		acceptedWritesCount: atomic.Int32{},
		buffer:              make(chan string, bufferSize),
		flushCondition:      sync.Cond{L: &sync.Mutex{}},
		flushError:          nil,
		flushedWritesCount:  atomic.Int32{},
		writer:              w,
	}

	go writer.runFlushLoop()
	return writer
}

func (w *AsyncWriter) Flush() error {
	w.flushCondition.L.Lock()
	defer w.flushCondition.L.Unlock()

	if w.flushError != nil {
		return w.flushError
	}

	currentAcceptedWritesCount := w.acceptedWritesCount.Load()

	if w.flushedWritesCount.Load() == currentAcceptedWritesCount {
		return w.flushError
	}

	for w.flushedWritesCount.Load() < currentAcceptedWritesCount {
		w.flushCondition.Wait()

		if w.flushError != nil {
			return w.flushError
		}
	}

	return w.flushError
}

func (w *AsyncWriter) Write(b []byte) (int, error) {
	// If we've encountered a flush error, let's return immediately
	if w.flushError != nil {
		return 0, w.flushError
	}

	if w.isClosed {
		return 0, errors.New("writer is closed")
	}

	w.buffer <- string(b)
	w.acceptedWritesCount.Add(1)

	return len(b), nil
}

func (w *AsyncWriter) runFlushLoop() {
	defer w.flushCondition.Signal() // Ensure we notify Flush() if we exit w/ an error

	for b := range w.buffer {
		currentBufferLength := len(w.buffer)

		bytesToFlush := bytes.NewBuffer([]byte{})
		bytesToFlush.WriteString(b)

		// Drain buffer based on current buffer length. Any future writes will be buffered.
		for i := 0; i < currentBufferLength; i++ {
			bytesToFlush.WriteString(<-w.buffer)
		}

		_, err := io.Copy(w.writer, bytesToFlush)
		if err != nil {
			w.flushError = err
			return
		}

		w.flushCondition.L.Lock()
		w.flushedWritesCount.Add(int32(currentBufferLength + 1))
		w.flushCondition.Signal()
		w.flushCondition.L.Unlock()
	}
}

func (w *AsyncWriter) Close() error {
	w.isClosed = true
	close(w.buffer)

	w.Flush()

	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
