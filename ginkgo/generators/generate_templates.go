package generators

var specText = `package {{.Package}}

import (
	{{if .IncludeImports}}. "github.com/onsi/ginkgo"{{end}}
	{{if .IncludeImports}}. "github.com/onsi/gomega"{{end}}

	{{if .ImportPackage}}"{{.PackageImportPath}}"{{end}}
)

var _ = Describe("{{.Subject}}", func() {

})
`

var agoutiSpecText = `package {{.Package}}

import (
	{{if .IncludeImports}}. "github.com/onsi/ginkgo"{{end}}
	{{if .IncludeImports}}. "github.com/onsi/gomega"{{end}}
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	{{if .ImportPackage}}"{{.PackageImportPath}}"{{end}}
)

var _ = Describe("{{.Subject}}", func() {
	var page *agouti.Page

	BeforeEach(func() {
		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})
})
`
