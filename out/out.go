// Package out provides methods for manipulating the standard outputs of vpnmasq at runtime
package out

import (
	"os"
)

// MustInterceptFile is the same as InterceptFile, except it panics on error.
func MustInterceptFile(f **os.File) (reader *os.File, original *os.File) {
	reader, original, err := InterceptFile(f)
	if err != nil {
		panic(err)
	}
	return
}

// InterceptFile replaces the given *os.File with a pipe, returning the original *os.File and an
// io.Reader for reading subsequent writes to the replaced file.
func InterceptFile(f **os.File) (reader *os.File, original *os.File, err error) {
	var w *os.File
	reader, w, err = os.Pipe()
	if err != nil {
		return
	}

	original = *f
	*f = w
	return
}
