module github.com/onsi/ginkgo

go 1.16

require (
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0
	github.com/onsi/gomega v1.11.1-0.20210307213111-c3c09204ab54
	golang.org/x/sys v0.0.0-20210112080510-489259a85091
	golang.org/x/tools v0.0.0-20201224043029-2b0845dc783e
)

retract v1.16.3 // git tag accidentally associated with incorrect git commit
