package out

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// MustInterceptBuffer is the same as InterceptBuffer, except it panics on error.
func MustInterceptBuffer(f **os.File) *Buffer {
	buff, err := InterceptBuffer(f)
	if err != nil {
		panic(err)
	}
	return buff
}

// InterceptBuffer returns a new *Buffer, transperantly intercepting writes to the given *os.File and buffering them.
// Closing the buffer will only close the file's replacement.
func InterceptBuffer(f **os.File) (*Buffer, error) {
	// Intercept writes to f
	reader, original, err := InterceptFile(f)
	if err != nil {
		return nil, err
	}

	// Copy writes to the original.
	src := ListenCopy(reader, original)

	// Create a buffer reading from src
	return NewBuffer(src), nil
}

func newBuffer(r io.ReadCloser) *bufferWriter {
	closed := make(chan error)
	w := &bufferWriter{
		Buffer: &Buffer{
			closed: closed,
			r:      r,
		},
	}

	go func(w *bufferWriter) {
		_, err := io.Copy(w, r)
		closed <- err
	}(w)

	return w
}

// bufferWriter hides the write function from users.
type bufferWriter struct {
	*Buffer
}

// mustBuff creates the buffer if it does not exist.
// lock should be held.
func (b *bufferWriter) mustBuff() {
	if b.buff == nil {
		b.buff = new(bytes.Buffer)
	}
}

// Write implements io.Writer.
func (b *bufferWriter) Write(buff []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.dst != nil {
		return b.dst.Write(buff)
	}

	b.mustBuff()
	return b.buff.Write(buff)
}

// NewBuffer returns a new *Buffer reading from r.
// Close() calls r.Close(), and will not return until calls to r.Read() return.
func NewBuffer(r io.ReadCloser) *Buffer {
	return newBuffer(r).Buffer
}

// Buffer implements io.Writer and Piper, containing a buffer for holding writes until the first call to Pipe().
// It is thread safe.
type Buffer struct {
	buff *bytes.Buffer
	// r is set to nil after close.
	r     io.Closer
	dst   io.Writer
	mutex sync.Mutex
	// closed is closed after close.
	closed <-chan error
}

// Read implements io.Reader. It reads from the currently buffered data. If Pipe() has been called it throws a nil pointer exception.
func (b *Buffer) Read(buff []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buff.Read(buff)
}

// Pipe implements Piper. It writes all buffered writes to out,
// returning any errors encountered, and then pipes all subsequent writes to out.
//
// If the Pipe is unsuccessfull, it can be retired, or Buffer() can be called after Close() to return the unwritten
// portion of the Buffer, along with any subsequent writes.
func (b *Buffer) Pipe(out io.Writer) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.buff != nil {
		// Buffer is non-nil, write buffer to writer, then perform switch
		l := b.buff.Len()
		n, err := b.buff.WriteTo(out)
		if n != int64(l) {
			if err != nil {
				return err
			}
			return io.ErrShortWrite
		}

		// Release empty and no longer required buffer
		b.buff = nil
	}

	b.dst = out
	return nil
}

// Close implements io.Closer. It closes the source reader, and waits for remaining data to be written.
func (b *Buffer) Close() error {
	b.mutex.Lock()
	if b.r == nil {
		b.mutex.Unlock()
		return os.ErrClosed
	}

	err := b.r.Close()
	b.r = nil
	b.mutex.Unlock()
	copyErr := <-b.closed

	if copyErr != nil {
		return copyErr
	}
	return err
}

// Remaining returns the underlying buffer after a call to Close().
func (b *Buffer) Remaining() *bytes.Buffer {
	err := <-b.closed
	if err != nil {
		panic("Close() must be called before Remaining()")
	}
	return b.buff
}
