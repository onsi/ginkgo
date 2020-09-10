package internal

import (
	"bytes"
	"io"
	"sync"
)

type WriterInterface interface {
	io.Writer

	Truncate()
	Bytes() []byte
}

type Writer struct {
	buffer    *bytes.Buffer
	outWriter io.Writer
	lock      *sync.Mutex
	stream    bool
}

func NewWriter(outWriter io.Writer) *Writer {
	return &Writer{
		buffer:    &bytes.Buffer{},
		lock:      &sync.Mutex{},
		outWriter: outWriter,
		stream:    true,
	}
}

func (w *Writer) SetStream(stream bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.stream = stream
}

func (w *Writer) Write(b []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.stream {
		return w.outWriter.Write(b)
	}
	return w.buffer.Write(b)
}

func (w *Writer) Truncate() {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buffer.Reset()
}

func (w *Writer) Bytes() []byte {
	w.lock.Lock()
	defer w.lock.Unlock()
	b := w.buffer.Bytes()
	copied := make([]byte, len(b))
	copy(copied, b)
	return copied
}
