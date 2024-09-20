package asyncwriter

import (
	"bytes"
	"io"
	"log"
)

type AsyncWriter struct {
	buffer chan []byte
	writer io.Writer
}

func New(w io.Writer) *AsyncWriter {
	return NewWithSize(w, 1000)
}

func NewWithSize(w io.Writer, bufferSize int) *AsyncWriter {
	writer := &AsyncWriter{
		writer: w,
		buffer: make(chan []byte, bufferSize),
	}

	go writer.runFlushLoop()
	return writer
}

func (w *AsyncWriter) Write(b []byte) (int, error) {
	w.buffer <- b
	return len(b), nil
}

func (w *AsyncWriter) runFlushLoop() {
	for b := range w.buffer {
		currentBufferLength := len(w.buffer)
		bytesToFlush := bytes.NewBuffer([]byte{})
		bytesToFlush.Write(b)

		// Drain buffer based on current buffer length. Any future writes will be buffered.
		for i := 0; i < currentBufferLength; i++ {
			bytesToFlush.Write(<-w.buffer)
		}

		_, err := w.writer.Write(bytesToFlush.Bytes())
		if err != nil {
			log.Printf("error writing to writer: %v", err)
		}
	}
}

func (w *AsyncWriter) Close() error {
	close(w.buffer)

	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
