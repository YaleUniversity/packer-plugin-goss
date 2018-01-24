package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

	// Enable debug for goss (defaults to false)
	Debug bool

	// An array of tests to run.
	Tests []string

	// username for ssl auth

	// skip ssl check flag
	SkipSSLChk   bool `mapstructure:"skip_ssl"`

	// The --gossfile flag
	GossFile string `mapstructure:"goss_file"`

	// The remote folder where the goss tests will be uploaded to.
	// This should be set to a pre-existing directory, it defaults to /tmp
	RemoteFolder string `mapstructure:"remote_folder"`

	// The remote path where the goss tests will be uploaded.
	// This defaults to remote_folder/goss
	RemotePath string `mapstructure:"remote_path"`

	ctx          interpolate.Context
}

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
		p.config.Version = "0.3.2"
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
		p.config.DownloadPath = fmt.Sprintf("/tmp/goss-%s-linux-%s", p.config.Version, p.config.Arch)
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
func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
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

	for _, src := range p.config.Tests {
		s, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("Error stating file: %s", err)
		}

		if s.Mode().IsRegular() == true {
			ui.Message(fmt.Sprintf("Uploading %s", src))
			dst := filepath.ToSlash(filepath.Join(p.config.RemotePath, filepath.Base(src)))
			if err := p.uploadFile(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading goss test: %s", err)
			}
		} else if s.Mode().IsDir() == true {
			ui.Message(fmt.Sprintf("Uploading Dir %s", src))
			dst := filepath.ToSlash(filepath.Join(p.config.RemotePath, filepath.Base(src)))
			if err := p.uploadDir(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading goss test: %s", err)
			}
		} else {
			ui.Message(fmt.Sprintf("Ignoring %s... not a regular file", src))
		}
	}

	ui.Say("Running goss tests...")
	if err := p.runGoss(ui, comm); err != nil {
		return fmt.Errorf("Error running Goss: %s", err)
	}

	return nil
}

// Cancel just exists when provision is cancelled
func (p *Provisioner) Cancel() {
	os.Exit(0)
}

// installGoss downloads the Goss binary on the remote host
func (p *Provisioner) installGoss(ui packer.Ui, comm packer.Communicator) error {
	ui.Message(fmt.Sprintf("Installing Goss from, %s", p.config.URL))

	cmd := &packer.RemoteCmd{
		// Fallback on wget if curl failed for any reason (such as not being installed)
		Command: fmt.Sprintf(
			"curl -L %s %s -o %s %s || wget %s %s -O %s %s",
			p.curlSslFlag(), p.curlUserPass(), p.config.DownloadPath, p.config.URL, p.wgetSslFlag(), p.wgetUserPass(), p.config.DownloadPath, p.config.URL),
	}
	ui.Message(fmt.Sprintf("Downloading Goss to %s", p.config.DownloadPath))
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Unable to download Goss: %s", err)
	}
	cmd = &packer.RemoteCmd{
		Command: fmt.Sprintf("chmod 555 %s && %s --version", p.config.DownloadPath, p.config.DownloadPath),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Unable to install Goss: %s", err)
	}

	return nil
}

// runGoss runs the Goss tests
func (p *Provisioner) runGoss(ui packer.Ui, comm packer.Communicator) error {
	goss := fmt.Sprintf("%s", p.config.DownloadPath)
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf(
			"cd %s && %s %s %s validate", p.config.RemotePath, goss, p.config.GossFile, p.debug()),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("goss non-zero exit status")
	}
	ui.Say(fmt.Sprintf("Goss tests ran successfully"))
	return nil
}

// debug returns the debug flag if debug is configured
func (p *Provisioner) debug() string {
	if p.config.Debug == true {
		return "-d"
	}
	return ""
}

func (p *Provisioner) curlSslFlag() string {
	if p.config.SkipSSLChk == true {
		return "-k"
	}
	return ""
}

func (p *Provisioner) wgetSslFlag() string {
	if p.config.SkipSSLChk == true {
		return "--no-check-certificate"
	}
	return ""
}

// Deal with Curl username and password
func (p *Provisioner) curlUserPass() string {
	if p.config.Username != "" {
		if p.config.Password == "" {
			return fmt.Sprintf("-u %s", p.config.Username)
		}
		return fmt.Sprintf("-u %s:%s", p.config.Username, p.config.Password)
	}
	return ""
}

// Deal with Wget username and password
func (p *Provisioner) wgetUserPass() string {
	if p.config.Username != "" {
		if p.config.Password == "" {
			return fmt.Sprintf("--user=%s", p.config.Username)
		}
		return fmt.Sprintf("--user=%s --password=%s", p.config.Username, p.config.Password)
	}
	return ""
}

// createDir creates a directory on the remote server
func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p '%s'", dir),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus != 0 {
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
