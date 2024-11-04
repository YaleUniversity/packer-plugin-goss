source "docker" "alpine" {
  image       = "alpine"
  export_path = "alpine.tar"
}

build {
  sources = ["docker.alpine"]

  provisioner "goss" {
    installation {
      version = "0.4.2"
    }

    validate {
      goss_file = "./testdata/goss.yaml"
    }
  }
}
