package asyncwriter

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

type AsyncWriter struct {
	buffer            chan []byte
	flushCondition    sync.Cond
	flushError        error
	flushedItemsCount int
	isClosed          bool
	writer            io.Writer
}

func New(w io.Writer) *AsyncWriter {
	return NewWithSize(w, 1000)
}

func NewWithSize(w io.Writer, bufferSize int) *AsyncWriter {
	writer := &AsyncWriter{
		buffer:            make(chan []byte, bufferSize),
		flushCondition:    sync.Cond{L: &sync.Mutex{}},
		flushError:        nil,
		flushedItemsCount: 0,
		writer:            w,
	}

	go writer.runFlushLoop()
	return writer
}

func (w *AsyncWriter) Flush() error {
	w.flushCondition.L.Lock()
	defer w.flushCondition.L.Unlock()

	currentFlushedItemsCount := w.flushedItemsCount
	currentBufferLength := len(w.buffer)

	if currentBufferLength == 0 {
		return w.flushError
	}

	for w.flushedItemsCount < currentFlushedItemsCount+currentBufferLength {
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

	w.buffer <- b

	return len(b), nil
}

func (w *AsyncWriter) runFlushLoop() {
	defer w.flushCondition.Signal() // Ensure we notify Flush() if we exit w/ an error

	for b := range w.buffer {
		currentBufferLength := len(w.buffer)
		bytesToFlush := bytes.NewBuffer([]byte{})
		bytesToFlush.Write(b)

		// Drain buffer based on current buffer length. Any future writes will be buffered.
		for i := 0; i < currentBufferLength; i++ {
			bytesToFlush.Write(<-w.buffer)
		}

		_, err := io.Copy(w.writer, bytesToFlush)
		if err != nil {
			w.flushError = err
			return
		}

		w.flushCondition.L.Lock()
		w.flushedItemsCount += currentBufferLength + 1 // +1 for the current item
		w.flushCondition.Signal()
		w.flushCondition.L.Unlock()
	}
}

func (w *AsyncWriter) Close() error {
	w.isClosed = true
	close(w.buffer)

	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
