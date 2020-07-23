//go:generate mapstructure-to-hcl2 -type GossConfig

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"

	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/hashicorp/packer/template/interpolate"
)

// GossConfig holds the config data coming in from the packer template
type GossConfig struct {
	// Goss installation
	Version      string
	Arch         string
	URL          string
	DownloadPath string
	Username     string
	Password     string
	SkipInstall  bool
	Inspect      bool

	// An array of tests to run.
	Tests []string

	// Goss options for retry and timeouts
	RetryTimeout string `mapstructure:"retry_timeout"`
	Sleep        string `mapstructure:"sleep"`

	// Use Sudo
	UseSudo bool `mapstructure:"use_sudo"`

	// skip ssl check flag
	SkipSSLChk bool `mapstructure:"skip_ssl"`

	// The --gossfile flag
	GossFile string `mapstructure:"goss_file"`

	// The --vars flag
	// Optional file containing variables, used within GOSS templating.
	// Must be one of the files contained in the Tests array.
	// Can be YAML or JSON.
	VarsFile string `mapstructure:"vars_file"`

	// The --vars-inline flag
	// Optional inline variables that overrides JSON file vars
	VarsInline map[string]string `mapstructure:"vars_inline"`

	// The remote folder where the goss tests will be uploaded to.
	// This should be set to a pre-existing directory, it defaults to /tmp
	RemoteFolder string `mapstructure:"remote_folder"`

	// The remote path where the goss tests will be uploaded.
	// This defaults to remote_folder/goss
	RemotePath string `mapstructure:"remote_path"`

	// The format to use for test output
	// Available: [documentation json json_oneline junit nagios nagios_verbose rspecish silent tap]
	// Default:   rspecish
	Format string `mapstructure:"format"`

	// The format options to use for printing test output
	// Available: [perfdata verbose pretty]
	// Default:   verbose
	FormatOptions string `mapstructure:"format_options"`

	ctx interpolate.Context
}

var validFormats = []string{"documentation", "json", "json_oneline", "junit", "nagios", "nagios_verbose", "rspecish", "silent", "tap"}
var validFormatOptions = []string{"perfdata", "verbose", "pretty"}

// Provisioner implements a packer Provisioner
type Provisioner struct {
	config GossConfig
}

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterProvisioner(new(Provisioner))
	server.Serve()
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

// Prepare gets the Goss Privisioner ready to run
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

	if p.config.Version == "" {
		p.config.Version = "0.3.9"
	}

	if p.config.Arch == "" {
		p.config.Arch = "amd64"
	}

	if p.config.URL == "" {
		p.config.URL = fmt.Sprintf(
			"https://github.com/aelsabbahy/goss/releases/download/v%s/goss-linux-%s",
			p.config.Version, p.config.Arch)
	}

	if p.config.DownloadPath == "" {
		if p.config.URL == "" {
			p.config.DownloadPath = fmt.Sprintf("/tmp/goss-%s-linux-%s", p.config.Version, p.config.Arch)
		} else {
			list := strings.Split(p.config.URL, "/")
			arch := strings.Split(list[len(list)-1], "-")[2]
			version := strings.TrimPrefix(list[len(list)-2], "v")
			p.config.DownloadPath = fmt.Sprintf("/tmp/goss-%s-linux-%s", version, arch)
		}
	}

	if p.config.RemoteFolder == "" {
		p.config.RemoteFolder = "/tmp"
	}

	if p.config.RemotePath == "" {
		p.config.RemotePath = fmt.Sprintf("%s/goss", p.config.RemoteFolder)
	}

	if p.config.Tests == nil {
		p.config.Tests = make([]string, 0)
	}

	if p.config.GossFile != "" {
		p.config.GossFile = fmt.Sprintf("--gossfile %s", p.config.GossFile)
	}

	var errs *packer.MultiError
	if p.config.Format != "" {
		valid := false
		for _, candidate := range validFormats {
			if p.config.Format == candidate {
				valid = true
				break
			}
		}
		if !valid {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Invalid format choice %s. Valid options: %v",
					p.config.Format, validFormats))
		}
	}

	if p.config.FormatOptions != "" {
		valid := false
		for _, candidate := range validFormatOptions {
			if p.config.FormatOptions == candidate {
				valid = true
				break
			}
		}
		if !valid {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Invalid format options choice %s. Valid options: %v",
					p.config.FormatOptions, validFormatOptions))
		}
	}

	if len(p.config.Tests) == 0 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("tests must be specified"))
	}

	for _, path := range p.config.Tests {
		if _, err := os.Stat(path); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Bad test '%s': %s", path, err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

// Provision runs the Goss Provisioner
func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Provisioning with Goss")

	if !p.config.SkipInstall {
		if err := p.installGoss(ui, comm); err != nil {
			return fmt.Errorf("Error installing Goss: %s", err)
		}
	} else {
		ui.Message("Skipping Goss installation")
	}

	ui.Say("Uploading goss tests...")
	if err := p.createDir(ui, comm, p.config.RemotePath); err != nil {
		return fmt.Errorf("Error creating remote directory: %s", err)
	}

	if p.config.VarsFile != "" {
		vf, err := os.Stat(p.config.VarsFile)
		if err != nil {
			return fmt.Errorf("Error stating file: %s", err)
		}
		if vf.Mode().IsRegular() {
			ui.Message(fmt.Sprintf("Uploading vars file %s", p.config.VarsFile))
			varsDest := filepath.ToSlash(filepath.Join(p.config.RemotePath, filepath.Base(p.config.VarsFile)))
			if err := p.uploadFile(ui, comm, varsDest, p.config.VarsFile); err != nil {
				return fmt.Errorf("Error uploading vars file: %s", err)
			}
		}
	}
	if len(p.config.VarsInline) != 0 {
		ui.Message(fmt.Sprintf("Inline variables are %v", p.config.VarsInline))
		ui.Message(fmt.Sprintf("Inline variable string is %s", p.inline_vars()))
	}

	for _, src := range p.config.Tests {
		s, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("Error stating file: %s", err)
		}

		if s.Mode().IsRegular() {
			ui.Message(fmt.Sprintf("Uploading %s", src))
			dst := filepath.ToSlash(filepath.Join(p.config.RemotePath, filepath.Base(src)))
			if err := p.uploadFile(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading goss test: %s", err)
			}
		} else if s.Mode().IsDir() {
			ui.Message(fmt.Sprintf("Uploading Dir %s", src))
			dst := filepath.ToSlash(filepath.Join(p.config.RemotePath, filepath.Base(src)))
			if err := p.uploadDir(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading goss test: %s", err)
			}
		} else {
			ui.Message(fmt.Sprintf("Ignoring %s... not a regular file", src))
		}
	}

	ui.Say("\n\n\nRunning goss tests...")
	if err := p.runGoss(ui, comm); err != nil {
		return fmt.Errorf("Error running Goss: %s", err)
	}

	return nil
}

// installGoss downloads the Goss binary on the remote host
func (p *Provisioner) installGoss(ui packer.Ui, comm packer.Communicator) error {
	ui.Message(fmt.Sprintf("Installing Goss from, %s", p.config.URL))
	ctx := context.TODO()

	cmd := &packer.RemoteCmd{
		// Fallback on wget if curl failed for any reason (such as not being installed)
		Command: fmt.Sprintf(
			"curl -L %s %s -o %s %s || wget %s %s -O %s %s",
			p.sslFlag("curl"), p.userPass("curl"), p.config.DownloadPath, p.config.URL,
			p.sslFlag("wget"), p.userPass("wget"), p.config.DownloadPath, p.config.URL),
	}
	ui.Message(fmt.Sprintf("Downloading Goss to %s", p.config.DownloadPath))
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return fmt.Errorf("Unable to download Goss: %s", err)
	}
	cmd = &packer.RemoteCmd{
		Command: fmt.Sprintf("chmod 555 %s && %s --version", p.config.DownloadPath, p.config.DownloadPath),
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return fmt.Errorf("Unable to install Goss: %s", err)
	}

	return nil
}

// runGoss runs the Goss tests
func (p *Provisioner) runGoss(ui packer.Ui, comm packer.Communicator) error {
	goss := fmt.Sprintf("%s", p.config.DownloadPath)
	ctx := context.TODO()

	strcmd := fmt.Sprintf("cd %s && %s %s %s %s %s validate --retry-timeout %s --sleep %s %s %s",
		p.config.RemotePath, p.enableSudo(), goss, p.config.GossFile,
		p.vars(), p.inline_vars(), p.retryTimeout(), p.sleep(), p.format(), p.formatOptions())
	ui.Message(fmt.Sprintf("Command : %s", strcmd))

	cmd := &packer.RemoteCmd{
		Command: strcmd,
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		// Inspect mode is on. Report failure but don't fail.
		if p.config.Inspect {
			ui.Say(fmt.Sprintf("Goss tests failed"))
			ui.Say(fmt.Sprintf("Inpect mode on : proceeding without failing Packer"))
		} else {
			return fmt.Errorf("goss non-zero exit status")
		}
	} else {
		ui.Say(fmt.Sprintf("Goss tests ran successfully"))
	}
	return nil
}

func (p *Provisioner) retryTimeout() string {
	if p.config.RetryTimeout == "" {
		return "0s" // goss default
	}
	return p.config.RetryTimeout
}

func (p *Provisioner) sleep() string {
	if p.config.Sleep == "" {
		return "1s" // goss default
	}
	return p.config.Sleep
}

func (p *Provisioner) format() string {
	if p.config.Format != "" {
		return fmt.Sprintf("-f %s", p.config.Format)
	}
	return ""
}

func (p *Provisioner) formatOptions() string {
	if p.config.FormatOptions != "" {
		return fmt.Sprintf("-o %s", p.config.FormatOptions)
	}
	return ""
}

func (p *Provisioner) vars() string {
	if p.config.VarsFile != "" {
		return fmt.Sprintf("--vars %s", filepath.ToSlash(filepath.Join(p.config.RemotePath, filepath.Base(p.config.VarsFile))))
	}
	return ""
}

func (p *Provisioner) inline_vars() string {
	if len(p.config.VarsInline) != 0 {
		inlineVarsJson, err := json.Marshal(p.config.VarsInline)
		if err == nil {
			return fmt.Sprintf("--vars-inline '%s'", string(inlineVarsJson))
		} else {
			fmt.Errorf("Error converting inline vars to json string %v", err)
		}
	}
	return ""
}

func (p *Provisioner) sslFlag(cmdType string) string {
	if p.config.SkipSSLChk {
		switch cmdType {
		case "curl":
			return "-k"
		case "wget":
			return "--no-check-certificate"
		default:
			return ""
		}
	}
	return ""
}

// enable sudo if required
func (p *Provisioner) enableSudo() string {
	if p.config.UseSudo {
		return "sudo"
	}
	return ""
}

// Deal with curl & wget username and password
func (p *Provisioner) userPass(cmdType string) string {
	if p.config.Username != "" {
		switch cmdType {
		case "curl":
			if p.config.Password == "" {
				return fmt.Sprintf("-u %s", p.config.Username)
			}
			return fmt.Sprintf("-u %s:%s", p.config.Username, p.config.Password)
		case "wget":
			if p.config.Password == "" {
				return fmt.Sprintf("--user=%s", p.config.Username)
			}
			return fmt.Sprintf("--user=%s --password=%s", p.config.Username, p.config.Password)
		default:
			return ""
		}
	}
	return ""
}

// createDir creates a directory on the remote server
func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	ctx := context.TODO()

	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p '%s'", dir),
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("non-zero exit status")
	}
	return nil
}

// uploadFile uploads a file
func (p *Provisioner) uploadFile(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Error opening: %s", err)
	}
	defer f.Close()

	if err = comm.Upload(dst, f, nil); err != nil {
		return fmt.Errorf("Error uploading %s: %s", src, err)
	}
	return nil
}

// uploadDir uploads a directory
func (p *Provisioner) uploadDir(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	var ignore []string
	if err := p.createDir(ui, comm, dst); err != nil {
		return err
	}

	if src[len(src)-1] != '/' {
		src = src + "/"
	}
	return comm.UploadDir(dst, src, ignore)
}
