package slogchi

import (
	"bytes"
	"io"
)

type bodyWriter struct {
	body    *bytes.Buffer
	maxSize int
}

// implements io.Writer
func (w *bodyWriter) Write(b []byte) (int, error) {
	length := len(b)

	if w.body.Len()+length > w.maxSize {
		w.body.Truncate(min(w.maxSize, length, w.body.Len()))
		return w.body.Write(b[:min(w.maxSize-w.body.Len(), length)])
	}
	return w.body.Write(b)
}

func newBodyWriter(maxSize int) *bodyWriter {
	return &bodyWriter{
		body:    bytes.NewBufferString(""),
		maxSize: maxSize,
	}
}

type bodyReader struct {
	io.ReadCloser
	body    *bytes.Buffer
	maxSize int
	bytes   int
}

// implements io.Reader
func (r *bodyReader) Read(b []byte) (int, error) {
	n, err := r.ReadCloser.Read(b)
	if r.body != nil && r.body.Len() < r.maxSize {
		if r.body.Len()+n > r.maxSize {
			r.body.Write(b[:min(r.maxSize-r.body.Len(), n)])
		} else {
			r.body.Write(b[:n])
		}
	}
	r.bytes += n
	return n, err
}

func newBodyReader(reader io.ReadCloser, maxSize int, recordBody bool) *bodyReader {
	var body *bytes.Buffer
	if recordBody {
		body = bytes.NewBufferString("")
	}
	return &bodyReader{
		ReadCloser: reader,
		body:       body,
		maxSize:    maxSize,
		bytes:      0,
	}
}
