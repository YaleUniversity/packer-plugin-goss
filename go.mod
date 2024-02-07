module github.com/YaleUniversity/packer-provisioner-goss

go 1.16

require (
	github.com/hashicorp/hcl/v2 v2.19.1
	github.com/hashicorp/packer-plugin-sdk v0.5.2
	github.com/zclconf/go-cty v1.14.2
)

replace github.com/zclconf/go-cty => github.com/nywilken/go-cty v1.13.3 // added by packer-sdc fix as noted in github.com/hashicorp/packer-plugin-sdk/issues/187
