package integration_test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Failures File", func() {
	var pathToTest string

	BeforeEach(func() {
		pathToTest = tmpPath("failures-file")
		copyIn("failures_file_fixture", pathToTest)
	})

	It("writes a file including test failures and reruns only the test in the file when told to", func() {
		session := startGinkgo(pathToTest, "--noColor", "--writeFailuresFile=failures-file.json", "--nodes=2", "--randomizeAllSpecs")
		Eventually(session).Should(gexec.Exit(1))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("3 Passed"))
		Ω(output).Should(ContainSubstring("4 Failed"))
		Ω(output).Should(ContainSubstring("0 Pending"))
		Ω(output).Should(ContainSubstring("0 Skipped"))

		failuresFile := filepath.Join(pathToTest, "failures-file.json")
		Ω(failuresFile).Should(BeAnExistingFile())
		encoded, err := ioutil.ReadFile(failuresFile)
		Ω(err).ShouldNot(HaveOccurred())

		failures := []types.FailuresFileEntry{}
		Ω(json.Unmarshal(encoded, &failures)).Should(Succeed())
		Ω(failures).Should(HaveLen(4))

		session = startGinkgo(pathToTest, "--noColor", "--runFailuresFile=failures-file.json", "--nodes=2", "--randomizeAllSpecs")
		Eventually(session).Should(gexec.Exit(1))
		output = string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("0 Passed"))
		Ω(output).Should(ContainSubstring("4 Failed"))
		Ω(output).Should(ContainSubstring("0 Pending"))
		Ω(output).Should(ContainSubstring("3 Skipped"))
	})
})
