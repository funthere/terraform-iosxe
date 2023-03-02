#!/bin/bash

VERSION=$(cat version)
echo "building terraform-provider-dd"
go build -o terraform-provider-dd
# rm ~/.terraform.d/plugins/ductus/local/dd/0.0.2/linux_amd64/terraform-provider-dd
mv terraform-provider-dd ~/.terraform.d/plugins/ductus/local/dd/0.0.2/linux_amd64/
echo "done"