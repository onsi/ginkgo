module github.com/onsi/ginkgo

go 1.15

require (
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0
	github.com/nxadm/tail v1.4.8
	github.com/onsi/gomega v1.14.0
	golang.org/x/sys v0.0.0-20210423082822-04245dca01da
	golang.org/x/tools v0.0.0-20201224043029-2b0845dc783e
)

retract v1.16.3 // git tag accidentally associated with incorrect git commit
