package cloudsigma

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cloudsigma/docker-machine-driver-cloudsigma/api"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

const (
	defaultDriveUUID = "e47f39e0-075f-4b38-83e6-1b9dce36d0f1"
)

type Driver struct {
	*drivers.BaseDriver
	DriveUUID  string
	Password   string
	ServerUUID string
	SSHKeyUUID string
	Username   string
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) Create() error {
	//TODO: create key, clone drive, create server (with keys), attach drive

	log.Infof("Creating SSH key...")
	key, err := d.createSSHKey()
	if err != nil {
		return err
	}
	d.SSHKeyUUID = key.UUID

	log.Infof("Cloning library drive...")
	drive, err := d.cloneDrive(defaultDriveUUID)
	if err != nil {
		return err
	}
	d.DriveUUID = drive.UUID

	log.Infof("Creating CloudSigma server...")
	server, err := d.createServer()
	if err != nil {
		return err
	}
	d.ServerUUID = server.UUID

	log.Debugf("Created server UUID %v, drive UUID %v", d.ServerUUID, d.DriveUUID)

	return nil
}

func (d *Driver) DriverName() string {
	return "cloudsigma"
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	//TODO: enhance with additional values
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_USERNAME",
			Name:   "cloudsigma-username",
			Usage:  "CloudSigma user email",
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_PASSWORD",
			Name:   "cloudsigma-password",
			Usage:  "CloudSigma password",
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_DRIVE",
			Name:   "cloudsigma-drive",
			Usage:  "CloudSigma drive uuid",
			Value:  defaultDriveUUID,
		},
	}
}

func (d *Driver) GetIP() (string, error) {
	//TODO: see libmachine/drivers/drivers.go
	return "", nil
}

//func (d *Driver) GetMachineName() string {
//	return d.GetMachineName()
//}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	if d.SSHKeyPath == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

//func (d *Driver) GetSSHPort() (int, error) {
//	//TODO: see libmachine/drivers/drivers.go
//	return 0, nil
//}

//func (d *Driver) GetSSHUsername() string {
//	//TODO: see libmachine/drivers/drivers.go
//	return ""
//}

func (d *Driver) GetURL() (string, error) {
	//TODO: see libmachine/drivers/drivers.go
	return "", nil
}

func (d *Driver) GetState() (state.State, error) {
	server, _, err := d.getClient().Servers.Get(d.ServerUUID)
	if err != nil {
		return state.Error, err
	}

	switch server.Status {
	case "paused":
		return state.Paused, nil
	case "running":
		return state.Running, nil
	case "starting":
		return state.Starting, nil
	case "stopped":
		return state.Stopped, nil
	case "stopping":
		return state.Stopping, nil
	}

	return state.None, nil
}

func (d *Driver) Kill() error {
	_, _, err := d.getClient().Servers.Stop(d.ServerUUID)
	return err
}

func (d *Driver) PreCreateCheck() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) Remove() error {
	client := d.getClient()

	log.Infof("Deleting CloudSigma SSH key...")
	if resp, err := client.Keypairs.Delete(d.SSHKeyUUID); err != nil {
		if resp.StatusCode == http.StatusNotFound {
			log.Infof("CloudSigma SSH key doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}

	log.Infof("Deleting CloudSigma server...")
	if resp, err := client.Servers.Delete(d.ServerUUID); err != nil {
		if resp.StatusCode == http.StatusNotFound {
			log.Infof("CloudSigma server doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}

	return nil
}

func (d *Driver) Restart() error {
	//TODO: see libmachine/drivers/drivers.go
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	//TODO: see libmachine/drivers/drivers.go
	d.Username = flags.String("cloudsigma-username")
	d.Password = flags.String("cloudsigma-password")

	if d.Username == "" {
		return fmt.Errorf("cloudsigma driver requires the --cloudsigma-username option")
	}

	return nil
}

func (d *Driver) Start() error {
	_, _, err := d.getClient().Servers.Start(d.ServerUUID)
	return err
}

func (d *Driver) Stop() error {
	_, _, err := d.getClient().Servers.Shutdown(d.ServerUUID)
	return err
}

func (d *Driver) getClient() *api.Client {
	return api.NewBasicAuthClient(d.Username, d.Password)
}

func (d *Driver) createSSHKey() (*api.Keypair, error) {
	d.SSHKeyPath = d.GetSSHKeyPath()

	if err := ssh.GenerateSSHKey(d.SSHKeyPath); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.SSHKeyPath + ".pub")
	if err != nil {
		return nil, err
	}

	keypairCreateRequest := &api.KeypairCreateRequest{
		Keypairs: []api.Keypair{
			{Name: d.MachineName, PublicKey: string(publicKey)},
		},
	}

	key, _, err := d.getClient().Keypairs.Create(keypairCreateRequest)
	if err != nil {
		return key, err
	}

	return key, nil
}

func (d *Driver) cloneDrive(uuid string) (*api.Drive, error) {
	driveCloneRequest := &api.DriveCloneRequest{
		Name:        d.MachineName,
		Size:        20 * 1024 * 1024 * 1024,
		StorageType: "dssd",
	}

	client := d.getClient()
	clonedDrive, _, err := client.LibraryDrives.Clone(uuid, driveCloneRequest)
	if err != nil {
		return clonedDrive, err
	}

	log.Debugf("Waiting until cloning process are done...")

	driveUUID := clonedDrive.UUID
	for {
		drive, _, err := client.Drives.Get(driveUUID)
		if err != nil {
			return clonedDrive, nil
		}

		if drive.Status == "unmounted" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return clonedDrive, nil
}

func (d *Driver) createServer() (*api.Server, error) {
	serverCreateRequest := &api.ServerCreateRequest{
		CPU:    2000,
		Memory: 512 * 1024 * 1024,
		Name:   d.MachineName,
		NICS: []api.NIC{
			{IPv4Configuration: api.IPConfiguration{Configuration: "dhcp"}, Model: "virtio"},
		},
		PublicKeys:  []string{d.SSHKeyUUID},
		VNCPassword: "cloudsigma",
	}

	log.Debugf("Creating CloudSigma virtual server...")

	client := d.getClient()
	server, _, err := client.Servers.Create(serverCreateRequest)
	if err != nil {
		return server, err
	}

	attachDriveRequest := &api.AttachDriveRequest{
		CPU: server.CPU,
		Drives: []api.ServerDrive{
			{BootOrder: 1, DevChannel: "0:0", Device: "virtio", DriveUUID: d.DriveUUID},
		},
		Memory:      server.Memory,
		Name:        server.Name,
		VNCPassword: server.VNCPassword,
	}

	log.Debugf("Attaching existing drive to virtual server...")

	serverWithAttachedDrive, _, err := client.Servers.AttachDrive(server, attachDriveRequest)
	if err != nil {
		return serverWithAttachedDrive, err
	}

	return serverWithAttachedDrive, nil
}
