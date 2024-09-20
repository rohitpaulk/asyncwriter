package asyncwriter

import (
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
	for {
		select {
		// There's 1 or more messages available in the buffer
		case b := <-w.buffer:
			if b == nil {
				// w.Buffer is closed
				return
			}

			currentBufferLength := len(w.buffer)

			bytesToFlush := make([]byte, currentBufferLength)
			bytesToFlush = append(bytesToFlush, b...)

			for i := 0; i < currentBufferLength; i++ {
				bytesToFlush = append(bytesToFlush, <-w.buffer...)
			}

			_, err := w.writer.Write(bytesToFlush)
			if err != nil {
				log.Printf("error writing to writer: %v", err)
			}
		}
	}
}
