package outline

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
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

		o, err := FromASTFile(fset, astFile)
		Expect(err).To(BeNil(), "error creating outline: %s", err)

		gotJSON, err := json.MarshalIndent(o, "", "  ")
		Expect(err).To(BeNil(), "error marshalling outline to json: %s", err)

		wantJSON, err := os.ReadFile(filepath.Join("_testdata", jsonOutlineFilename))
		Expect(err).To(BeNil(), "error reading JSON outline fixture: %s", err)
		Expect(gotJSON).To(MatchJSON(wantJSON))

		gotCSV := o.String()

		wantCSV, err := os.ReadFile(filepath.Join("_testdata", csvOutlineFilename))
		Expect(err).To(BeNil(), "error reading CSV outline fixture: %s", err)

		Expect(gotCSV).To(Equal(string(wantCSV)))

		ensureRecordsAreIdentical(csvOutlineFilename, jsonOutlineFilename)
	},
	// To add a test:
	// 1. Create the input, e.g., `myspecialcase_test.go`
	// 2. Create the sample CSV and JSON results: Run `bash ./_testdata/create_result.sh ./_testdata/myspecialcase_test.go`
	// 3. Add an Entry below, by copying an existing one, and substituting `myspecialcase` where needed.
	// To re-create the sample results for a test:
	// 1. Run `bash ./_testdata/create_result.sh ./testdata/myspecialcase_test.go`
	// To re-create the sample results for all tests:
	// 1. Run `for name in ./_testdata/*_test.go; do bash ./_testdata/create_result.sh $name; done`
	Entry("normal import of ginkgo package (no dot, no alias), normal container and specs", "nodot_test.go", "nodot_test.go.json", "nodot_test.go.csv"),
	Entry("aliased import of ginkgo package, normal container and specs", "alias_test.go", "alias_test.go.json", "alias_test.go.csv"),
	Entry("normal containers and specs", "normal_test.go", "normal_test.go.json", "normal_test.go.csv"),
	Entry("focused containers and specs", "focused_test.go", "focused_test.go.json", "focused_test.go.csv"),
	Entry("pending containers and specs", "pending_test.go", "pending_test.go.json", "pending_test.go.csv"),
	Entry("nested focused containers and specs", "nestedfocused_test.go", "nestedfocused_test.go.json", "nestedfocused_test.go.csv"),
	Entry("mixed focused containers and specs", "mixed_test.go", "mixed_test.go.json", "mixed_test.go.csv"),
	Entry("specs used to verify position", "position_test.go", "position_test.go.json", "position_test.go.csv"),
	Entry("suite setup", "suite_test.go", "suite_test.go.json", "suite_test.go.csv"),
	Entry("core dsl import", "dsl_core_test.go", "dsl_core_test.go.json", "dsl_core_test.go.csv"),
	Entry("labels decorator on containers and specs", "labels_test.go", "labels_test.go.json", "labels_test.go.csv"),
	Entry("pending decorator on containers and specs", "pending_decorator_test.go", "pending_decorator_test.go.json", "pending_decorator_test.go.csv"),
	Entry("proper csv escaping of all fields", "csv_proper_escaping_test.go", "csv_proper_escaping_test.go.json", "csv_proper_escaping_test.go.csv"),
)

var _ = Describe("Validate position", func() {
	It("should report the correct start and end byte offsets of the ginkgo container or spec", func() {
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, filepath.Join("_testdata", "position_test.go"), nil, 0)
		Expect(err).To(BeNil(), "error parsing source: %s", err)

		if err != nil {
			log.Fatalf("error parsing source: %s", err)
		}

		o, err := FromASTFile(fset, astFile)
		Expect(err).To(BeNil(), "error creating outline: %s", err)

		for _, n := range o.Nodes {
			n.PreOrder(func(n *ginkgoNode) {
				wantPositions := strings.Split(n.Text, ",")
				Expect(len(wantPositions)).To(Equal(2), "test fixture node text should be \"start position,end position")
				wantStart, err := strconv.Atoi(wantPositions[0])
				Expect(err).To(BeNil(), "could not parse start offset")
				wantEnd, err := strconv.Atoi(wantPositions[1])
				Expect(err).To(BeNil(), "could not parse end offset")

				Expect(int(n.Start)).To(Equal(wantStart), fmt.Sprintf("test fixture node text says the node should start at %d, but it starts at %d", wantStart, n.Start))
				Expect(int(n.End)).To(Equal(wantEnd), fmt.Sprintf("test fixture node text says the node should end at %d, but it ends at %d", wantEnd, n.End))
			})
		}

	})
})

func ensureRecordsAreIdentical(csvOutlineFilename, jsonOutlineFilename string) {
	csvFile, err := os.Open(filepath.Join("_testdata", csvOutlineFilename))
	Expect(err).To(BeNil(), "error opening CSV outline fixture: %s", err)
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	csvRows, err := csvReader.ReadAll()
	Expect(err).To(BeNil(), "error reading CSV outline fixture: %s", err)

	// marshal csvRows into some comparable shape
	csvFields := csvRows[0]
	var csvRecords []ginkgoMetadata
	for i := 1; i < len(csvRows); i++ {
		var record ginkgoMetadata

		for j, field := range csvFields {

			field = strings.ToLower(field)
			value := csvRows[i][j]

			switch field {
			case "name":
				record.Name = value
			case "text":
				record.Text = value
			case "start":
				start, err := strconv.Atoi(value)
				Expect(err).To(BeNil(), "error converting start to int: %s", err)
				record.Start = start
			case "end":
				end, err := strconv.Atoi(value)
				Expect(err).To(BeNil(), "error converting end to int: %s", err)
				record.End = end
			case "spec":
				spec, err := strconv.ParseBool(value)
				Expect(err).To(BeNil(), "error converting spec to bool: %s", err)
				record.Spec = spec
			case "focused":
				focused, err := strconv.ParseBool(value)
				Expect(err).To(BeNil(), "error converting focused to bool: %s", err)
				record.Focused = focused
			case "pending":
				pending, err := strconv.ParseBool(value)
				Expect(err).To(BeNil(), "error converting pending to bool: %s", err)
				record.Pending = pending
			case "labels":
				// strings.Split will return [""] for an empty string, we want []
				if value == "" {
					record.Labels = []string{}
				} else {
					record.Labels = strings.Split(value, ", ")
				}
			default:
				Fail(fmt.Sprintf("unexpected field: %s", field))
			}
		}

		// "By" is a special case; nil out its labels so we can compare to the parsed JSON
		if record.Name == "By" {
			record.Labels = nil
		}

		csvRecords = append(csvRecords, record)
	}

	jsonFile, err := os.Open(filepath.Join("_testdata", jsonOutlineFilename))
	Expect(err).To(BeNil(), "error opening JSON outline fixture: %s", err)
	defer jsonFile.Close()

	jsonDecoder := json.NewDecoder(jsonFile)
	var jsonRows []ginkgoNode
	err = jsonDecoder.Decode(&jsonRows)
	Expect(err).To(BeNil(), "error reading JSON outline fixture: %s", err)

	// marshal jsonRows into some comparable shape - the hierarchical structure needs to be flattened
	var jsonRecords []ginkgoMetadata
	flattenNodes(jsonRows, &jsonRecords)

	Expect(csvRecords).To(Equal(jsonRecords))
}

// flattenNodes converts the hierarchical json output into a list of records
func flattenNodes(nodes []ginkgoNode, flatNodes *[]ginkgoMetadata) {
	for _, node := range nodes {
		record := ginkgoMetadata{
			Name:    node.Name,
			Text:    node.Text,
			Start:   node.Start,
			End:     node.End,
			Spec:    node.Spec,
			Focused: node.Focused,
			Pending: node.Pending,
			Labels:  node.Labels,
		}

		*flatNodes = append(*flatNodes, record)

		// handle nested nodes
		if len(node.Nodes) > 0 {
			var nestedNodes []ginkgoNode
			for _, nestedNode := range node.Nodes {
				nestedNodes = append(nestedNodes, *nestedNode)
			}
			flattenNodes(nestedNodes, flatNodes)
		}
	}
}
