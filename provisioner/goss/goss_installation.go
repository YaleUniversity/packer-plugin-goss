//go:generate packer-sdc struct-markdown

package goss

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// Installation holds all installation params.
type Installation struct {
	// execute goss validate with sudo permissions
	UseSudo bool `mapstructure:"use_sudo"`

	/// Goss Version to download
	Version string `mapstructure:"version"`

	// Architecture of the target system
	Arch string `mapstructure:"arch"`

	// OS of the target system.
	OS string `mapstructure:"os"`

	// URL to download the goss binary from.
	URL string `mapstructure:"url"`

	// If true SSL checks are skipped.
	SkipSSL bool `mapstructure:"skip_ssl"`

	// Path to download the goss binary to.
	DownloadPath string `mapstructure:"download_path"`

	// Username for basic auth.
	Username string `mapstructure:"username"`

	// Password for basic auth.
	Password string `mapstructure:"password"`

	// EnvVars any env vars.
	EnvVars map[string]string `mapstructure:"env_vars"`

	// If true installation of the goss binary is skipped.
	SkipInstallation bool `mapstructure:"skip_installation"`
}

func (i *Installation) Name() string {
	return "curl/wget installation"
}

func (i *Installation) Validate() error {
	// no sudo on windows
	//nolint: staticcheck
	if i.UseSudo && strings.Title(i.OS) == Windows {
		return fmt.Errorf("sudo is not supported on windows")
	}

	// if no version specified, use latest
	if i.Version == "" {
		i.Version = "latest"
	} else {
		// remove any leading v as that is already in the DownloadURL
		i.Version = strings.TrimPrefix(i.Version, "v")
	}

	// if no arch specified, assume linux_x86
	if i.Arch == "" {
		i.Arch = DefaultArch
	}

	// if no OS specified, assume linux_x86
	//nolint: staticcheck
	if i.OS == "" {
		i.OS = DefaultOS
	} else if !slices.Contains(validOS, strings.Title(i.OS)) {
		return fmt.Errorf("invalid OS. Valid options: %v", validOS)
	}

	if i.URL == "" {
		if i.Version == Latest {
			i.URL = fmt.Sprintf(LatestVersionDownloadURL, i.OS, i.Arch)
		} else {
			i.URL = fmt.Sprintf(VersionDownloadURL, i.Version, i.OS, i.Arch)
		}
	}

	if i.DownloadPath == "" {
		i.DownloadPath = fmt.Sprintf(DefaultDownloadPath, i.Version, i.OS, i.Arch)
	}

	return nil
}

func (i *Installation) Run(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	if i.SkipInstallation {
		ui.Say(fmt.Sprintf("Skipping %s", i.Name()))

		return nil
	}

	// create download paths
	ui.Say(fmt.Sprintf("Creating download path \"%s\" ...", filepath.Dir(i.DownloadPath)))
	mkDirsCmd := &packer.RemoteCmd{Command: fmt.Sprintf(mkDirCmd, filepath.Dir(i.DownloadPath))}

	err := mkDirsCmd.RunWithUi(ctx, comm, ui)
	if err != nil || mkDirsCmd.ExitStatus() != 0 {
		return fmt.Errorf("error creating directories: \"%s\"", filepath.Dir(i.DownloadPath))
	}

	// download
	ui.Say(fmt.Sprintf("Installing goss version %s from %s", i.Version, i.URL))
	downloadCmd := &packer.RemoteCmd{Command: i.String()}

	err = downloadCmd.RunWithUi(ctx, comm, ui)
	if err != nil || downloadCmd.ExitStatus() != 0 {
		return fmt.Errorf("unable to download goss: %w", err)
	}

	// make executable, and test invocation
	ui.Say("Trying to invoke goss ...")
	installCmd := &packer.RemoteCmd{Command: fmt.Sprintf(InstallCmd, i.DownloadPath, i.DownloadPath)}

	err = installCmd.RunWithUi(ctx, comm, ui)
	if err != nil || installCmd.ExitStatus() != 0 {
		return fmt.Errorf("unable to install goss: %w", err)
	}

	return nil
}

// nolint: cyclop
func (i *Installation) String() string {
	return SanitizeCommands(fmt.Sprintf(
		DownloadCmd,
		// curl

		// env vars
		ExportEnvVars(i.EnvVars, i.OS),

		// sudo
		func() string {
			if i.UseSudo {
				return "sudo"
			}

			return ""
		}(),

		// ssl flag curl
		func() string {
			if i.SkipSSL {
				return "-k"
			}

			return ""
		}(),

		// basic auth curl
		func() string {
			basicAuth := ""

			if i.Username != "" {
				basicAuth += fmt.Sprintf("-u=\"%s\"", i.Username)
			}

			if i.Password != "" {
				basicAuth += fmt.Sprintf(":\"%s\"", i.Password)
			}

			return basicAuth
		}(),

		// output path curl
		i.DownloadPath,

		// download url
		i.URL,

		// wget
		// env vars
		ExportEnvVars(i.EnvVars, i.OS),

		// sudo
		func() string {
			if i.UseSudo {
				return "sudo"
			}

			return ""
		}(),

		// ssl flag wget
		func() string {
			if i.SkipSSL {
				return "--no-check-certificate"
			}

			return ""
		}(),

		// basic auth wget
		func() string {
			basicAuth := ""

			if i.Username != "" {
				basicAuth += fmt.Sprintf("--user=\"%s\" ", i.Username)
			}

			if i.Password != "" {
				basicAuth += fmt.Sprintf("--password=\"%s\" ", i.Password)
			}

			return basicAuth
		}(),

		// output path curl
		i.DownloadPath,

		// download url
		i.URL,
	))
}
