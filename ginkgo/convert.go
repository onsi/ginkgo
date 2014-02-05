package main

import (
	"fmt"
	"github.com/onsi/ginkgo/ginkgo/convert"
	"os"
)

func convertPackage() {
	if len(os.Args) != 3 {
		println(fmt.Sprintf("usage: %s convert /path/to/your/package", os.Args[2]))
		os.Exit(1)
	}

	defer func() {
		err := recover()
		if err != nil {
			switch err := err.(type) {
			case error:
				println(err.Error())
			case string:
				println(err)
			default:
				println(fmt.Sprintf("unexpected error: %#v", err))
			}
			os.Exit(1)
		}
	}()

	convert.RewritePackage(os.Args[2])
}
