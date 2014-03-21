package ginkgo

import (
	"bytes"
	"io"
	"sync"
)

type ginkgoWriterInterface interface {
	Truncate()
	DumpOut()
}

type ginkgoWriter struct {
	buffer         *bytes.Buffer
	outWriter      io.Writer
	lock           *sync.Mutex
	directToStdout bool
}

func newGinkgoWriter(outWriter io.Writer) *ginkgoWriter {
	return &ginkgoWriter{
		buffer:         &bytes.Buffer{},
		lock:           &sync.Mutex{},
		outWriter:      outWriter,
		directToStdout: true,
	}
}

func (w *ginkgoWriter) setDirectToStdout(directToStdout bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.directToStdout = directToStdout
}

func (w *ginkgoWriter) Write(b []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.directToStdout {
		return w.outWriter.Write(b)
	} else {
		return w.buffer.Write(b)
	}
}

func (w *ginkgoWriter) Truncate() {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buffer.Truncate(0)
}

func (w *ginkgoWriter) DumpOut() {
	w.lock.Lock()
	defer w.lock.Unlock()
	if !w.directToStdout {
		w.buffer.WriteTo(w.outWriter)
	}
}
