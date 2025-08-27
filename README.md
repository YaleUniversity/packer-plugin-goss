# Goss Packer Provisioner
Wouldn't it be nice if you could run [goss](https://github.com/goss-org/goss) tests against an image during a packer build?
Well, I thought it would, so now you can!

This runs during the provisioning process since the machine being provisioned is only available at that time.

## Configuration
```hcl
packer {
  required_version = ">= 1.9.0"

  required_plugins {
    goss = {
      version = "~> 4"
      source  = "github.com/YaleUniversity/goss"
    }
  }
}
```

## Usage
```hcl
build {
  sources = [""]

  provisioner "goss" {

    # block specifying any goss download parameters. Goss is downloaded using curl and wget as a fallback installation method
    installation {
      # wether to use sudo for the curl/wget download; # optional; default:  false
      use_sudo = false

      # the goss version to download; optional; default: "latest"
      version = "latest"

      # the architecture to download; optional; default: "amd64"
      arch = "amd64"

      # the operating system to download; optional; default: "Linux", options: "Windows", "Linux"
      os = ""

      # the url to download goss from; optional; default: "https://github.com/goss-org/goss/releases/download/{{ Version }}/goss-{{ Version }}-{{ Os }}-{{ Arch }}"
      url = ""

      # the checksum to verify the downloaded goss binary; optional; default: false
      skip_ssl = false

      # the path to download goss to; optional; default: "/tmp/goss-{{ Version }}-{{ Os }}-{{ Arch }}"
      download_path = ""

      # username for basic auth; optional; default: ""
      username = ""

      # password for basic auth; optional; default: ""
      password = ""

      # a map of any extra env vars to pass to the download request; optional; default: {}
      env_vars = nil

      # wether to skip the installation
      skip_installation = false
    }

    # block specifying any goss validate parameters
    validate {
      # wether to use sudo for the goss validate command; # optional; default:  false
      use_sudo = false

      # the path to cd into before running goss validate; optional; default: "/tmp/"
      remote_path = ""

      # a goss vars file; optional; default: ""
      vars_file = ""

      # a gossfile; optional; default: "./goss.yaml"
      goss_file = ""

      # a map of any goss inline vars for rendering a gossfile; optional; default: {}
      vars_inline = nil

      # a map of any extra env vars to pass to the download request; optional; default: {}
      env_vars = nil

      # where to write the goss test results to; optional; default: ""
      output_file = ""

      # a retry timeout for goss validate; optional; default: "0s"
      retry_timeut = ""

      # a sllep timeout for goss validate; optonal; default: "1s"
      sleep = ""

      # the goss test results format; optional; values: "documentation", "json", "json_oneline", "junit", "nagios", "nagios_verbose", "rspecish", "silent", "tap"
      format = ""

      # the goss test results format options; values; default: "perfdata", "verbose", "pretty"
      format_options = ""
    }
  }
}
```

## Getting Started
This is an example `packer` project using `packers` `docker` builder to build an `alpine` container and running tests against it using the `goss` provisioner.

### 1. Required Plugin definition
Define the required plugins:

```hcl
# examples/packer.pkr.hcl
packer {
  required_version = ">= 1.9.0"

  required_plugins {
    goss = {
      version = "~> 4"
      source  = "github.com/YaleUniversity/goss"
    }
    docker = {
      source  = "github.com/hashicorp/docker"
      version = "v1.0.10"
    }
  }
}
```

### 2. Build Sources
Define a `source.docker.linux` block that fetches an `alpine` container, applies the build specs and exports the container to `linux.tar`:

```hcl
# examples/source.pkr.hcl
# fetch a normal alpine container and export the image as alpine.tar
source "docker" "linux" {
  image       = "alpine"
  export_path = "linux.tar"
}
```

### 3. Build instructions
Define a `build` block, that references the source `docker.linux` and configure the `goss` provisioner to run tests against it.

This example showcases a lot of `goss` features, such as includings, templating and variables.

This blocks specifies to install `goss` in version `0.4.2`. Sets `goss_templ.yaml` as the main `gossfile` and passes a `vars.yaml` as the vars file. Note that `goss_templ.yaml` includes two other `gossfiles` that are also included by the `goss`-provisioner and copied to the target machine. Additionally it sets an inline var `pkg_installed` to `true` and an env var of `curl_installed` to `false`. The results of `goss` are written in `junit` format to `test-results.xml`, which will be downloaded from the target machine to the node that is executing `packer` and written to the specified path.

Please pay attention to the file paths of any included Var or gossfiles. It is recommended to `packer build` from the same directory, where the `packer` files are located

```hcl
# examples/build.pkr.hcl
build {
  # apply build params against the alpine container
  sources = ["docker.linux"]

  # run goss tests using goss provisioner
  provisioner "goss" {

    # goss download parameter
    installation {
      version = "0.4.2"
    }

    # goss validate parameter
    validate {
      goss_file = "./goss_tmpl.yaml"
      vars_file = "./vars.yaml"

      vars_inline = {
        pkg_installed = true
      }

      env_vars = {
        curl_installed = false
      }

      format         = "junit"
      format_options = "pretty"
      output_file    = "test-results.xml"
    }
  }
}
```

### 4. Build and Verify
The final rendered and merged `gossfile` used here, looks like this:

```bash
$> curl_installed=true goss -g goss_tmpl.yaml --vars vars.yaml --vars-inline "pkg_installed: true" render
package:
    ansible:
        installed: true
    curl:
        installed: false
    telnet:
        installed: true
    terraform:
        installed: true
```

Which will test, whether `ansible`, `telnet` and `terraform` are installed and `curl` is absent.

When running `cd examples && packer build .` (or `make example`), you should see `packer` fetching the `alpine` container and applying the build instruction and running the `goss` tests. The `junit` test results file (`examples/test-results.xml`) will be available, as well as the exported `alpine` container (`examples/linux.tar`):

```bash
$> file examples/linux.tar
examples/linux.tar: POSIX tar archive
$> cat examples/test-results.xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="goss" errors="0" tests="4" failures="3" skipped="0" time="0.105" timestamp="2024-10-06T06:02:14Z">
<testcase name="Package terraform installed" time="0.099">
<system-err>Package: terraform: installed: Expected false to equal true</system-err>
<failure>Package: terraform: installed: Expected false to equal true</failure>
</testcase>
<testcase name="Package curl installed" time="0.101">
<system-out>Package: curl: installed: matches expectation: false</system-out>
</testcase>
<testcase name="Package telnet installed" time="0.096">
<system-err>Package: telnet: installed: Expected false to equal true</system-err>
<failure>Package: telnet: installed: Expected false to equal true</failure>
</testcase>
<testcase name="Package ansible installed" time="0.100">
<system-err>Package: ansible: installed: Expected false to equal true</system-err>
<failure>Package: ansible: installed: Expected false to equal true</failure>
</testcase>
```
