package tmp

import (
	"testing"
)

func TestSomethingLessImportant(t *testing.T) {
	somethingImportant(t, &"hello!")
}

func somethingImportant(t *testing.T, message *string) {
	t.Log("Something important happened in a test: " + *message)
}
