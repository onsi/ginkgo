package table

import (
	"reflect"

	"github.com/onsi/ginkgo"
)

/*
TableEntry represents an entry in a table test.  You generally use the `Entry` constructor.
*/
type TableEntry struct {
	Description string
	Parameters  []interface{}
	Pending     bool
	Focused     bool
}

func (t TableEntry) generateIt(itBody reflect.Value) {
	if t.Pending {
		ginkgo.PIt(t.Description)
		return
	}

	values := make([]reflect.Value, 0)
	itBodyType := itBody.Type()

	for i, param := range t.Parameters {
		var value reflect.Value

		if param == nil {
			inType := itBodyType.In(i)
			value = reflect.Zero(inType)
		} else {
			value = reflect.ValueOf(param)
		}

		values = append(values, value)
	}

	var (
		body    interface{}
		timeout []float64
	)

	if itBodyType.NumIn() >= 1 && itBodyType.In(0).Kind() == reflect.Chan &&
		itBodyType.In(0).Elem().Kind() == reflect.Interface {

		lenValues := len(values)
		if lenValues > 0 && values[lenValues-1].Kind() == reflect.Float64 {
			timeout = append(timeout, values[lenValues-1].Interface().(float64))
			values = values[:lenValues-1]
		}

		body = func(done chan<- interface{}) {
			values = append([]reflect.Value{reflect.ValueOf(done)}, values...)
			itBody.Call(values)
		}
	} else {
		body = func() {
			itBody.Call(values)
		}
	}

	if t.Focused {
		ginkgo.FIt(t.Description, body, timeout...)
	} else {
		ginkgo.It(t.Description, body, timeout...)
	}
}

/*
Entry constructs a TableEntry.

The first argument is a required description (this becomes the content of the generated Ginkgo `It`).
Subsequent parameters are saved off and sent to the callback passed in to `DescribeTable`.

Each Entry ends up generating an individual Ginkgo It.
*/
func Entry(description string, parameters ...interface{}) TableEntry {
	return TableEntry{description, parameters, false, false}
}

/*
You can focus a particular entry with FEntry.  This is equivalent to FIt.
*/
func FEntry(description string, parameters ...interface{}) TableEntry {
	return TableEntry{description, parameters, false, true}
}

/*
You can mark a particular entry as pending with PEntry.  This is equivalent to PIt.
*/
func PEntry(description string, parameters ...interface{}) TableEntry {
	return TableEntry{description, parameters, true, false}
}

/*
You can mark a particular entry as pending with XEntry.  This is equivalent to XIt.
*/
func XEntry(description string, parameters ...interface{}) TableEntry {
	return TableEntry{description, parameters, true, false}
}
