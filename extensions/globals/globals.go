// Package `globals` provides an interface to alter the global state of ginkgo suite.
//
// ginkgo currently registers a few singleton global vars that hold all the
// test blocks and failure management. These vars are global per package, which means
// that only one Suite definition can coexist in one package.
//
// However, there can be some use cases where applications using ginkgo may want to
// have a bit more control about this. For instance, a package may be using ginkgo
// to dynamically generate different tests and groups depending on some configuration.
// In this particular case, if the application wants to test how these different groups
// are generated, they will need access to change these global variables, so they
// can re-generate this global state, and ensure that different configuration generate
// indeed different tests.
//
// Note that this package is not intended to be used as part of normal ginkgo setups, and
// usually, you will never need to worry about the global state of ginkgo
package globals

import "github.com/onsi/ginkgo/internal/global"

// Reset calls `global.InitializeGlobals()` which will basically create a new instance
// of Suite, and therefore, will effectively reset the global variables to the init state.
// This will effectively remove all groups, tests and blocks that were added to the Suite.
func Reset() {
	global.InitializeGlobals()
}
