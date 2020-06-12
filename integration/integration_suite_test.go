package integration_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
	"time"
)

var tmpDir string
var pathToGinkgo string

func TestIntegration(t *testing.T) {
	SetDefaultEventuallyTimeout(30 * time.Second)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	pathToGinkgo, err := gexec.Build("../ginkgo")
	Ω(err).ShouldNot(HaveOccurred())
	return []byte(pathToGinkgo)
}, func(computedPathToGinkgo []byte) {
	pathToGinkgo = string(computedPathToGinkgo)
})

var _ = BeforeEach(func() {
	tmpDir = fmt.Sprintf("./ginko-run-%d", GinkgoParallelNode())
	err := os.Mkdir(tmpDir, 0700)
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	os.RemoveAll(tmpDir)
	gexec.CleanupBuildArtifacts()
})

func tmpPath(destination string) string {
	return filepath.Join(tmpDir, destination)
}

func fixturePath(name string) string {
	return filepath.Join("_fixtures", name)
}

func copyIn(sourcePath, destinationPath string, recursive bool) {
	err := os.MkdirAll(destinationPath, 0777)
	Expect(err).NotTo(HaveOccurred())

	files, err := ioutil.ReadDir(sourcePath)
	Expect(err).NotTo(HaveOccurred())
	for _, f := range files {
		srcPath := filepath.Join(sourcePath, f.Name())
		dstPath := filepath.Join(destinationPath, f.Name())
		if f.IsDir() {
			if recursive {
				copyIn(srcPath, dstPath, recursive)
			}
			continue
		}

		src, err := os.Open(srcPath)

		Expect(err).NotTo(HaveOccurred())
		defer src.Close()

		dst, err := os.Create(dstPath)
		Expect(err).NotTo(HaveOccurred())
		defer dst.Close()

		_, err = io.Copy(dst, src)
		Expect(err).NotTo(HaveOccurred())
	}
}

func sameFile(filePath, otherFilePath string) bool {
	content, readErr := ioutil.ReadFile(filePath)
	Expect(readErr).NotTo(HaveOccurred())
	otherContent, readErr := ioutil.ReadFile(otherFilePath)
	Expect(readErr).NotTo(HaveOccurred())
	Expect(string(content)).To(Equal(string(otherContent)))
	return true
}

func sameFolder(sourcePath, destinationPath string) bool {
	files, err := ioutil.ReadDir(sourcePath)
	Expect(err).NotTo(HaveOccurred())
	for _, f := range files {
		srcPath := filepath.Join(sourcePath, f.Name())
		dstPath := filepath.Join(destinationPath, f.Name())
		if f.IsDir() {
			sameFolder(srcPath, dstPath)
			continue
		}
		Expect(sameFile(srcPath, dstPath)).To(BeTrue())
	}
	return true
}

func ginkgoCommand(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command(pathToGinkgo, args...)
	cmd.Dir = dir

	return cmd
}

func startGinkgo(dir string, args ...string) *gexec.Session {
	cmd := ginkgoCommand(dir, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	return session
}

func removeSuccessfully(path string) {
	err := os.RemoveAll(path)
	Expect(err).NotTo(HaveOccurred())
}

func raceDetectorSupported() bool {
	// https://github.com/golang/go/blob/1a370950/src/cmd/internal/sys/supported.go#L12
	switch runtime.GOOS {
	case "linux":
		return runtime.GOARCH == "amd64" || runtime.GOARCH == "ppc64le" || runtime.GOARCH == "arm64"
	case "darwin", "freebsd", "netbsd", "windows":
		return runtime.GOARCH == "amd64"
	default:
		return false
	}
}
