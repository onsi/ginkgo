module example.com/dependency_fetcher

go 1.16

require github.com/onsi/gomega v1.13.0

require github.com/gorilla/mux v1.8.0

require gopkg.in/yaml.v2 v2.4.0

require (
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/tdewolff/minify/v2 v2.9.17
	golang.org/x/crypto v0.17.0
)

replace github.com/onsi/ginkgo => ../../ginkgo
