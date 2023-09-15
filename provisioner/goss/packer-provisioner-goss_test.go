//go:generate packer-sdc mapstructure-to-hcl2 -type GossConfig

package goss

import (
	"reflect"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func fakeContext() interpolate.Context {
	var data map[interface{}]interface{}
	var funcs map[string]interface{}
	var userVars map[string]string
	var sensitiveVars []string
	return interpolate.Context{
		Data:               data,
		Funcs:              funcs,
		UserVariables:      userVars,
		SensitiveVariables: sensitiveVars,
		EnableEnv:          false,
		BuildName:          "",
		BuildType:          "",
		TemplatePath:       "",
	}
}

func TestProvisioner_Prepare(t *testing.T) {

	var tests = []struct {
		name       string
		input      []interface{}
		wantErr    bool
		wantConfig GossConfig
	}{
		{
			name: "defaults",
			input: []interface{}{
				map[string]interface{}{
					"tests": []string{"../../example/goss"},
				},
			},
			wantErr: false,
			wantConfig: GossConfig{
				Version:       "0.4.2",
				Arch:          "amd64",
				URL:           "https://github.com/goss-org/goss/releases/download/v0.4.2/goss-linux-amd64",
				DownloadPath:  "/tmp/goss-0.4.2-linux-amd64",
				Username:      "",
				Password:      "",
				SkipInstall:   false,
				Inspect:       false,
				TargetOs:      "Linux",
				Tests:         []string{"../../example/goss"},
				RetryTimeout:  "",
				Sleep:         "",
				UseSudo:       false,
				SkipSSLChk:    false,
				GossFile:      "",
				VarsFile:      "",
				VarsInline:    nil,
				VarsEnv:       nil,
				RemoteFolder:  "/tmp",
				RemotePath:    "/tmp/goss",
				Format:        "",
				FormatOptions: "",
				ctx:           fakeContext(),
			},
		},
		{
			name: "Windows",
			input: []interface{}{
				map[string]interface{}{
					"tests":     []string{"../../example/goss"},
					"target_os": "Windows",
					"vars_env": map[string]string{
						"GOSS_USE_ALPHA": "1",
					},
				},
			},
			wantErr: false,
			wantConfig: GossConfig{
				Version:      "0.4.2",
				Arch:         "amd64",
				URL:          "https://github.com/goss-org/goss/releases/download/v0.4.2/goss-alpha-windows-amd64.exe",
				DownloadPath: "/tmp/goss-0.4.2-windows-amd64.exe",
				Username:     "",
				Password:     "",
				SkipInstall:  false,
				Inspect:      false,
				TargetOs:     "Windows",
				Tests:        []string{"../../example/goss"},
				RetryTimeout: "",
				Sleep:        "",
				UseSudo:      false,
				SkipSSLChk:   false,
				GossFile:     "",
				VarsFile:     "",
				VarsInline:   nil,
				VarsEnv: map[string]string{
					"GOSS_USE_ALPHA": "1",
				},
				RemoteFolder:  "/tmp",
				RemotePath:    "/tmp/goss",
				Format:        "",
				FormatOptions: "",
				ctx:           fakeContext(),
			},
		},
		{
			name: "Windows non alpha",
			input: []interface{}{
				map[string]interface{}{
					"tests":     []string{"../../example/goss"},
					"target_os": "Windows",
				},
			},
			wantErr: false,
			wantConfig: GossConfig{
				Version:       "0.4.2",
				Arch:          "amd64",
				URL:           "https://github.com/goss-org/goss/releases/download/v0.4.2/goss-windows-amd64.exe",
				DownloadPath:  "/tmp/goss-0.4.2-windows-amd64.exe",
				Username:      "",
				Password:      "",
				SkipInstall:   false,
				Inspect:       false,
				TargetOs:      "Windows",
				Tests:         []string{"../../example/goss"},
				RetryTimeout:  "",
				Sleep:         "",
				UseSudo:       false,
				SkipSSLChk:    false,
				GossFile:      "",
				VarsFile:      "",
				VarsInline:    nil,
				VarsEnv:       nil,
				RemoteFolder:  "/tmp",
				RemotePath:    "/tmp/goss",
				Format:        "",
				FormatOptions: "",
				ctx:           fakeContext(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{
				config: GossConfig{
					ctx: interpolate.Context{},
				},
			}
			err := p.Prepare(tt.input...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provisioner.Prepare() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && !reflect.DeepEqual(p.config, tt.wantConfig) {
				t.Error("configs do not match")
				t.Logf("got config= %v", p.config)
				t.Logf("want config= %v", tt.wantConfig)
			}

		})
	}
}

func TestProvisioner_envVars(t *testing.T) {

	tests := []struct {
		name   string
		config GossConfig
		want   string
	}{
		{
			name: "Linux",
			config: GossConfig{
				TargetOs: "Linux",
				VarsEnv: map[string]string{
					"somevar": "1",
				},
			},
			want: "somevar=\"1\" ",
		},
		{
			name: "Windows",
			config: GossConfig{
				TargetOs: "Windows",
				VarsEnv: map[string]string{
					"GOSS_USE_ALPHA": "1",
				},
			},
			want: "set \"GOSS_USE_ALPHA=1\" && ",
		},
		{
			name: "no vars windows",
			config: GossConfig{
				TargetOs: "Windows",
				VarsEnv:  map[string]string{},
			},
			want: "",
		},
		{
			name: "no vars linux",
			config: GossConfig{
				TargetOs: "Linux",
				VarsEnv:  map[string]string{},
			},
			want: "",
		},
		{
			name: "no configured target os",
			config: GossConfig{
				VarsEnv: map[string]string{
					"somevar": "1",
				},
			},
			want: "somevar=\"1\" ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{
				config: tt.config,
			}
			if got := p.envVars(); got != tt.want {
				t.Errorf("Provisioner.envVars() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func TestProvisioner_mkDir(t *testing.T) {
	tests := []struct {
		name    string
		config  GossConfig
		dir     string
		wantcmd string
	}{
		{
			name: "linux",
			config: GossConfig{
				TargetOs: linux,
			},
			dir:     "/tmp",
			wantcmd: "mkdir -p '/tmp'",
		},
		{
			name: "windows",
			config: GossConfig{
				TargetOs: windows,
			},
			dir:     "/tmp",
			wantcmd: "powershell /c mkdir -p '/tmp'",
		},
		{
			name:    "no configured os",
			config:  GossConfig{},
			dir:     "/tmp",
			wantcmd: "mkdir -p '/tmp'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{
				config: tt.config,
			}
			if got := p.mkDir(tt.dir); got != tt.wantcmd {
				t.Errorf("Provisioner.mkDir() = %v, want %v", got, tt.wantcmd)
			}
		})
	}
}
