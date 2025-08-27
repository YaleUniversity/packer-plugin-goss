package goss

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestAccGossProvisioner(t *testing.T) {
	testCases := []struct {
		template string
		teardown func() error
		err      bool
	}{
		{
			template: "latest.pkr.hcl",
		},
		{
			template: "version.pkr.hcl",
		},
		{
			template: "vars.pkr.hcl",
		},
		{
			template: "format.pkr.hcl",
			teardown: func() error {
				if err := os.Remove("test-results.xml"); err != nil {
					return err
				}

				return nil
			},
		},
		{
			template: "included.pkr.hcl",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.template, func(t *testing.T) {
			tmpl, err := os.ReadFile(fmt.Sprintf("testdata/%s", tc.template))
			if err != nil {
				t.Fatalf("Error reading template: %v", err)
			}

			testCase := &acctest.PluginTestCase{
				Name:     tc.template,
				Template: string(tmpl),
				Type:     "packer-provisioner-goss",
				Teardown: tc.teardown,
				Check: func(buildCommand *exec.Cmd, logfile string) error {
					if buildCommand.ProcessState != nil {
						if buildCommand.ProcessState.ExitCode() != 0 {
							b, _ := os.ReadFile(logfile)

							fmt.Println(stripansi.Strip(string(b)))

							return fmt.Errorf("Bad exit code")
						}
					}
					return nil
				},
			}

			acctest.TestPlugin(t, testCase)
		})
	}
}
