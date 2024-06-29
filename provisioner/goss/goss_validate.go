//go:generate packer-sdc struct-markdown

package goss

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// Validate holds all goss validate params.
type Validate struct {
	gossBinaryPath string
	targetOS       string

	// execute goss validate with sudo permissions
	UseSudo bool `mapstructure:"use_sudo"`

	// Path to the goss file
	GossFile string `mapstructure:"goss_file"`

	// Path to the vars file
	VarsFile string `mapstructure:"vars_file"`

	// map of vars to render the goss file with.
	VarsInline map[string]string `mapstructure:"vars_inline"`

	// Package type to use
	Package string `mapstructure:"package"`

	// Loglevel to use
	Loglevel string `mapstructure:"log_level"`

	// retry timeout.
	RetryTimeout string `mapstructure:"retry_timeout"`

	// time to sleep between retries
	Sleep string `mapstructure:"sleep"`

	// Goss Output Format.
	Format string `mapstructure:"format"`

	// Path to write the goss output to.
	OutputFile string `mapstructure:"output_file"`

	// Format options for the output.
	FormatOptions string `mapstructure:"format_options"`

	// EnvVars any env vars.
	EnvVars map[string]string `mapstructure:"env_vars"`
}

func (v *Validate) Name() string {
	return "goss validate"
}

// nolint: cyclop
func (v *Validate) Validate() error {
	if v.gossBinaryPath == "" {
		v.gossBinaryPath = fmt.Sprintf(DefaultDownloadPath, Latest, DefaultOS, DefaultArch)
	}

	if v.Sleep == "" {
		v.Sleep = DefaultSleep
	}

	if v.RetryTimeout == "" {
		v.RetryTimeout = DefaultRetryTimeout
	}

	if v.GossFile == "" {
		v.GossFile = DefaultGossFile
	}

	if v.Loglevel != "" && !slices.Contains(validLogLevel, v.Loglevel) {
		return fmt.Errorf("invalid log level. Valid options: %v", validLogLevel)
	}

	if v.Package != "" && !slices.Contains(ValidPackageTypes, v.Package) {
		return fmt.Errorf("invalid package type. Valid options: %v", ValidPackageTypes)
	}

	if v.Format != "" && !slices.Contains(ValidFormats, v.Format) {
		return fmt.Errorf("invalid format. Valid options: %v", ValidFormats)
	}

	if v.FormatOptions != "" && !slices.Contains(ValidFormatOptions, v.FormatOptions) {
		return fmt.Errorf("invalid format option. Valid options: %v", ValidFormatOptions)
	}

	return nil
}

// nolint: cyclop
func (v *Validate) Run(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	files := []string{v.GossFile}

	ui.Message(fmt.Sprintf("Detecting wether \"%s\" includes other gossfiles ...", v.GossFile))

	gossfiles, err := GetIncludedGossFiles(v.GossFile, v.VarsFile, v.VarsInline, v.EnvVars)
	if err != nil {
		ui.Message(fmt.Sprintf("Error detecting included goss files: %s. Continuing as not fatal", err.Error()))
	} else {
		ui.Message(fmt.Sprintf("Found %v referenced in \"%s\"", gossfiles, v.GossFile))

		files = append(files, gossfiles...)
		if v.VarsFile != "" {
			files = append(files, v.VarsFile)
		}
	}

	ui.Message(fmt.Sprintf("Uploading %v to target system ....", files))

	for _, src := range files {
		s, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("error stating file: %w", err)
		}

		if !s.Mode().IsRegular() {
			return fmt.Errorf("file is not regular: %s", src)
		}

		f, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}

		defer f.Close()

		dst := path.Join(DefaultRemotePath, src)

		// create dirs
		mkDirsCmd := &packer.RemoteCmd{Command: fmt.Sprintf(mkDirCmd, path.Join(DefaultRemotePath, filepath.Dir(src)))}

		err = mkDirsCmd.RunWithUi(ctx, comm, ui)
		if err != nil || mkDirsCmd.ExitStatus() != 0 {
			return fmt.Errorf("error creating directories: %s", path.Join(DefaultRemotePath, filepath.Dir(src)))
		}

		ui.Message(fmt.Sprintf("Uploading \"%s\" to \"%s\"", src, dst))

		if err := comm.Upload(dst, f, nil); err != nil {
			return fmt.Errorf("error uploading file \"%s\" to \"%s\": %w", src, dst, err)
		}
	}

	ui.Message("Running goss validate ...")
	validateCmd := &packer.RemoteCmd{Command: v.String()}

	// test successful -> exit code 0
	// tests fail -> exit code 1
	// gossfile unparsable -> exit code 78
	err = validateCmd.RunWithUi(ctx, comm, ui)
	if err != nil || validateCmd.ExitStatus() != 0 && validateCmd.ExitStatus() != 1 {
		return err
	}

	ui.Message("goss validate finished")

	if v.OutputFile != "" {
		ui.Message(fmt.Sprintf("Downloading goss test result file \"%s\" (target system) to \"%s\" (local) ...", path.Join(DefaultRemotePath, v.OutputFile), filepath.Base(v.OutputFile)))

		if err := DownloadFile(comm, path.Join(DefaultRemotePath, v.OutputFile), filepath.Base(v.OutputFile)); err != nil {
			return fmt.Errorf("error downloading \"%s\": %w", path.Join(DefaultRemotePath, v.OutputFile), err)
		}

		resultsPath := filepath.Base(v.OutputFile)
		if abs, err := filepath.Abs(v.OutputFile); err == nil {
			resultsPath = abs
		}

		// output the absolute path so its easier for folks to upload the file to their CI/CD system as a test artifact ...
		ui.Message(fmt.Sprintf("Successfully downloaded test result file from target system to \"%s\"", resultsPath))
	}

	return nil
}

// nolint: cyclop
func (v *Validate) String() string {
	return SanitizeCommands(fmt.Sprintf(
		ValidateCmd,

		// sudo
		func() string {
			if v.UseSudo {
				return "sudo"
			}

			return ""
		}(),

		// env vars
		ExportEnvVars(v.EnvVars, v.targetOS),

		// goss binary path
		v.gossBinaryPath,

		// package
		func() string {
			if v.Package != "" {
				return fmt.Sprintf("--package=\"%s\"", v.Package)
			}

			return ""
		}(),

		// loglevel
		func() string {
			if v.Package != "" {
				return fmt.Sprintf("--log-level=\"%s\"", v.Loglevel)
			}

			return ""
		}(),
		// goss file
		func() string {
			if v.GossFile != "" {
				return fmt.Sprintf("--gossfile=\"%s\"", path.Join(DefaultRemotePath, v.GossFile))
			}

			return ""
		}(),

		// goss vars
		func() string {
			if v.VarsFile != "" {
				return fmt.Sprintf("--vars=\"%s\"", path.Join(DefaultRemotePath, v.VarsFile))
			}

			return ""
		}(),

		// inline vars
		func() string {
			if len(v.VarsFile) == 0 {
				return ""
			}

			vars := ""

			keys := MapKeys(v.VarsInline)
			slices.Sort(keys)

			for _, k := range keys {
				vars += fmt.Sprintf("--vars-inline='%s: %s' ", k, v.VarsInline[k])
			}

			return vars
		}(),

		// retry timeout
		v.RetryTimeout,

		// sleep
		v.Sleep,

		// format
		func() string {
			if v.Format != "" {
				return fmt.Sprintf("--format=\"%s\"", v.Format)
			}

			return ""
		}(),

		// format options
		func() string {
			if v.FormatOptions != "" {
				return fmt.Sprintf("--format-options=\"%s\"", v.FormatOptions)
			}

			return ""
		}(),

		// output file
		func() string {
			if v.OutputFile != "" {
				if v.targetOS == Linux {
					return fmt.Sprintf("| tee \"%s\"", path.Join(DefaultRemotePath, v.OutputFile))
				}

				if v.targetOS == Windows {
					return fmt.Sprintf("| Tee-Object -FilePath \"%s\"", path.Join(DefaultRemotePath, v.OutputFile))
				}
			}

			return ""
		}(),
	))
}
