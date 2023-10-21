package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"

	"github.com/YaleUniversity/packer-provisioner-goss/provisioner/goss"
)

var (
	Version = "0.0.1"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterProvisioner(plugin.DEFAULT_NAME, new(goss.Provisioner))
	pps.SetVersion(version.InitializePluginVersion(Version, ""))

	err := pps.Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
