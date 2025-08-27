//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Installation,Validate
//go:generate packer-sdc struct-markdown
package goss

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	Installation Installation `mapstructure:"installation" required:"false"`
	Validate     Validate     `mapstructrue:"validate"`

	ctx interpolate.Context
}

// Provisioner implements a packer Provisioner.
type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

// Prepare gets the Goss Provisioner ready to run.
func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if err := ValidateBlocks(&p.config.Installation, &p.config.Validate); err != nil {
		return err
	}

	p.config.Validate.gossBinaryPath = p.config.Installation.DownloadPath
	p.config.Validate.targetOS = p.config.Installation.OS

	return nil
}

// Provision runs the Goss Provisioner.
func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Starting packer provisioner goss ...")
	ui.Say(fmt.Sprintf("Configured to run on target system %s/%s", p.config.Installation.OS, p.config.Installation.Arch))

	return RunBlocks(ctx, ui, comm, &p.config.Installation, &p.config.Validate)
}
