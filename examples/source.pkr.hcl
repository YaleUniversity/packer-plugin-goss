# fetch a normal alpine container and export the image as alpine.tar
source "docker" "linux" {
  image       = "alpine"
  export_path = "linux.tar"
  platform    = "linux/amd64"
}
