`AsyncWriter` is an alternative to Go's [`bufio.Writer`](https://pkg.go.dev/bufio#Writer) that only buffers writes when a write is already in progress.

### Example

With `bufio.Writer`:

```
Write 1 -> buffered
Write 2 -> buffered
Write 3 -> buffered
Write 4 -> buffered
          -> flush (1, 2, 3, 4) since buffer size is reached
```

With `AsyncWriter`:

```
Write 1 -> buffered
          -> start flushing (1) immediately
Write 2 -> buffered
Write 3 -> buffered
Write 4 -> buffered
        -> flush (2, 3, 4) now since (1) is flushed
```

### Usage

```go
import "github.com/rohitpaulk/asyncwriter"

writer := asyncwriter.New(slowWriter)

// "Hello, world!" is written to slowWriter immediately
writer.Write([]byte("Hello, world!"))

// These writes are buffered and written to slowWriter as fast as it'll accept writes
for i := 0; i < 1000; i++ {
	writer.Write([]byte(fmt.Sprintf("Hello, async writer! %d", i)))
}

// Waits for all current writes to be flushed
writer.Flush()

// Close writer and the underlying slowWriter
writer.Close()
```
