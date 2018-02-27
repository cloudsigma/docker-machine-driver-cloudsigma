package main

import (
	"github.com/cloudsigma/docker-machine-driver-cloudsigma"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(cloudsigma.NewDriver("", ""))
}
