//go:build tools

package tools

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

//go:generate go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint
//go:generate go install mvdan.cc/gofumpt
//go:generate go install github.com/daixiang0/gci
//go:generate go install gotest.tools/gotestsum
//go:generate go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc
import (
	// gci
	_ "github.com/daixiang0/gci"
	// golangci-lint
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	// packer-plugin-sdk
	_ "github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc"
	// gotestsum
	_ "gotest.tools/gotestsum"
	// gofumpt
	_ "mvdan.cc/gofumpt"
)
