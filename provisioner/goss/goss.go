package goss

import (
	"context"
	"fmt"
	"os"

	gs "github.com/goss-org/goss"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"gopkg.in/yaml.v2"
)

// Block is an interface that all blocks must implement.
type Block interface {
	// Name returns the name of the block.
	Name() string

	// Validate validates the block attributes and sets sane defaults.
	Validate() error

	// Run executes the blocks command.
	Run(ctx context.Context, ui packer.Ui, comm packer.Communicator) error

	// String returns a string representation of the command.
	String() string
}

func ValidateBlocks(blocks ...Block) error {
	var errs *packer.MultiError

	for _, b := range blocks {
		if err := b.Validate(); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("error in %s block: %w", b.Name(), err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func RunBlocks(ctx context.Context, ui packer.Ui, comm packer.Communicator, blocks ...Block) error {
	for _, b := range blocks {
		ui.Say(fmt.Sprintf("Start execution of \"%s\" ...", b.Name()))

		ui.Say(fmt.Sprintf("executing \"%s\"", SanitizeCommands(b.String())))

		if err := b.Run(ctx, ui, comm); err != nil {
			return fmt.Errorf("error running \"%s\": %w", b.Name(), err)
		}

		ui.Say(fmt.Sprintf("Successfully finished \"%s\"", b.Name()))
	}

	return nil
}

// GetIncludedGossFiles returns a list of included goss files (if any)
// which then needs to be uploaded to the target system.
func GetIncludedGossFiles(gossFile string, varsFile string, varsInline map[string]string, envVars map[string]string) ([]string, error) {
	var cfg gs.GossConfig
	var currentTemplateFilter gs.TemplateFilter

	// actually set any env vars, its ok to set these as env vars are inherited down to child processes
	// but we do not spawn any child processes so this wont have any affects on the system
	for k, v := range envVars {
		if err := os.Setenv(k, v); err != nil {
			return nil, fmt.Errorf("cannot set env var %s: %w", k, err)
		}
	}

	varsInlineStr := ""
	for k, v := range varsInline {
		varsInlineStr += fmt.Sprintf("%s: %s\n", k, v)
	}

	currentTemplateFilter, err := gs.NewTemplateFilter(varsFile, varsInlineStr)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(gossFile)
	if err != nil {
		return nil, err
	}

	data, err := currentTemplateFilter(b)
	if err != nil {
		return nil, err
	}

	if yaml.Unmarshal(data, &cfg) != nil {
		return nil, fmt.Errorf("cannot parse gossfile: %w", err)
	}

	if len(cfg.Gossfiles) > 0 {
		files := []string{}

		for k := range cfg.Gossfiles {
			files = append(files, k)
		}

		return files, nil
	}

	return []string{}, nil
}
