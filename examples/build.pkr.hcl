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
