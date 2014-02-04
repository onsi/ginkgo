package tmp

import (
	"testing"
)

type UselessStruct struct {
	ImportantField string
	T              *testing.T
}

var testFunc = func(t *testing.T, arg *string) { }

func TestSomethingImportant(t *testing.T) {
	whatever := &UselessStruct{
		T:            t,
		ImportantField: "twisty maze of passages",
	}
	app := "string value"
	something := &UselessStruct{ImportantField: app}

	t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
	assert.Equal(t, whatever.ImportantField, "SECRET_PASSWORD")
	var foo = func(t *testing.T) {}
	foo()
	testFunc(t, "something")
}
