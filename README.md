# Goss Packer Provisioner

Wouldn't it be nice if you could run [goss](https://github.com/aelsabbahy/goss) tests against an image during a packer build?

Well, I thought it would, so now you can!  This currently only works for `linux` since goss only runs in linux.

This runs during the provisioning process since the machine being provisioned is only available at that time.

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
    "remoteFolder": "/tmp",
    "remotePath": "/tmp/goss",
    "skipInstall": false,
    "debug": false
  }
]
```

