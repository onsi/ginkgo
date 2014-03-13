package integration_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"testing"
)

var tmpDir string

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)

	installGinkgoCommand := exec.Command("go", "install", "github.com/onsi/ginkgo/ginkgo")
	err := installGinkgoCommand.Run()
	if err != nil {
		fmt.Printf("Failed to compile Ginkgo\n\t%s", err.Error())
	}

	RunSpecs(t, "Integration Suite")
}

var _ = BeforeEach(func() {
	var err error
	tmpDir, err = ioutil.TempDir("", "ginkgo-run")
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())
})

func tmpPath(destination string) string {
	return filepath.Join(tmpDir, destination)
}

func copyIn(fixture string, destination string) {
	err := os.MkdirAll(destination, 0777)
	Ω(err).ShouldNot(HaveOccurred())

	output, err := exec.Command("cp", "-r", filepath.Join("_fixtures", fixture)+"/", destination).CombinedOutput()
	if !Ω(err).ShouldNot(HaveOccurred()) {
		fmt.Println(output)
	}
}

func runGinkgo(dir string, args ...string) (string, error) {
	cmd := exec.Command("ginkgo", args...)
	cmd.Dir = dir
	cmd.Env = []string{}
	for _, env := range os.Environ() {
		if !strings.Contains(env, "GINKGO_REMOTE_REPORTING_SERVER") {
			cmd.Env = append(cmd.Env, env)
		}
	}
	output, err := cmd.CombinedOutput()
	GinkgoWriter.Write(output)
	return string(output), err
}
