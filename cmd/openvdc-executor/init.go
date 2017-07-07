// +build !linux

package main

import (
	_ "github.com/axsh/openvdc/hypervisor/null"
	_ "github.com/axsh/openvdc/hypervisor/esxi"
)
