E2E_DIR ?= examples

default: help

.PHONY: help
help: ## list makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

PHONY: lint
lint: ## lint go files
	golangci-lint run -c .golang-ci.yml

.PHONY: fmt
fmt: ## format go files
	gofumpt -w .
	gci write .
	packer fmt -recursive -write .

.PHONY: build
build: ## build the plugin
	go build -ldflags="-X github.com/YaleUniversity/packer-provisioner-goss/version.VersionPrerelease=dev" -o packer-plugin-goss

.PHONY: install
install: build ## install the plugin
	packer plugins install --path packer-plugin-goss github.com/YaleUniversity/goss

.PHONY: test
test: ## run tests
	PACKER_ACC=1 gotestsum

.PHONY: test-acc
test-acc: clean build install ## run acceptance tests
	PACKER_ACC=1 go test -count 1 -v ./provisioner/goss/provisioner_goss_test.go  -timeout=120m

.PHONY: test-e2e
test-e2e: clean build install ## run e2e tests
	cd $(E2E_DIR) && packer init .
	cd $(E2E_DIR) && packer build .

.PHONY: clean
clean: ## remove tmp files
	rm -f $(E2E_DIR)/*.tar $(E2E_DIR)/test-results.xml packer-plugin-goss

.PHONY: generate
generate: ## go generate
	go generate ./...

.PHONY: plugin-check
plugin-check: build ## will check whether a plugin binary seems to work with packer
	@packer-sdc plugin-check packer-plugin-goss

.PHONY: docs
docs: ## gen packer plugin docs
	@go generate ./...
	@rm -rf .docs
	@packer-sdc renderdocs -src "docs" -partials docs-partials/ -dst ".docs/"
	@./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "YaleUniversity"
	@rm -r ".docs"

.PHONY: example
example: install ## run example
	cd examples && packer build .
