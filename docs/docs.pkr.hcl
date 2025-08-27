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
      os = "Linux"

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
      env_vars = {}

      # wether to skip the installation
      skip_installation = false
    }

    # block specifying any goss validate parameters
    validate {
      # wether to use sudo for the goss validate command; # optional; default:  false
      use_sudo = false

      # a goss vars file; optional; default: ""
      vars_file = ""

      # a gossfile; optional; default: "./goss.yaml"
      goss_file = ""

      # a map of any goss inline vars for rendering a gossfile; optional; default: {}
      vars_inline = {}

      # a map of any extra env vars to pass to the download request; optional; default: {}
      env_vars = {}

      # loglevel; optional; values: "TRACE", "DEBUG", "INFO", "WARN", "ERROR"
      log_level = ""

      # package type; optional; values: "apk", "dpkg", "pacman", "rpm"
      package = ""

      # a retry timeout for goss validate; optional; default: "0s"
      retry_timeut = ""

      # a sllep timeout for goss validate; optonal; default: "1s"
      sleep = ""

      # the goss test results format; optional; values: "documentation", "json", "json_oneline", "junit", "nagios", "nagios_verbose", "rspecish", "silent", "tap"
      format = ""

      # the goss test results format options; values; default: "perfdata", "verbose", "pretty"
      format_options = ""

      # where to write the goss test results to; optional; default: ""
      output_file = ""
    }
  }
}
