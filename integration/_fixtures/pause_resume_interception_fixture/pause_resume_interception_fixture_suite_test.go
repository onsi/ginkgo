package pause_resume_interception_fixture_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPauseResumeInterceptionFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PauseResumeInterceptionFixture Suite")
}

var _ = It("can pause and resume interception", func() {
	fmt.Println("CAPTURED OUTPUT A")
	PauseOutputInterception()
	fmt.Println("OUTPUT TO CONSOLE")
	ResumeOutputInterception()
	fmt.Println("CAPTURED OUTPUT B")
})
