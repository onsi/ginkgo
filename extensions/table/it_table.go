package table

/*
ItTable creates a Ginkgo It for every entry provided within the Ginkgo Describe that contains it.

For example:

    Describe("a simple test", func() {
        ItTable(func(x, y int, expected bool) {
            return fmt.Sprintf("should assert that '%d > %d' is %t", x, y, expected)
        },
            func(x, y int, expected bool) {
                Ω(x > y).Should(Equal(expected))
            },
            Entry(1, 0, true),
            Entry(0, 0, false),
            Entry(0, 1, false),
        )
    })

The first argument to `ItTable` can either be a string or a function. If it is a function, the function must accept all the same arguments as the callback function and must return a string.
The second argument is the callback function called in each generated Ginkgo It. This is where you should put your assertions.
All subsequent arguments must be of type `TableEntry`. If you choose to use the `Entry` constructor, you should NOT put a description as the first argument and instead put only the arguments you want sent to the callback function. The description of the `ItTable` will be used for the description of the generated Ginkgo It fields. If you choose to create a `TableEntry` struct yourself, do not set the `Description` field.

Individual Entries can be focused (with FEntry) or marked pending (with PEntry or XEntry).  In addition, the entire table can be focused or marked pending with FItTable and PItTable/XItTable.
*/
func ItTable(description, body interface{}, entries ...TableEntry) bool {
	return true
}

/*
You can focus a table with `FItTable`. This is equivalent to `FIt`.
*/
func FItTable(description, body interface{}, entries ...TableEntry) bool {
	return true
}

/*
You can mark a table as pending with `PItTable`. This is equivalent to `PIt`.
*/
func PItTable(description, body interface{}, entries ...TableEntry) bool {
	return true
}

/*
You can mark a table as pending with `XItTable`. This is equivalent to `XIt`.
*/
func XItTable(description, body interface{}, entries ...TableEntry) bool {
	return true
}
