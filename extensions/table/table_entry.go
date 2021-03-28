package table

import (
	"fmt"
	"github.com/onsi/ginkgo/internal/codelocation"
	"reflect"

	"github.com/onsi/ginkgo/internal/global"
	"github.com/onsi/ginkgo/types"
)

/*
TableEntry represents an entry in a table test.  You generally use the `Entry` constructor.
*/
type TableEntry struct {
	Description  interface{}
	Parameters   []interface{}
	Pending      bool
	Focused      bool
	codeLocation types.CodeLocation
}

func (t TableEntry) generateIt(itBody reflect.Value) {
	var description string
	descriptionValue := reflect.ValueOf(t.Description)
	switch descriptionValue.Kind() {
	case reflect.String:
		description = descriptionValue.String()
	case reflect.Func:
		values := castParameters(descriptionValue, t.Parameters)
		res := descriptionValue.Call(values)
		if len(res) != 1 {
			panic(fmt.Sprintf("The describe function should return only a value, returned %d", len(res)))
		}
		if res[0].Kind() != reflect.String {
			panic(fmt.Sprintf("The describe function should return a string, returned %#v", res[0]))
		}
		description = res[0].String()
	default:
		panic(fmt.Sprintf("Description can either be a string or a function, got %#v", descriptionValue))
	}

	if t.Pending {
		global.Suite.PushItNode(description, func() {}, types.FlagTypePending, t.codeLocation, 0)
		return
	}

	values := castParameters(itBody, t.Parameters)
	body := func() {
		itBody.Call(values)
	}

	if t.Focused {
		global.Suite.PushItNode(description, body, types.FlagTypeFocused, t.codeLocation, global.DefaultTimeout)
	} else {
		global.Suite.PushItNode(description, body, types.FlagTypeNone, t.codeLocation, global.DefaultTimeout)
	}
}

func castParameters(function reflect.Value, parameters []interface{}) []reflect.Value {
	res := make([]reflect.Value, len(parameters))
	funcType := function.Type()
	for i, param := range parameters {
		if param == nil {
			inType := funcType.In(i)
			res[i] = reflect.Zero(inType)
		} else {
			res[i] = reflect.ValueOf(param)
		}
	}
	return res
}

/*
Entry constructs a TableEntry.

When used in `DescribeTable`, the first argument should be a description (this becomes the content of the generated Ginkgo `It`) and all subsequent parameters are saved off and sent to the callback passed in to `DescribeTable`.
This description can either be a string or a function. If the description is a function, the function must accept all of the arguments contained in the entry and must return a string.

When used in `ItTable`, all arguments are saved and sent to the callback method of `ItTable` (and the description of `ItTable` when it is a function).

Each Entry ends up generating an individual Ginkgo It.
*/
func Entry(parameters ...interface{}) TableEntry {
	return entry(false, false, parameters)
}

/*
You can focus a particular entry with FEntry.  This is equivalent to FIt.
*/
func FEntry(parameters ...interface{}) TableEntry {
	return entry(false, true, parameters)
}

/*
You can mark a particular entry as pending with PEntry.  This is equivalent to PIt.
*/
func PEntry(parameters ...interface{}) TableEntry {
	return entry(true, false, parameters)
}

/*
You can mark a particular entry as pending with XEntry.  This is equivalent to XIt.
*/
func XEntry(parameters ...interface{}) TableEntry {
	return entry(true, false, parameters)
}

func entry(pending, focused bool, parameters ...interface{}) TableEntry {
	if len(parameters) == 0 {
		panic("Entry requires at least one parameter")
	}

	p := parameters[0].([]interface{})
	if len(p) == 0 {
		panic("Entry requires at least one parameter")
	}

	return TableEntry{
		p[0],
		p[1:],
		pending,
		focused,
		codelocation.New(1),
	}
}
