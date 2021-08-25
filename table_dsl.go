package ginkgo

import (
	"fmt"
	"reflect"

	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/types"
)

/*
DescribeTable describes a table-driven test.

For example:

    DescribeTable("a simple table",
        func(x int, y int, expected bool) {
            Ω(x > y).Should(Equal(expected))
        },
        Entry("x > y", 1, 0, true),
        Entry("x == y", 0, 0, false),
        Entry("x < y", 0, 1, false),
    )

The first argument to `DescribeTable` is a string description.
The second argument is a function that will be run for each table entry.  Your assertions go here - the function is equivalent to a Ginkgo It.
The subsequent arguments must be of type `TableEntry`.  We recommend using the `Entry` convenience constructors.

The `Entry` constructor takes a string description followed by an arbitrary set of parameters.  These parameters are passed into your function.

Under the hood, `DescribeTable` simply generates a new Ginkgo `Describe`.  Each `Entry` is turned into an `It` within the `Describe`.

It's important to understand that the `Describe`s and `It`s are generated at evaluation time (i.e. when Ginkgo constructs the tree of tests and before the tests run).

Individual Entries can be focused (with FEntry) or marked pending (with PEntry or XEntry).  In addition, the entire table can be focused or marked pending with FDescribeTable and PDescribeTable/XDescribeTable.

A description function can be passed to Entry in place of the description. The function is then fed with the entry parameters to generate the description of the It corresponding to that particular Entry.

For example:

	describe := func(desc string) func(int, int, bool) string {
		return func(x, y int, expected bool) string {
			return fmt.Sprintf("%s x=%d y=%d expected:%t", desc, x, y, expected)
		}
	}

	DescribeTable("a simple table",
		func(x int, y int, expected bool) {
			Ω(x > y).Should(Equal(expected))
		},
		Entry(describe("x > y"), 1, 0, true),
		Entry(describe("x == y"), 0, 0, false),
		Entry(describe("x < y"), 0, 1, false),
	)
*/
func DescribeTable(description string, itBody interface{}, entries ...TableEntry) bool {
	describeTable(description, itBody, entries, false, false)
	return true
}

/*
You can focus a table with `FDescribeTable`.  This is equivalent to `FDescribe`.
*/
func FDescribeTable(description string, itBody interface{}, entries ...TableEntry) bool {
	describeTable(description, itBody, entries, true, false)
	return true
}

/*
You can mark a table as pending with `PDescribeTable`.  This is equivalent to `PDescribe`.
*/
func PDescribeTable(description string, itBody interface{}, entries ...TableEntry) bool {
	describeTable(description, itBody, entries, false, true)
	return true
}

/*
You can mark a table as pending with `XDescribeTable`.  This is equivalent to `XDescribe`.
*/
func XDescribeTable(description string, itBody interface{}, entries ...TableEntry) bool {
	describeTable(description, itBody, entries, false, true)
	return true
}

func describeTable(description string, itBody interface{}, entries []TableEntry, markedFocus bool, markedPending bool) {
	itBodyValue := reflect.ValueOf(itBody)
	if itBodyValue.Kind() != reflect.Func {
		panic(fmt.Sprintf("DescribeTable expects a function, got %#v", itBody))
	}

	args := []interface{}{
		func() {
			for _, entry := range entries {
				entry.generateIt(itBodyValue)
			}
		},
		types.NewCodeLocation(2),
	}
	if markedFocus {
		args = append(args, internal.Focus)
	}
	if markedPending {
		args = append(args, internal.Pending)
	}

	pushNode(internal.NewNode(deprecationTracker, types.NodeTypeContainer, description, args...))
}

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
			exitIfErr(fmt.Errorf("The describe function should return only a value, returned %d", len(res)))
		}
		if res[0].Kind() != reflect.String {
			exitIfErr(fmt.Errorf("The describe function should return a string, returned %#v", res[0]))
		}
		description = res[0].String()
	default:
		exitIfErr(fmt.Errorf("Description can either be a string or a function, got %#v", descriptionValue))
	}

	args := []interface{}{t.codeLocation}

	if t.Pending {
		args = append(args, internal.Pending)
	} else {
		values := castParameters(itBody, t.Parameters)
		body := func() {
			itBody.Call(values)
		}
		args = append(args, body)
	}
	if t.Focused {
		args = append(args, internal.Focus)
	}

	pushNode(internal.NewNode(deprecationTracker, types.NodeTypeIt, description, args...))
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

The first argument is a required description (this becomes the content of the generated Ginkgo `It`).
Subsequent parameters are saved off and sent to the callback passed in to `DescribeTable`.

Each Entry ends up generating an individual Ginkgo It.
*/
func Entry(description interface{}, parameters ...interface{}) TableEntry {
	return TableEntry{
		Description:  description,
		Parameters:   parameters,
		Pending:      false,
		Focused:      false,
		codeLocation: types.NewCodeLocation(1),
	}
}

/*
You can focus a particular entry with FEntry.  This is equivalent to FIt.
*/
func FEntry(description interface{}, parameters ...interface{}) TableEntry {
	return TableEntry{
		Description:  description,
		Parameters:   parameters,
		Pending:      false,
		Focused:      true,
		codeLocation: types.NewCodeLocation(1),
	}
}

/*
You can mark a particular entry as pending with PEntry.  This is equivalent to PIt.
*/
func PEntry(description interface{}, parameters ...interface{}) TableEntry {
	return TableEntry{
		Description:  description,
		Parameters:   parameters,
		Pending:      true,
		Focused:      false,
		codeLocation: types.NewCodeLocation(1),
	}
}

/*
You can mark a particular entry as pending with XEntry.  This is equivalent to XIt.
*/
func XEntry(description interface{}, parameters ...interface{}) TableEntry {
	return TableEntry{
		Description:  description,
		Parameters:   parameters,
		Pending:      true,
		Focused:      false,
		codeLocation: types.NewCodeLocation(1),
	}
}
