package cloudsigma

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
}

func NewDriver(hostName, storePath string) *Driver {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) Create() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) DriverName() string {
	return "cloudsigma"
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) GetIP() (string, error) {
	//TODO: see libmachine/drivers/drivers.go
	return "", nil
}

func (d *Driver) GetMachineName() string {
	//TODO: see libmachine/drivers/drivers.go
	return ""
}

func (d *Driver) GetSSHHostname() (string, error) {
	//TODO: see libmachine/drivers/drivers.go
	return "", nil
}

func (d *Driver) GetSSHKeyPath() string {
	//TODO: see libmachine/drivers/drivers.go
	return ""
}

func (d *Driver) GetSSHPort() (int, error) {
	//TODO: see libmachine/drivers/drivers.go
	return 0, nil
}

func (d *Driver) GetSSHUsername() string {
	//TODO: see libmachine/drivers/drivers.go
	return ""
}

func (d *Driver) GetURL() (string, error) {
	//TODO: see libmachine/drivers/drivers.go
	return "", nil
}

func (d *Driver) GetState() (state.State, error) {
	//TODO: see libmachine/drivers/drivers.go
	return state.None, nil
}

func (d *Driver) Kill() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) PreCreateCheck() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) Remove() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) Restart() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) Start() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) Stop() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}
