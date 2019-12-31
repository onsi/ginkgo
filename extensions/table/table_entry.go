package table

import (
	"reflect"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/types"
)

/*
TableEntry represents an entry in a table test.  You generally use the `Entry` constructor.
*/
type TableEntry struct {
	Description string
	Parameters  []interface{}
	Flag        types.FlagType

	// TODO: what if this isn't set?
	codeLocation types.CodeLocation
}

func (t TableEntry) generateIt(itBody reflect.Value) {
	if t.Flag == types.FlagTypePending {
		ginkgo.ExplicitItNode(t.Description, func() {}, types.FlagTypePending, t.codeLocation, 0)
		return
	}

	values := make([]reflect.Value, len(t.Parameters))
	iBodyType := itBody.Type()
	for i, param := range t.Parameters {
		if param == nil {
			inType := iBodyType.In(i)
			values[i] = reflect.Zero(inType)
		} else {
			values[i] = reflect.ValueOf(param)
		}
	}

	ginkgo.ExplicitItNode(t.Description, func() {
		itBody.Call(values)
	}, t.Flag, t.codeLocation, 0)
}

/*
Entry constructs a TableEntry.

The first argument is a required description (this becomes the content of the generated Ginkgo `It`).
Subsequent parameters are saved off and sent to the callback passed in to `DescribeTable`.

Each Entry ends up generating an individual Ginkgo It.
*/
func Entry(description string, parameters ...interface{}) TableEntry {
	return TableEntry{description, parameters, types.FlagTypeNone, codelocation.New(1)}
}

/*
You can focus a particular entry with FEntry.  This is equivalent to FIt.
*/
func FEntry(description string, parameters ...interface{}) TableEntry {
	return TableEntry{description, parameters, types.FlagTypeFocused, codelocation.New(1)}
}

/*
You can mark a particular entry as pending with PEntry.  This is equivalent to PIt.
*/
func PEntry(description string, parameters ...interface{}) TableEntry {
	return TableEntry{description, parameters, types.FlagTypePending, codelocation.New(1)}
}

/*
You can mark a particular entry as pending with XEntry.  This is equivalent to XIt.
*/
func XEntry(description string, parameters ...interface{}) TableEntry {
	return TableEntry{description, parameters, types.FlagTypePending, codelocation.New(1)}
}
