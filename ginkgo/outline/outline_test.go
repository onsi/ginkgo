package outline

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("Validate outline from file with",
	func(srcFilename, jsonOutlineFilename, csvOutlineFilename string) {
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, filepath.Join("_testdata", srcFilename), nil, 0)
		Expect(err).To(BeNil(), "error parsing source: %s", err)

		if err != nil {
			log.Fatalf("error parsing source: %s", err)
		}

		o, err := FromASTFile(astFile)
		Expect(err).To(BeNil(), "error creating outline: %s", err)

		gotJSON, err := json.MarshalIndent(o, "", "  ")
		Expect(err).To(BeNil(), "error marshalling outline to json: %s", err)

		wantJSON, err := ioutil.ReadFile(filepath.Join("_testdata", jsonOutlineFilename))
		Expect(err).To(BeNil(), "error reading JSON outline fixture: %s", err)

		Expect(gotJSON).To(MatchJSON(wantJSON))

		gotCSV := o.String()

		wantCSV, err := ioutil.ReadFile(filepath.Join("_testdata", csvOutlineFilename))
		Expect(err).To(BeNil(), "error reading CSV outline fixture: %s", err)

		Expect(gotCSV).To(Equal(string(wantCSV)))
	},
	Entry("normal import of ginkgo package (no dot, no alias), normal container and specs", "nodot_test.go", "nodot_test_outline.json", "nodot_test_outline.csv"),
	Entry("aliased import of ginkgo package, normal container and specs", "alias_test.go", "alias_test_outline.json", "alias_test_outline.csv"),
	Entry("normal containers and specs", "normal_test.go", "normal_test_outline.json", "normal_test_outline.csv"),
	Entry("focused containers and specs", "focused_test.go", "focused_test_outline.json", "focused_test_outline.csv"),
	Entry("pending containers and specs", "pending_test.go", "pending_test_outline.json", "pending_test_outline.csv"),
)
