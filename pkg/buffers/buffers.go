package buffers

import (
	"bytes"
	"sync"
)

// Buffer is a goroutine safe bytes.Buffer
type Buffer struct {
	buf   bytes.Buffer
	mutex sync.Mutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed. It returns
// the number of bytes written.
func (b *Buffer) Write(d []byte) (n int, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.buf.Write(d)
}

// String returns the contents of the unread portion of the buffer
// as a string.  If the Buffer is a nil pointer, it returns "<nil>".
func (b *Buffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.buf.String()
}
