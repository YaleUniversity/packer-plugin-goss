packer {
  required_version = ">= 1.9.0"

  required_plugins {
    goss = {
      version = "v0.0.1"
      source  = "github.com/YaleUniversity/goss"
    }
    docker = {
      source  = "github.com/hashicorp/docker"
      version = "v1.0.10"
    }
  }
}

# fetch a normal alpine container and export the image as alpine.tar
source "docker" "alpine" {
  image       = "alpine"
  export_path = "alpine.tar"
}

build {
  # apply build params against the alpine container
  sources = ["docker.alpine"]

  # run goss tests using goss provisioner
  provisioner "goss" {
    # download and install goss to /tmp/goss_install
    download_path = "/tmp/goss_install"

    # run goss tests in goss.yaml
    tests = ["./goss.yaml"]

    # output results as junit
    format = "junit"

    # write results to /tmp/goss_test_results.xml, which will be copied to the host
    output_file = "/tmp/goss_test_results.xml"
  }


  # output the test results just for demo purposes
  provisioner "shell-local" {
    inline = ["cat goss_test_results.xml"]
  }
}