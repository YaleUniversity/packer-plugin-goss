package goss

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// nolint: lll
func TestBlocks(t *testing.T) {
	testCases := []struct {
		name string
		b    Block
		exp  string
		err  bool
	}{
		{
			// TODO: avoid hardcoding the version for latest
			name: "default installation",
			b:    &Installation{},
			exp:  "curl -sL -o /tmp/goss-latest-Linux-amd64 https://github.com/goss-org/goss/releases/latest/download/goss-Linux-amd64 || wget -q -O /tmp/goss-latest-Linux-amd64 https://github.com/goss-org/goss/releases/latest/download/goss-Linux-amd64",
		},
		{
			name: "various parameters",
			b: &Installation{
				Version:      "0.4.2",
				SkipSSL:      true,
				UseSudo:      true,
				Arch:         "amd64",
				OS:           "Linux",
				DownloadPath: "/tmp",
				Username:     "user",
				Password:     "pass",
				EnvVars: map[string]string{
					"FOO": "bar",
				},
			},
			exp: "export FOO=bar ; sudo curl -sL -k -u=\"user\":\"pass\" -o /tmp https://github.com/goss-org/goss/releases/download/v0.4.2/goss-Linux-amd64 || export FOO=bar ; sudo wget -q --no-check-certificate --user=\"user\" --password=\"pass\" -O /tmp https://github.com/goss-org/goss/releases/download/v0.4.2/goss-Linux-amd64",
		},
		{
			name: "forbidden installation",
			b: &Installation{
				UseSudo: true,
				OS:      "Windows",
			},
			err: true,
		},
		// {
		// 	name: "default render",
		// 	b:    &Render{},
		// 	exp:  "cd /tmp && /tmp/goss-latest-Linux-amd64 --gossfile=\"./goss.yaml\" render > /tmp/goss-spec.yaml",
		// },
		// {
		// 	name: "various parameter render",
		// 	b: &Render{
		// 		UseSudo:  true,
		// 		Debug:    true,
		// 		GossFile: "./goss_file.yaml",
		// 		VarsFile: "./vars_file.yaml",
		// 		VarsInline: map[string]string{
		// 			"Foo":  "Bar",
		// 			"Foo2": "Bar",
		// 		},
		// 		EnvVars: map[string]string{
		// 			"FOO": "bar",
		// 		},
		// 		OutputFile: "goss-spec.yaml",
		// 	},
		// 	exp: "cd /tmp && export FOO=bar ; sudo /tmp/goss-latest-Linux-amd64 --gossfile=\"./goss_file.yaml\" --vars=\"./vars_file.yaml\" --vars-inline='Foo: Bar' --vars-inline='Foo2: Bar' render -d > /tmp/goss-spec.yaml",
		// },
		{
			name: "default validate",
			b:    &Validate{},
			exp:  "/tmp/goss-latest-Linux-amd64 --gossfile=\"/tmp/goss.yaml\" validate --retry-timeout=0s --sleep=1s",
		},
		{
			name: "various parameter validate",
			b: &Validate{
				UseSudo:  true,
				GossFile: "./goss_file.yaml",
				VarsFile: "./vars_file.yaml",
				VarsInline: map[string]string{
					"Foo":  "Bar",
					"Foo2": "Bar",
				},
				EnvVars: map[string]string{
					"FOO": "bar",
				},
				targetOS:      Linux,
				RetryTimeout:  "4s",
				Format:        "junit",
				FormatOptions: "perfdata",
				Loglevel:      "TRACE",
				Package:       "rpm",
				OutputFile:    "output.xml",
			},
			exp: "sudo export FOO=bar ; /tmp/goss-latest-Linux-amd64 --package=\"rpm\" --log-level=\"TRACE\" --gossfile=\"/tmp/goss_file.yaml\" --vars=\"/tmp/vars_file.yaml\" --vars-inline='Foo: Bar' --vars-inline='Foo2: Bar' validate --retry-timeout=4s --sleep=1s --format=\"junit\" --format-options=\"perfdata\" | tee \"/tmp/output.xml\"",
		},
	}

	for _, tc := range testCases {
		err := ValidateBlocks(tc.b)

		if tc.err {
			require.Error(t, err, tc.name)

			continue
		} else {
			require.NoError(t, err, tc.name)
		}

		require.Equal(t, tc.exp, tc.b.String(), tc.name)
	}
}

func TestGetIncludedGossFiles(t *testing.T) {
	testCases := []struct {
		name       string
		gossFile   string
		varsFile   string
		varsInline map[string]string
		envVars    map[string]string
		exp        []string
		err        bool
	}{
		{
			name:     "test",
			gossFile: "./testdata/gossfile_included.yaml",
			varsFile: "./testdata/vars.yaml",
			varsInline: map[string]string{
				"installed": "true",
			},
			envVars: map[string]string{
				"installed": "true",
			},
			exp: []string{"./testdata/goss.yaml"},
		},
		{
			name:     "no files",
			gossFile: "./testdata/goss.yaml",
			exp:      []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := GetIncludedGossFiles(tc.gossFile, tc.varsFile, tc.varsInline, tc.envVars)
			if tc.err {
				require.Error(t, err, tc.name)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.exp, files, tc.name)
			}
		})
	}
}
