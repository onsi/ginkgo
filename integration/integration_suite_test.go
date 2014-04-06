package integration_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
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

	filepath.Walk(filepath.Join("_fixtures", fixture), func(path string, info os.FileInfo, err error) error {
		base := filepath.Base(path)
		if base == fixture {
			return nil
		}

		src, err := os.Open(path)
		Ω(err).ShouldNot(HaveOccurred())

		dst, err := os.Create(filepath.Join(destination, base))
		Ω(err).ShouldNot(HaveOccurred())

		_, err = io.Copy(dst, src)
		Ω(err).ShouldNot(HaveOccurred())
		return nil
	})
}

func ginkgoCommand(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("ginkgo", args...)
	cmd.Dir = dir
	cmd.Env = []string{}
	for _, env := range os.Environ() {
		if !strings.Contains(env, "GINKGO_REMOTE_REPORTING_SERVER") {
			cmd.Env = append(cmd.Env, env)
		}
	}

	return cmd
}

func runGinkgo(dir string, args ...string) (string, error) {
	cmd := ginkgoCommand(dir, args...)
	output, err := cmd.CombinedOutput()
	GinkgoWriter.Write(output)
	return string(output), err
}
