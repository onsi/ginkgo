/*
Ginkgo isusually dot-imported via:

    import . "github.com/onsi/ginkgo/v2"

however some parts of the DSL may conflict with existing symbols in the user's code.

To mitigate this without losing the brevity of dot-importing Ginkgo the various packages in the
dsl directory provide pieces of the Ginkgo DSL that can be dot-imported separately.

This "decorators" package pulls in the various decorators defined in the Ginkgo DSL.
*/
package decorators

import (
	"github.com/onsi/ginkgo/v2"
)

type Offset = ginkgo.Offset
type FlakeAttempts = ginkgo.FlakeAttempts
type Labels = ginkgo.Labels

const Focus = ginkgo.Focus
const Pending = ginkgo.Pending
const Serial = ginkgo.Serial
const Ordered = ginkgo.Ordered
const OncePerOrdered = ginkgo.OncePerOrdered

var Label = ginkgo.Label
