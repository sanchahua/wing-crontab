#!/bin/bash

export TEST_XLSOA_CONFIG_SERVER_ADDR="10.10.134.102:8500"
export TEST_XLSOA_CONFIG_SERVER_CONSUL_ADDR="10.10.134.102:8500"

go test -v
#go test -run ConfigCenterLoaderWatch -v

