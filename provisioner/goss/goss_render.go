// //go:generate packer-sdc struct-markdown

package goss

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"path"
// 	"path/filepath"
// 	"slices"

// 	"github.com/hashicorp/packer-plugin-sdk/packer"
// )

// // Render holds all goss render params.
// type Render struct {
// 	remotePath     string
// 	gossBinaryPath string

// 	// execute goss render with sudo permissions
// 	UseSudo bool `mapstructure:"use_sudo"`

// 	// if true enabling goss render debug mode.
// 	Debug bool `mapstructure:"debug"`

// 	// Path to the goss file
// 	GossFile string `mapstructure:"goss_file"`

// 	// Path to the vars file
// 	VarsFile string `mapstructure:"vars_file"`

// 	// map of vars to render the goss file with.
// 	VarsInline map[string]string `mapstructure:"vars_inline"`

// 	// EnvVars any env vars.
// 	EnvVars map[string]string `mapstructure:"env_vars"`

// 	// Path to write the goss output to.
// 	OutputFile string `mapstructure:"output_file"`

// 	// If true goss rendering is skipped.
// 	SkipRender bool `mapstructure:"skip"`
// }

// func (r *Render) Name() string {
// 	return "goss render"
// }

// func (r *Render) Validate() error {
// 	if r.gossBinaryPath == "" {
// 		r.gossBinaryPath = fmt.Sprintf(DefaultDownloadPath, Latest, DefaultOS, DefaultArch)
// 	}

// 	if r.OutputFile == "" {
// 		r.OutputFile = DefaultGossSpecFile
// 	}

// 	if r.GossFile == "" {
// 		r.GossFile = DefaultGossFile
// 	}

// 	if r.remotePath == "" {
// 		r.remotePath = DefaultRemotePath
// 	}

// 	return nil
// }

// // nolint: cyclop
// func (r *Render) Run(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
// 	if r.SkipRender {
// 		ui.Message(fmt.Sprintf("Skipping %s", r.Name()))

// 		return nil
// 	}

// 	files := []string{r.GossFile}

// 	if r.VarsFile != "" {
// 		files = append(files, r.VarsFile)
// 	}

// 	ui.Message(fmt.Sprintf("uploading %v to target system ....", files))

// 	for _, src := range files {
// 		s, err := os.Stat(src)
// 		if err != nil {
// 			return fmt.Errorf("error stating file: %w", err)
// 		}

// 		if !s.Mode().IsRegular() {
// 			return fmt.Errorf("file is not regular: %s", src)
// 		}

// 		f, err := os.Open(src)
// 		if err != nil {
// 			return fmt.Errorf("error opening file: %w", err)
// 		}

// 		defer f.Close()

// 		dst := path.Join(r.remotePath, filepath.Base(src))

// 		ui.Message(fmt.Sprintf("Uploading \"%s\" to \"%s\"", src, dst))

// 		if err := comm.Upload(dst, f, nil); err != nil {
// 			return fmt.Errorf("error uploading file \"%s\" to \"%s\": %w", src, dst, err)
// 		}
// 	}

// 	ui.Message(fmt.Sprintf("Running \"%s\"", r))

// 	renderCmd := &packer.RemoteCmd{Command: r.String()}

// 	if err := renderCmd.RunWithUi(ctx, comm, ui); err != nil {
// 		return err
// 	}

// 	if r.Debug {
// 		ui.Message("rendered goss spec:")

// 		outputCmd := &packer.RemoteCmd{Command: fmt.Sprintf("cat %s", path.Join(r.remotePath, r.OutputFile))}

// 		if err := outputCmd.RunWithUi(ctx, comm, ui); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (r *Render) String() string {
// 	return SanitizeCommands(fmt.Sprintf(
// 		RenderCmd,

// 		// remote path
// 		r.remotePath,

// 		// env vars
// 		func() string {
// 			if len(r.EnvVars) == 0 {
// 				return ""
// 			}

// 			envs := ""

// 			keys := MapKeys(r.EnvVars)
// 			slices.Sort(keys)

// 			for _, k := range keys {
// 				envs += fmt.Sprintf("%s=%s ", k, r.EnvVars[k])
// 			}

// 			return "export " + envs + ";"
// 		}(),

// 		// sudo
// 		func() string {
// 			if r.UseSudo {
// 				return "sudo"
// 			}

// 			return ""
// 		}(),

// 		// goss binary path
// 		r.gossBinaryPath,

// 		// goss file
// 		func() string {
// 			if r.GossFile != "" {
// 				return fmt.Sprintf("--gossfile=\"%s\"", r.GossFile)
// 			}

// 			return ""
// 		}(),

// 		// goss vars
// 		func() string {
// 			if r.VarsFile != "" {
// 				return fmt.Sprintf("--vars=\"%s\"", r.VarsFile)
// 			}

// 			return ""
// 		}(),

// 		// vars inline
// 		func() string {
// 			if len(r.VarsFile) == 0 {
// 				return ""
// 			}

// 			vars := ""
// 			keys := MapKeys(r.VarsInline)
// 			slices.Sort(keys)

// 			for _, k := range keys {
// 				vars += fmt.Sprintf("--vars-inline='%s: %s' ", k, r.VarsInline[k])
// 			}

// 			return vars
// 		}(),

// 		// debug
// 		func() string {
// 			if r.Debug {
// 				return "-d"
// 			}

// 			return ""
// 		}(),

// 		// spec file
// 		path.Join(r.remotePath, r.OutputFile),
// 	))
// }
