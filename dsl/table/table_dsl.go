/*
Ginkgo isusually dot-imported via:

    import . "github.com/onsi/ginkgo/v2"

however some parts of the DSL may conflict with existing symbols in the user's code.

To mitigate this without losing the brevity of dot-importing Ginkgo the various packages in the
dsl directory provide pieces of the Ginkgo DSL that can be dot-imported separately.

This "table" package pulls in the Ginkgo's table-testing DSL
*/
package table

import (
	"github.com/onsi/ginkgo/v2"
)

type EntryDescription = ginkgo.EntryDescription

var DescribeTable = ginkgo.DescribeTable
var FDescribeTable = ginkgo.FDescribeTable
var PDescribeTable = ginkgo.PDescribeTable
var XDescribeTable = ginkgo.XDescribeTable

type TableEntry = ginkgo.TableEntry

var Entry = ginkgo.Entry
var FEntry = ginkgo.FEntry
var PEntry = ginkgo.PEntry
var XEntry = ginkgo.XEntry
