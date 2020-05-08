package out

import (
	"fmt"
	"io"
)

// ListenCopy wraps src, copying data to copy.
func ListenCopy(src io.ReadCloser, copy io.Writer) io.ReadCloser {
	return &listener{
		src:  src,
		copy: copy,
	}
}

// listener copies reads to a writer as they are read.
type listener struct {
	src  io.ReadCloser
	copy io.Writer
}

// Read implements io.Reader. It copies all read data to the writer, returning errors.
func (l *listener) Read(buff []byte) (int, error) {
	n, err := l.src.Read(buff)
	w, wErr := l.copy.Write(buff)

	if w == n && wErr == nil {
		// No write error
		return n, err
	}

	// Write error
	if wErr == nil {
		wErr = io.ErrShortWrite
	}

	return n, fmt.Errorf("Write error copying stream: %w", wErr)
}

// Close calls close on the source reader.
func (l *listener) Close() error {
	return l.src.Close()
}
