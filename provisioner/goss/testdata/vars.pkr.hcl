source "docker" "alpine" {
  image       = "alpine"
  export_path = "alpine.tar"
}

build {
  sources = ["docker.alpine"]
  provisioner "goss" {
    installation {}
    validate {
      goss_file = "./testdata/goss_tmpl.yaml"

      vars_file = "./testdata/vars.yaml"
      vars_inline = {
        installed = true
      }

      env_vars = {
        curl = true
      }
    }
  }
}
