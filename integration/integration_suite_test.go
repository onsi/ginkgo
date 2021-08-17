package integration_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gexec"

	"testing"
	"time"
)

var pathToGinkgo string
var DEBUG bool
var fm FixtureManager

func init() {
	flag.BoolVar(&DEBUG, "debug", false, "keep assets around after test run")
}

func TestIntegration(t *testing.T) {
	SetDefaultEventuallyTimeout(30 * time.Second)
	format.TruncatedDiff = false
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
	fm = NewFixtureManager(fmt.Sprintf("tmp_%d", GinkgoParallelNode()))
})

var _ = AfterEach(func() {
	if !DEBUG {
		fm.Cleanup()
	}
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

type FixtureManager struct {
	TmpDir string
}

func NewFixtureManager(tmpDir string) FixtureManager {
	err := os.MkdirAll(tmpDir, 0700)
	Ω(err).ShouldNot(HaveOccurred())
	return FixtureManager{
		TmpDir: tmpDir,
	}
}

func (f FixtureManager) Cleanup() {
	Ω(os.RemoveAll(f.TmpDir)).Should(Succeed())
}

func (f FixtureManager) MountFixture(fixture string, subPackage ...string) {
	src := filepath.Join("_fixtures", fixture+"_fixture")
	dst := filepath.Join(f.TmpDir, fixture)

	if len(subPackage) > 0 {
		src = filepath.Join(src, subPackage[0])
		dst = filepath.Join(dst, subPackage[0])
	}

	f.copyAndRewrite(src, dst)
}

func (f FixtureManager) copyAndRewrite(src string, dst string) {
	Expect(os.MkdirAll(dst, 0777)).To(Succeed())

	files, err := os.ReadDir(src)
	Expect(err).NotTo(HaveOccurred())

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		dstPath := filepath.Join(dst, file.Name())
		if file.IsDir() {
			f.copyAndRewrite(srcPath, dstPath)
			continue
		}

		srcContent, err := os.ReadFile(srcPath)
		Ω(err).ShouldNot(HaveOccurred())
		//rewrite import statements so that fixtures can work in the fixture folder when developing them, and in the tmp folder when under test
		srcContent = bytes.ReplaceAll(srcContent, []byte("github.com/onsi/ginkgo/integration/_fixtures"), []byte(f.PackageRoot()))
		srcContent = bytes.ReplaceAll(srcContent, []byte("_fixture"), []byte(""))
		Ω(os.WriteFile(dstPath, srcContent, 0666)).Should(Succeed())
	}
}

func (f FixtureManager) AbsPathTo(pkg string, target ...string) string {
	path, err := filepath.Abs(f.PathTo(pkg, target...))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return path
}

func (f FixtureManager) PathTo(pkg string, target ...string) string {
	if len(target) == 0 {
		return filepath.Join(f.TmpDir, pkg)
	}
	return filepath.Join(f.TmpDir, pkg, target[0])
}

func (f FixtureManager) PathToFixtureFile(pkg string, target string) string {
	return filepath.Join("_fixtures", pkg+"_fixture", target)
}

func (f FixtureManager) WriteFile(pkg string, target string, content string) {
	dst := f.PathTo(pkg, target)
	err := os.WriteFile(dst, []byte(content), 0666)
	Ω(err).ShouldNot(HaveOccurred())
}

func (f FixtureManager) ContentOf(pkg string, target string) string {
	content, err := os.ReadFile(f.PathTo(pkg, target))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return string(content)
}

func (f FixtureManager) LoadJSONReports(pkg string, target string) []types.Report {
	data := []byte(f.ContentOf(pkg, target))
	reports := []types.Report{}
	ExpectWithOffset(1, json.Unmarshal(data, &reports)).Should(Succeed())
	return reports
}

func (f FixtureManager) LoadJUnitReport(pkg string, target string) reporters.JUnitTestSuites {
	data := []byte(f.ContentOf(pkg, target))
	reports := reporters.JUnitTestSuites{}
	ExpectWithOffset(1, xml.Unmarshal(data, &reports)).Should(Succeed())
	return reports
}

func (f FixtureManager) ContentOfFixture(pkg string, target string) string {
	content, err := os.ReadFile(f.PathToFixtureFile(pkg, target))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return string(content)
}

func (f FixtureManager) RemoveFile(pkg string, target string) {
	Expect(os.RemoveAll(f.PathTo(pkg, target))).To(Succeed())
}

func (f FixtureManager) PackageRoot() string {
	return "github.com/onsi/ginkgo/integration/" + f.TmpDir
}

func (f FixtureManager) PackageNameFor(target string) string {
	return f.PackageRoot() + "/" + target
}

func sameFile(filePath, otherFilePath string) bool {
	content, readErr := os.ReadFile(filePath)
	Expect(readErr).NotTo(HaveOccurred())
	otherContent, readErr := os.ReadFile(otherFilePath)
	Expect(readErr).NotTo(HaveOccurred())
	Expect(string(content)).To(Equal(string(otherContent)))
	return true
}

func sameFolder(sourcePath, destinationPath string) bool {
	files, err := os.ReadDir(sourcePath)
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
