# Goss Packer Provisioner
Wouldn't it be nice if you could run [goss](https://github.com/aelsabbahy/goss) tests against an image during a packer build?
Well, I thought it would, so now you can!  

This runs during the provisioning process since the machine being provisioned is only available at that time.
There is an example packer build with goss tests in the `example/` directory.

## Configuration

```hcl
packer {
  required_version = ">= 1.9.0"

  required_plugins {
    goss = {
      version = "3.2.0"
      source  = "github.com/YaleUniversity/goss"
    }
  }
}
```

### Additional (optional) properties

```hcl
build {
  sources = [".."]
      
  provisioner "goss" {
    # Provisioner Args
    arch ="amd64" 
    download_path = "/tmp/goss-VERSION-linux-ARCH"
    inspect = "{{ inspect_mode }}",
    password = ""
    skip_install = false
    url = "https://github.com/aelsabbahy/goss/releases/download/vVERSION/goss-linux-ARCH"
    username = ""
    version = "0.3.2"

    # GOSS Args
    tests = [
      "goss/goss.yaml"
    ]

    remote_folder = "/tmp"
    remote_path  = "/tmp/goss"
    skip_ssl = false
    use_sudo = false
    format = ""
    goss_file = ""
    vars_file  = ""
    target_os = "Linux"

    vars_env = {
      ARCH = "amd64"
      PROVIDER = "{{ cloud-provider }}"
    }

    vars_inline = {
      OS = "centos",
      version = "{{ version }}"
    }

    retry_timeout = "0s"
    sleep = "1s"
  }
}
```

## Spec files
Goss spec file and debug spec file (`goss render -d`) are downloaded to `/tmp` folder on local machine from the remote VM. These files are exact specs GOSS validated on the VM. The downloaded GOSS spec can be used to validate any other VM image for equivalency.  

## Windows support

This now has support for Windows. Set the optional parameter `target_os` to `Windows`. Currently, the `vars_env` parameter must include `GOSS_USE_ALPHA=1` as specified in [goss's feature parity document](https://github.com/aelsabbahy/goss/blob/master/docs/platform-feature-parity.md#platform-feature-parity).  In the future when goss come of of alpha for Windows this parameter will not be required.

## Build

### Using Golang docker image

```bash
docker run --rm -it -v "$PWD":/usr/src/packer-provisioner-goss -w /usr/src/packer-provisioner-goss -e 'VERSION=v1.0.0' golang:1.13 bash
go test ./...
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    export GOOS GOARCH
      go get -v ./...
      go build -v -o packer-provisioner-goss-${VERSION}-$GOOS-$GOARCH
  done
done
```

## Author

E. Camden Fisher <camden.fisher@yale.edu>

## License

### MIT

Copyright 2017-2021 Yale University

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
