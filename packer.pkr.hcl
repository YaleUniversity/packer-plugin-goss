packer {
  required_version = ">= 1.9.0"

  required_plugins {
    goss = {
      version = "~> 3"
      source  = "github.com/YaleUniversity/goss"
    }
  }
}