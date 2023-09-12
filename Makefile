GO ?= go
GOLANGCI_LINT ?= $$($(GO) env GOPATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.54.2
DEST ?= "$(shell (go list ./...))"

.PHONY: lint
lint: linter
	$(GOLANGCI_LINT) run

.PHONY: linter
linter:
	test -f $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: release
release:
	docker run --rm --privileged \
		--env-file .release-env \
		-v ~/go/pkg/mod:/go/pkg/mod \
		-v `pwd`:/go/src/github.com/fairDataSociety/fave \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/fairDataSociety/fave \
		ghcr.io/goreleaser/goreleaser-cross:v1.20.2 release --clean

.PHONY: release-dry-run
release-dry-run:
	docker run --rm --privileged \
		-v ~/go/pkg/mod:/go/pkg/mod \
		-v ~/go/bin:/go/bin \
		-v `pwd`:/go/src/github.com/fairDataSociety/fave \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/fairDataSociety/fave \
		ghcr.io/goreleaser/goreleaser-cross:v1.20.2 release --clean \
		--skip-validate=true \
		--skip-publish

FORCE:
