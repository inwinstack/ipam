VERSION_MAJOR ?= 0
VERSION_MINOR ?= 4
VERSION_BUILD ?= 0
VERSION ?= v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)

GOOS ?= $(shell go env GOOS)

ORG := github.com
OWNER := inwinstack
REPOPATH ?= $(ORG)/$(OWNER)/ipam-operator

$(shell mkdir -p ./out)

.PHONY: build
build: out/operator

.PHONY: out/operator
out/operator: 
	GOOS=$(GOOS) go build \
	  -ldflags="-X $(REPOPATH)/pkg/version.version=$(VERSION)" \
	  -a -o $@ cmd/main.go

.PHONY: dep 
dep:
	dep ensure

.PHONY: test
test:
	./hack/test-go.sh

.PHONY: build_image
build_image:
	docker build -t $(OWNER)/ipam-operator:$(VERSION) .

.PHONY: push_image
push_image:
	docker push $(OWNER)/ipam-operator:$(VERSION)

.PHONY: clean
clean:
	rm -rf out/

