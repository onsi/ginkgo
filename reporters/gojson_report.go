package reporters

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/onsi/ginkgo/v2/internal/reporters"
	"github.com/onsi/ginkgo/v2/types"
	"golang.org/x/tools/go/packages"
)

func suitePathToPkg(dir string) (string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return "", err
	}
	if len(pkgs) != 1 {
		return "", errors.New("error")
	}
	return pkgs[0].ID, nil
}

// GenerateGoTestJSONReport produces a JSON-formatted in the test2json format used by `go test -json`
func GenerateGoTestJSONReport(report types.Report, destination string) error {
	// walk report and generate test2json-compatible objects
	// JSON-encode the objects into filename
	if err := os.MkdirAll(path.Dir(destination), 0770); err != nil {
		return err
	}
	f, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	r := reporters.NewGoJSONEventWriter(enc)

	// NOTE: could the Ginkgo report include the go package name?
	goPkg, err := suitePathToPkg(report.SuitePath)
	if err != nil {
		return err
	}
	// suite start events
	r.WriteSuiteStart(goPkg, report)
	for _, specReport := range report.SpecReports {
		if specReport.LeafNodeType == types.NodeTypeIt {
			r.WriteSpecStart(goPkg, specReport)
			r.WriteSpecOut(goPkg, specReport)
			r.WriteSpecResult(goPkg, specReport)
		} else {
			r.WriteSuiteLeafNodesOut(goPkg, specReport)
		}
	}
	r.WriteSuiteResult(goPkg, report)
	// suite end event
	return nil
}

// MergeJSONReports produces a single JSON-formatted report at the passed in destination by merging the JSON-formatted reports provided in sources
// It skips over reports that fail to decode but reports on them via the returned messages []string
func MergeAndCleanupGoTestJSONReports(sources []string, destination string) ([]string, error) {
	messages := []string{}
	if err := os.MkdirAll(path.Dir(destination), 0770); err != nil {
		return messages, err
	}
	f, err := os.Create(destination)
	if err != nil {
		return messages, err
	}
	defer f.Close()

	for _, source := range sources {
		data, err := os.ReadFile(source)
		if err != nil {
			messages = append(messages, fmt.Sprintf("Could not open %s:\n%s", source, err.Error()))
			continue
		}
		_, err = f.Write(data)
		if err != nil {
			messages = append(messages, fmt.Sprintf("Could not write to %s:\n%s", destination, err.Error()))
			continue
		}
	}
	return messages, nil
}
