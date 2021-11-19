package ginkgo

import (
	_ "github.com/go-task/slim-sprig"
	_ "golang.org/x/tools/go/ast/inspector"
)

// This file imports the CLI dependencies so that consuming packages have all the dependenceies they need to compile and run the Ginkgo CLI
