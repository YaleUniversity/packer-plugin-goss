default: help

.PHONY: help
help: ## list makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: ## format go files
	gofumpt -w .
	gci write .
	packer fmt -recursive -write .

.PHONY: build 
build: ## build the plugin
	go build -ldflags="-X github.com/YaleUniversity/packer-provisioner-goss/version.VersionPrerelease=dev" -o packer-plugin-goss

.PHONY: install
install: ## install the plugin
	packer plugins install --path packer-plugin-goss github.com/YaleUniversity/goss

.PHONY: local
local: clean build install ## build and install the plugin locally
	cd example && packer init .
	cd example && packer build alpine.pkr.hcl

.PHONY: clean
clean: ## remove tmp files
	rm -f  example/alpine.tar example/goss_test_results.xml example/debug-goss-spec.yaml example/goss-spec.yaml packer-plugin-goss

.PHONY: generate
generate: ## go generate
	go generate ./...