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

		writer.Write(data1)
		writer.Write(data2)

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
}
