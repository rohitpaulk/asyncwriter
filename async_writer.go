package asyncwriter

import (
	"io"
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
}

func (w *AsyncWriter) runFlushLoop() {
	for {
		select {
		case b := <-w.buffer:
			_, err := w.writer.Write(b)
			if err != nil {
				log.Printf("error writing to writer: %v", err)
			}
		}
	}
}
