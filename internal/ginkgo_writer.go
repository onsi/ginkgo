package internal

import (
	"bytes"
	"io"
	"sync"
)

type ginkgoWriterInterface interface {
	Truncate()
	DumpOut()
}

type GinkgoWriter struct {
	buffer         *bytes.Buffer
	outWriter      io.Writer
	lock           *sync.Mutex
	directToStdout bool
}

func NewGinkgoWriter(outWriter io.Writer) *GinkgoWriter {
	return &GinkgoWriter{
		buffer:         &bytes.Buffer{},
		lock:           &sync.Mutex{},
		outWriter:      outWriter,
		directToStdout: true,
	}
}

func (w *GinkgoWriter) SetDirectToStdout(directToStdout bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.directToStdout = directToStdout
}

func (w *GinkgoWriter) Write(b []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.directToStdout {
		return w.outWriter.Write(b)
	} else {
		return w.buffer.Write(b)
	}
}

func (w *GinkgoWriter) Truncate() {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buffer.Truncate(0)
}

func (w *GinkgoWriter) DumpOut() {
	w.lock.Lock()
	defer w.lock.Unlock()
	if !w.directToStdout {
		w.buffer.WriteTo(w.outWriter)
	}
}
