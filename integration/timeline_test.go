package integration_test

import (
	"fmt"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Timeline output", func() {
	denoter := "•"
	retryDenoter := "↺"
	if runtime.GOOS == "windows" {
		denoter = "+"
		retryDenoter = "R"
	}
	BeforeEach(func() {
		fm.MountFixture("timeline")
	})

	Context("when running with succinct and normal verbosity", func() {
		argGroups := [][]string{
			{"--no-color", "--seed=17"},
			{"--no-color", "--seed=17", "--nodes=2"},
			{"--no-color", "--seed=17", "--succinct"},
			{"--no-color", "--seed=17", "--nodes=2", "--succinct"},
		}
		for _, args := range argGroups {
			args := args
			It(fmt.Sprintf("should emit a timeline (%s)", strings.Join(args, " ")), func() {
				session := startGinkgo(fm.PathTo("timeline"), args...)
				Eventually(session).Should(gexec.Exit(1))

				Ω(session).Should(gbytes.Say(`3 specs`))
				Ω(session).Should(gbytes.Say(`Automatically polling progress`))
				Ω(session).Should(gbytes.Say(`>\s*time\.Sleep\(time\.Millisecond \* 200\)`))
				Ω(session).Should(gbytes.Say(retryDenoter + ` \[FLAKEY TEST - TOOK 3 ATTEMPTS TO PASS\]`))
				Ω(session).Should(gbytes.Say(`a full timeline a flaky test retries a few times`))
				Ω(session).Should(gbytes.Say(`Report Entries >>`))
				Ω(session).Should(gbytes.Say(`a report! - `))
				Ω(session).Should(gbytes.Say(`  Of great value`))
				Ω(session).Should(gbytes.Say(`a report! - `))
				Ω(session).Should(gbytes.Say(`  Of great value`))
				Ω(session).Should(gbytes.Say(`a report! - `))
				Ω(session).Should(gbytes.Say(`  Of great value`))
				Ω(session).Should(gbytes.Say(`<< Report Entries`))
				Ω(session).Should(gbytes.Say(denoter + ` \[TIMEDOUT\]`))
				Ω(session).Should(gbytes.Say(`a full timeline a test with multiple failures \[It\] times out`))
				Ω(session).Should(gbytes.Say(`Timeline >>`))
				Ω(session).Should(gbytes.Say(`STEP: waiting...`))
				Ω(session).Should(gbytes.Say(`\[TIMEDOUT] in \[It\]`))
				Ω(session).Should(gbytes.Say(`then failing!`))
				Ω(session).Should(gbytes.Say(`\[FAILED\] in \[It\]`))
				Ω(session).Should(gbytes.Say(`\[PANICKED\] in \[AfterEach\]`))
				Ω(session).Should(gbytes.Say(`<< Timeline`))
				Ω(session).Should(gbytes.Say(`\[TIMEDOUT\] A node timeout occurred`))
				Ω(session).Should(gbytes.Say(`This is the Progress Report generated when the node timeout occurred:`))
				Ω(session).Should(gbytes.Say(`>.*<-ctx.Done\(\)`))
				Ω(session).Should(gbytes.Say(`\[FAILED\] A node timeout occurred and then the following failure was recorded in the timedout node before it exited:`))
				Ω(session).Should(gbytes.Say(`welp`))
				Ω(session).Should(gbytes.Say(`In \[It\] at:`))
				Ω(session).Should(gbytes.Say(`There were additional failures detected.  To view them in detail run ginkgo -vv`))
				Ω(session).Should(gbytes.Say(denoter))
				Ω(session).Should(gbytes.Say(`Summarizing 1 Failure:`))
				Ω(session).Should(gbytes.Say(`\[TIMEDOUT\]`))
			})
		}
	})

	Context("when running with -v", func() {
		It("should emit a timeline", func() {
			session := startGinkgo(fm.PathTo("timeline"), "--no-color", "--seed=17", "-v")
			Eventually(session).Should(gexec.Exit(1))

			Ω(session).Should(gbytes.Say(`3 specs`))

			Ω(session).Should(gbytes.Say(`a full timeline a flaky test retries a few times`))
			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] in \[It\]`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(`Attempt #1 Failed.  Retrying ` + retryDenoter))

			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] in \[It\]`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(`Attempt #2 Failed.  Retrying ` + retryDenoter))

			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`hooray!`))
			Ω(session).Should(gbytes.Say(`feeling sleepy...`))
			Ω(session).Should(gbytes.Say(`Automatically polling progress`))
			Ω(session).Should(gbytes.Say(`>\s*time\.Sleep\(time.Millisecond \* 200\)`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(retryDenoter + ` \[FLAKEY TEST - TOOK 3 ATTEMPTS TO PASS\]`))

			Ω(session).Should(gbytes.Say(`a full timeline a test with multiple failures \[It\] times out`))
			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\] A node timeout occurred`))
			Ω(session).Should(gbytes.Say(`This is the Progress Report generated when the node timeout occurred:`))
			Ω(session).Should(gbytes.Say(`>\s*<-ctx\.Done\(\)`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] A node timeout occurred and then the following failure was recorded in the timedout node before it exited:`))
			Ω(session).Should(gbytes.Say(`welp`))
			Ω(session).Should(gbytes.Say(`In \[It\]`))
			Ω(session).Should(gbytes.Say(`There were additional failures detected.  To view them in detail run ginkgo -vv`))

			Ω(session).Should(gbytes.Say(`a full timeline passes happily`))
			Ω(session).Should(gbytes.Say(`a verbose-only report`))
			Ω(session).ShouldNot(gbytes.Say(`a hidden report`))
			Ω(session).Should(gbytes.Say(denoter))

			Ω(session).Should(gbytes.Say(`Summarizing 1 Failure:`))
			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\]`))
		})
	})

	Context("when running with -vv", func() {
		It("should emit a timeline", func() {
			session := startGinkgo(fm.PathTo("timeline"), "--no-color", "--seed=17", "-vv")
			Eventually(session).Should(gexec.Exit(1))

			Ω(session).Should(gbytes.Say(`3 specs`))

			Ω(session).Should(gbytes.Say(`a full timeline\n`))
			Ω(session).Should(gbytes.Say(`a flaky test\n`))
			Ω(session).Should(gbytes.Say(`retries a few times\n`))
			Ω(session).Should(gbytes.Say(`> Enter \[BeforeEach\] a flaky test`))
			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`< Exit \[BeforeEach\] a flaky test`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] bam!`))
			Ω(session).Should(gbytes.Say(`In \[It\]`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(`END STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`Attempt #1 Failed.  Retrying ` + retryDenoter))

			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] bam!`))
			Ω(session).Should(gbytes.Say(`In \[It\]`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(`Attempt #2 Failed.  Retrying ` + retryDenoter))

			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`hooray!`))
			Ω(session).Should(gbytes.Say(`feeling sleepy...`))
			Ω(session).Should(gbytes.Say(`Automatically polling progress`))
			Ω(session).Should(gbytes.Say(`>\s*time\.Sleep\(time.Millisecond \* 200\)`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(retryDenoter + ` \[FLAKEY TEST - TOOK 3 ATTEMPTS TO PASS\]`))

			Ω(session).Should(gbytes.Say(`times out`))
			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\] A node timeout occurred`))
			Ω(session).Should(gbytes.Say(`This is the Progress Report generated when the node timeout occurred:`))
			Ω(session).Should(gbytes.Say(`>\s*<-ctx\.Done\(\)`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] A node timeout occurred and then the following failure was recorded in the timedout node before it exited:`))
			Ω(session).Should(gbytes.Say(`welp`))
			Ω(session).Should(gbytes.Say(`In \[It\]`))
			Ω(session).Should(gbytes.Say(`\[PANICKED\] Test Panicked`))
			Ω(session).Should(gbytes.Say(`aaah!`))
			Ω(session).Should(gbytes.Say(`Full Stack Trace`))
			Ω(session).Should(gbytes.Say(denoter + ` \[TIMEDOUT\]`))

			Ω(session).Should(gbytes.Say(`passes happily`))
			Ω(session).Should(gbytes.Say(`> Enter \[It\] passes happily`))
			Ω(session).Should(gbytes.Say(`a verbose-only report`))
			Ω(session).ShouldNot(gbytes.Say(`a hidden report`))
			Ω(session).Should(gbytes.Say(`< Exit \[It\] passes happily`))
			Ω(session).Should(gbytes.Say(denoter))

			Ω(session).Should(gbytes.Say(`Summarizing 1 Failure:`))
			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\]`))
		})
	})

	Context("when running with -v in parallel", func() {
		It("should emit a timeline", func() {
			session := startGinkgo(fm.PathTo("timeline"), "--no-color", "--seed=17", "-v", "-nodes=2")
			Eventually(session).Should(gexec.Exit(1))

			Ω(session).Should(gbytes.Say(`3 specs`))

			Ω(session).Should(gbytes.Say(retryDenoter + ` \[FLAKEY TEST - TOOK 3 ATTEMPTS TO PASS\]`))
			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] in \[It\]`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(`Attempt #1 Failed.  Retrying ` + retryDenoter))

			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] in \[It\]`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(`Attempt #2 Failed.  Retrying ` + retryDenoter))

			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`hooray!`))
			Ω(session).Should(gbytes.Say(`feeling sleepy...`))
			Ω(session).Should(gbytes.Say(`Automatically polling progress`))
			Ω(session).Should(gbytes.Say(`>\s*time\.Sleep\(time.Millisecond \* 200\)`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))

			Ω(session).Should(gbytes.Say(`a full timeline a test with multiple failures \[It\] times out`))
			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\] A node timeout occurred`))
			Ω(session).Should(gbytes.Say(`This is the Progress Report generated when the node timeout occurred:`))
			Ω(session).Should(gbytes.Say(`>\s*<-ctx\.Done\(\)`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] A node timeout occurred and then the following failure was recorded in the timedout node before it exited:`))
			Ω(session).Should(gbytes.Say(`welp`))
			Ω(session).Should(gbytes.Say(`In \[It\]`))
			Ω(session).Should(gbytes.Say(`There were additional failures detected.  To view them in detail run ginkgo -vv`))

			Ω(session).Should(gbytes.Say(denoter))
			Ω(session).Should(gbytes.Say(`a full timeline passes happily`))
			Ω(session).Should(gbytes.Say(`a verbose-only report`))
			Ω(session).ShouldNot(gbytes.Say(`a hidden report`))

			Ω(session).Should(gbytes.Say(`Summarizing 1 Failure:`))
			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\]`))
		})
	})

	Context("when running with -vv in parallel", func() {
		It("should emit a timeline", func() {
			session := startGinkgo(fm.PathTo("timeline"), "--no-color", "--seed=17", "-vv", "-nodes=2")
			Eventually(session).Should(gexec.Exit(1))

			Ω(session).Should(gbytes.Say(`3 specs`))

			Ω(session).Should(gbytes.Say(retryDenoter + ` \[FLAKEY TEST - TOOK 3 ATTEMPTS TO PASS\]`))
			Ω(session).Should(gbytes.Say(`a full timeline\n`))
			Ω(session).Should(gbytes.Say(`a flaky test\n`))
			Ω(session).Should(gbytes.Say(`retries a few times\n`))
			Ω(session).Should(gbytes.Say(`> Enter \[BeforeEach\] a flaky test`))
			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`< Exit \[BeforeEach\] a flaky test`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] Failure recorded during attempt 1:`))
			Ω(session).Should(gbytes.Say(`bam!`))
			Ω(session).Should(gbytes.Say(`In \[It\]`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(`END STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`Attempt #1 Failed.  Retrying ` + retryDenoter))

			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] Failure recorded during attempt 2:`))
			Ω(session).Should(gbytes.Say(`bam!`))
			Ω(session).Should(gbytes.Say(`In \[It\]`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))
			Ω(session).Should(gbytes.Say(`Attempt #2 Failed.  Retrying ` + retryDenoter))

			Ω(session).Should(gbytes.Say(`STEP: logging some events`))
			Ω(session).Should(gbytes.Say(`hello!`))
			Ω(session).Should(gbytes.Say(`a report!`))
			Ω(session).Should(gbytes.Say(`  Of great value`))
			Ω(session).Should(gbytes.Say(`let's try...`))
			Ω(session).Should(gbytes.Say(`hooray!`))
			Ω(session).Should(gbytes.Say(`feeling sleepy...`))
			Ω(session).Should(gbytes.Say(`Automatically polling progress`))
			Ω(session).Should(gbytes.Say(`>\s*time\.Sleep\(time.Millisecond \* 200\)`))
			Ω(session).Should(gbytes.Say(`STEP: cleaning up a bit`))
			Ω(session).Should(gbytes.Say(`all done!`))

			Ω(session).Should(gbytes.Say(denoter + ` \[TIMEDOUT\]`))
			Ω(session).Should(gbytes.Say(`times out`))
			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\] A node timeout occurred`))
			Ω(session).Should(gbytes.Say(`This is the Progress Report generated when the node timeout occurred:`))
			Ω(session).Should(gbytes.Say(`>\s*<-ctx\.Done\(\)`))
			Ω(session).Should(gbytes.Say(`\[FAILED\] A node timeout occurred and then the following failure was recorded in the timedout node before it exited:`))
			Ω(session).Should(gbytes.Say(`welp`))
			Ω(session).Should(gbytes.Say(`In \[It\]`))
			Ω(session).Should(gbytes.Say(`\[PANICKED\] Test Panicked`))
			Ω(session).Should(gbytes.Say(`aaah!`))
			Ω(session).Should(gbytes.Say(`Full Stack Trace`))

			Ω(session).Should(gbytes.Say(denoter))
			Ω(session).Should(gbytes.Say(`passes happily`))
			Ω(session).Should(gbytes.Say(`> Enter \[It\] passes happily`))
			Ω(session).Should(gbytes.Say(`a verbose-only report`))
			Ω(session).ShouldNot(gbytes.Say(`a hidden report`))
			Ω(session).Should(gbytes.Say(`< Exit \[It\] passes happily`))

			Ω(session).Should(gbytes.Say(`Summarizing 1 Failure:`))
			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\]`))
		})
	})
})
