package asyncwriter

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"
)

func TestAsyncWriter(t *testing.T) {
	t.Run("Write and Close", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := New(buf)

		data := []byte("Hello, AsyncWriter!")
		n, err := writer.Write(data)
		if err != nil {
			t.Fatalf("Write error: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected to write %d bytes, but wrote %d", len(data), n)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Close error: %v", err)
		}

		time.Sleep(10 * time.Millisecond) // Allow time for flushing

		if buf.String() != string(data) {
			t.Errorf("Expected %q, but got %q", string(data), buf.String())
		}
	})

	t.Run("Multiple Writes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := New(buf)

		data1 := []byte("Hello, ")
		data2 := []byte("AsyncWriter!")

		n, err := writer.Write(data1)
		if err != nil {
			t.Fatalf("Write error: %v", err)
		}
		if n != len(data1) {
			t.Errorf("Expected to write %d bytes, but wrote %d", len(data1), n)
		}

		n, err = writer.Write(data2)
		if err != nil {
			t.Fatalf("Write error: %v", err)
		}
		if n != len(data2) {
			t.Errorf("Expected to write %d bytes, but wrote %d", len(data2), n)
		}

		writer.Close()
		time.Sleep(10 * time.Millisecond) // Allow time for flushing

		expected := "Hello, AsyncWriter!"
		if buf.String() != expected {
			t.Errorf("Expected %q, but got %q", expected, buf.String())
		}
	})

	t.Run("Concurrent Writes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := New(buf)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				data := []byte("Write ")
				writer.Write(data)
			}(i)
		}

		wg.Wait()
		writer.Close()
		time.Sleep(10 * time.Millisecond) // Allow time for flushing

		if len(buf.Bytes()) != 600 { // "Write " is 6 bytes, 100 times
			t.Errorf("Expected 600 bytes, but got %d", len(buf.Bytes()))
		}
	})

	t.Run("NewWithSize", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := NewWithSize(buf, 10)

		for i := 0; i < 20; i++ {
			writer.Write([]byte("a"))
		}

		writer.Close()
		time.Sleep(10 * time.Millisecond) // Allow time for flushing

		if buf.String() != "aaaaaaaaaaaaaaaaaaaa" {
			t.Errorf("Expected 20 'a's, but got %q", buf.String())
		}
	})

	t.Run("Close with non-Closer writer", func(t *testing.T) {
		writer := New(io.Discard)
		err := writer.Close()
		if err != nil {
			t.Errorf("Expected no error on Close, but got: %v", err)
		}
	})

	t.Run("Flush", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := New(buf)

		data := []byte("Hello, AsyncWriter!")
		_, err := writer.Write(data)
		if err != nil {
			t.Fatalf("Write error: %v", err)
		}

		err = writer.Flush()
		if err != nil {
			t.Fatalf("Flush error: %v", err)
		}

		if buf.String() != string(data) {
			t.Errorf("Expected %q after Flush, but got %q", string(data), buf.String())
		}
	})

	t.Run("Flush Empty Buffer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := New(buf)

		err := writer.Flush()
		if err != nil {
			t.Fatalf("Flush error on empty buffer: %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("Expected empty buffer after Flush, but got %d bytes", buf.Len())
		}
	})

	t.Run("Flush After Error", func(t *testing.T) {
		errWriter := &errorWriter{err: io.ErrShortWrite}
		writer := New(errWriter)

		n, err := writer.Write([]byte("Hello"))
		if n != 5 {
			t.Errorf("Expected to write 5 bytes, but wrote %d", n)
		}
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		time.Sleep(10 * time.Millisecond) // Allow time for error to be set

		err = writer.Flush()
		if err != io.ErrShortWrite {
			t.Errorf("Expected ErrShortWrite, but got: %v", err)
		}
	})

	t.Run("Write After Close", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := New(buf)

		err := writer.Close()
		if err != nil {
			t.Fatalf("Close error: %v", err)
		}

		_, err = writer.Write([]byte("Hello"))
		if err == nil {
			t.Error("Expected error when writing to closed AsyncWriter, but got nil")
		}
	})
}

// errorWriter is a test helper that always returns an error on Write
type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}
