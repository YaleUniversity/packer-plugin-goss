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
