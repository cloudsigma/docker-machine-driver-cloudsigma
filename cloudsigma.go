package cloudsigma

import (
	"fmt"
	"io/ioutil"
	"net"
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
	defaultDriveUUID = "6fe24a6b-b5c5-40ba-8860-771044d2500d"
	defaultSSHPort   = 22
	defaultSSHUser   = "cloudsigma"
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
	log.Info("Creating SSH key...")
	key, err := d.createSSHKey()
	if err != nil {
		return err
	}
	d.SSHKeyUUID = key.UUID

	log.Info("Cloning library drive...")
	drive, err := d.cloneDrive(defaultDriveUUID)
	if err != nil {
		return err
	}
	d.DriveUUID = drive.UUID

	log.Info("Creating CloudSigma server...")
	server, err := d.createServer()
	if err != nil {
		return err
	}
	d.ServerUUID = server.UUID

	log.Info("Starting CloudSigma server...")
	err = d.startServer()
	if err != nil {
		return err
	}

	log.Debugf("Created server UUID %v, drive UUID %v, IP address %v", d.ServerUUID, d.DriveUUID, d.IPAddress)

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
		mcnflag.IntFlag{
			EnvVar: "CLOUDSIGMA_SSH_PORT",
			Name:   "cloudsigma-ssh-port",
			Usage:  "SSH port",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_SSH_USER",
			Name:   "cloudsigma-ssh-user",
			Usage:  "SSH username",
			Value:  defaultSSHUser,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	if d.SSHKeyPath == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", nil
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
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

	log.Info("Stopping CloudSigma server...")
	err := d.stopServer()
	if err != nil {
		return err
	}

	log.Info("Deleting SSH key...")
	if resp, err := client.Keypairs.Delete(d.SSHKeyUUID); err != nil {
		if resp.StatusCode == http.StatusNotFound {
			log.Info("SSH key doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}

	log.Infof("Deleting CloudSigma server...")
	if resp, err := client.Servers.Delete(d.ServerUUID); err != nil {
		if resp.StatusCode == http.StatusNotFound {
			log.Info("CloudSigma server doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}

	return nil
}

func (d *Driver) Restart() error {
	err := d.stopServer()
	if err != nil {
		return err
	}
	return d.startServer()
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Password = flags.String("cloudsigma-password")
	d.SSHPort = flags.Int("cloudsigma-ssh-port")
	d.SSHUser = flags.String("cloudsigma-ssh-user")
	d.Username = flags.String("cloudsigma-username")

	if d.Username == "" {
		return fmt.Errorf("cloudsigma driver requires the --cloudsigma-username option")
	}

	return nil
}

func (d *Driver) Start() error {
	return d.startServer()
}

func (d *Driver) Stop() error {
	return d.stopServer()
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
		Memory: 1024 * 1024 * 1024,
		Name:   d.MachineName,
		NICS: []api.NIC{
			{IPv4Configuration: api.IPConfiguration{Configuration: "dhcp"}, Model: "virtio"},
		},
		PublicKeys:  []string{d.SSHKeyUUID},
		VNCPassword: "cloudsigma",
	}

	log.Debug("Creating CloudSigma virtual server...")

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

	log.Debug("Attaching existing drive to virtual server...")
	server, _, err = client.Servers.AttachDrive(server, attachDriveRequest)
	if err != nil {
		return server, err
	}

	return server, nil
}

func (d *Driver) startServer() error {
	client := d.getClient()

	log.Debug("Checking server state...")
	server, _, err := client.Servers.Get(d.ServerUUID)
	if err != nil {
		return nil
	}
	if server.Status == "running" {
		log.Debug("Server is already running")
		return nil
	}

	log.Debug("Starting CloudSigma virtual server...")
	_, _, err = client.Servers.Start(d.ServerUUID)
	if err != nil {
		return err
	}

	d.IPAddress = ""
	log.Debug("Waiting for IP address to be assigned to the server...")
	for {
		//TODO: add max-retry-count logic
		server, _, err = client.Servers.Get(d.ServerUUID)
		if err != nil {
			return nil
		}
		for _, nic := range server.Runtime.RuntimeNICS {
			if nic.InterfaceType == "public" {
				d.IPAddress = nic.IPv4.UUID
			}
		}
		if d.IPAddress != "" && server.Status == "running" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *Driver) stopServer() error {
	client := d.getClient()

	log.Debug("Checking server state...")
	server, _, err := client.Servers.Get(d.ServerUUID)
	if err != nil {
		return nil
	}
	if server.Status == "stopped" {
		log.Debug("Server is already stopped")
		return nil
	}

	log.Debug("Stopping CloudSigma virtual server...")
	_, _, err = client.Servers.Shutdown(d.ServerUUID)
	if err != nil {
		return err
	}

	log.Debug("Waiting until server is stopped...")
	for {
		server, _, err := client.Servers.Get(d.ServerUUID)
		if err != nil {
			return err
		}
		if server.Status == "running" {
			return fmt.Errorf("could not stop server %v", d.ServerUUID)
		}
		if server.Status == "stopped" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
