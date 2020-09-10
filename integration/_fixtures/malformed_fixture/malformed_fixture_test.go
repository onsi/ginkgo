package malformed_fixture_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("MalformedFixture", func() {
	It("tries to install a container within an It...", func() {
		Context("...which is not allowed!", func() {

		})
	})
})
