package sleep_on_failure_fixture_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("sleep on failure", func() {
	It("passes and is never paused", func() {
		fmt.Fprintln(os.Stdout, "PASSING-SPEC-RAN")
	})

	It("fails and should be paused before teardown", func() {
		fmt.Fprintln(os.Stdout, "FAILING-SPEC-BODY-RAN")
		Fail("intentional failure")
	})

	AfterEach(func() {
		fmt.Fprintln(os.Stdout, "TEARDOWN-RAN")
	})
})
