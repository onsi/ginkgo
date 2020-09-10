package generators

var bootstrapText = `package {{.Package}}

import (
	"testing"

	{{.GinkgoImport}}
	{{.GomegaImport}}
)

func Test{{.FormattedName}}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{.FormattedName}} Suite")
}
`

var agoutiBootstrapText = `package {{.Package}}

import (
	"testing"

	{{.GinkgoImport}}
	{{.GomegaImport}}
	"github.com/sclevine/agouti"
)

func Test{{.FormattedName}}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{.FormattedName}} Suite")
}

var agoutiDriver *agouti.WebDriver

var _ = BeforeSuite(func() {
	// Choose a WebDriver:

	agoutiDriver = agouti.PhantomJS()
	// agoutiDriver = agouti.Selenium()
	// agoutiDriver = agouti.ChromeDriver()

	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(agoutiDriver.Stop()).To(Succeed())
})
`
