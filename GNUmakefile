all: check

check: vet

vet:
	go vet $(CURDIR)/...

test:
	go test $(CURDIR)/...

.PHONY: all check vet test
