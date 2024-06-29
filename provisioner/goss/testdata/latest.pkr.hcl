source "docker" "alpine" {
  image       = "alpine"
  export_path = "alpine.tar"
}

build {
  sources = ["docker.alpine"]

  provisioner "goss" {
    installation {}

    validate {
      goss_file = "./testdata/goss.yaml"
    }
  }
}
