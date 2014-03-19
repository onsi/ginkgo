package testrunner

import (
	"bytes"
	"log"
	"strings"
	"sync"
)

func init() {
	log.SetFlags(0)
}

type logWriter struct {
	prefix string
	buffer *bytes.Buffer
	lock   *sync.Mutex
}

func newLogWriter(prefix string) *logWriter {
	return &logWriter{
		prefix: prefix,
		buffer: &bytes.Buffer{},
		lock:   &sync.Mutex{},
	}
}

func (w *logWriter) Write(data []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.buffer.Write(data)
	contents := w.buffer.String()

	lines := strings.Split(contents, "\n")
	for _, line := range lines[0 : len(lines)-1] {
		log.Printf("%s %s\n", w.prefix, line)
	}

	w.buffer.Reset()
	w.buffer.Write([]byte(lines[len(lines)-1]))
	return len(data), nil
}

func (w *logWriter) Close() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.buffer.Len() > 0 {
		log.Printf("%s %s\n", w.prefix, w.buffer.String())
	}

	return nil
}
