package ginkgo

import "github.com/onsi/ginkgo/internal"

//Offset(uint) is a decorator that allows you to change the stack-frame offset used when computing the line number of the node in question.
type Offset = internal.Offset

//FlakeAttempts(uint N) is a decorator that allows you to mark individual tests or test containers as flaky.  Ginkgo will run them up to `N` times until they pass.
type FlakeAttempts = internal.FlakeAttempts

//Focus is a decorator that allows you to mark a test or container as focused.  Identical to FIt and FDescribe.
const Focus = internal.Focus

//Pending is a decorator that allows you to mark a test or container as pending.  Identical to PIt and PDescribe.
const Pending = internal.Pending
