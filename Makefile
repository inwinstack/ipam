VERSION_MAJOR  ?= 0
VERSION_MINOR  ?= 7
VERSION_BUILD  ?= 0
VERSION_TSTAMP ?= $(shell date -u +%Y%m%d-%H%M%S)
VERSION_SHA    ?= $(shell git rev-parse --short HEAD)
VERSION ?= v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)-$(VERSION_TSTAMP)-$(VERSION_SHA)

GOOS ?= $(shell go env GOOS)

ORG := github.com
OWNER := xenolog
REPOPATH ?= $(ORG)/$(OWNER)/ipam

$(shell mkdir -p ./out)

.PHONY: build
build: out/controller

.PHONY: out/controller
out/controller:
	GOOS=$(GOOS) go build \
	  -ldflags="-s -w -X $(REPOPATH)/pkg/version.version=$(VERSION)" \
	  -a -o $@ cmd/main.go

.PHONY: test
test:
	./hack/test-go.sh

.PHONY: build_image
build_image:
	docker build -t $(OWNER)/ipam:$(VERSION) .

.PHONY: push_image
push_image:
	docker push $(OWNER)/ipam:$(VERSION)

.PHONY: clean
clean:
	rm -rf out/

