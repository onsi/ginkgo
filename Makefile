# default task since it's first
.PHONY: all
all: install vet test

.PHONY: install
install:
	which ginkgo 2>&1 >/dev/null || go install ./...

.PHONY: test
test:
	ginkgo -r -p

.PHONY: vet
vet:
	go vet ./...
