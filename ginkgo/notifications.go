package main

import (
	"fmt"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"os"
	"os/exec"
)

func verifyNotificationsAreAvailable() {
	_, err := exec.LookPath("terminal-notifier")
	if err != nil {
		fmt.Printf(`--notify requires terminal-notifier, which you don't seem to have installed.

To remedy this:

    brew install terminal-notifier

To learn more about terminal-notifier:

    https://github.com/alloy/terminal-notifier
`)
		os.Exit(1)
	}
}

func sendSuiteCompletionNotification(suite *testsuite.TestSuite, suitePassed bool) {
	if suitePassed {
		sendNotification("Ginkgo [PASS]", fmt.Sprintf(`Test suite for "%s" passed.`, suite.PackageName))
	} else {
		sendNotification("Ginkgo [FAIL]", fmt.Sprintf(`Test suite for "%s" failed.`, suite.PackageName))
	}
}

func sendNotification(title string, subtitle string) {
	args := []string{"-title", title, "-subtitle", subtitle, "-group", "com.onsi.ginkgo"}

	terminal := os.Getenv("TERM_PROGRAM")
	if terminal == "iTerm.app" {
		args = append(args, "-activate", "com.googlecode.iterm2")
	} else if terminal == "Apple_Terminal" {
		args = append(args, "-activate", "com.apple.Terminal")
	}

	if notify {
		exec.Command("terminal-notifier", args...).Run()
	}
}
