# Goss Packer Provisioner

Wouldn't it be nice if you could run [goss](https://github.com/aelsabbahy/goss) tests against an image during a packer build?

Well, I thought it would, so now you can!  This currently only works for building a `linux` image since goss only runs in linux.

This runs during the provisioning process since the machine being provisioned is only available at that time.

There is an example packer build with goss tests in the `example/` directory.

## Configuration

```json
"provisioners" : [
  {
    "type": "goss",
    "tests": [
      "goss/goss.yaml"
    ]
  }
]
```

### Additional (optional) properties

```json
"provisioners" : [
  {
    "type": "goss",
    "version": "0.3.2",
    "arch": "amd64",
    "url":"https://github.com/aelsabbahy/goss/releases/download/vVERSION/goss-linux-ARCH",
    "tests": [
      "goss/goss.yaml"
    ],
    "downloadPath": "/tmp/goss-VERSION-linux-ARCH",
    "remote_folder": "/tmp",
    "remote_path": "/tmp/goss",
    "skipInstall": false,
    "skip_ssl": false,
    "use_sudo": false,
    "goss_file": "",
    "username": "",
    "password": "",
    "debug": false
  }
]
```

## Author

E. Camden Fisher <camden.fisher@yale.edu>

## License

### MIT

Copyright 2017 Yale University

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
