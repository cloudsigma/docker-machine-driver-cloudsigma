package cloudsigma

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cloudsigma/cloudsigma-sdk-go/cloudsigma"
	"github.com/cloudsigma/docker-machine-driver-cloudsigma/api"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

const (
	defaultCPU       = 2000
	defaultDriveName = "ubuntu"
	defaultDriveSize = 20
	defaultMemory    = 1024
	defaultSSHPort   = 22
	defaultSSHUser   = "cloudsigma"
)

type Driver struct {
	*drivers.BaseDriver
	APILocation         string
	ClonedDriveUUID     string
	CPU                 int
	CPUType             string
	CPUEnclavePageCache string
	DriveName           string
	DriveSize           int
	DriveUUID           string
	Memory              int
	Password            string
	ServerUUID          string
	SSHKeyUUID          string
	StaticIP            string
	Username            string
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

	log.Info("Cloning CloudSigma library drive...")
	drive, err := d.cloneLibraryDrive()
	if err != nil {
		return err
	}
	d.ClonedDriveUUID = drive.UUID

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

	log.Debugf("Created server UUID %v, drive UUID %v, IP address %v", d.ServerUUID, d.ClonedDriveUUID, d.IPAddress)

	return nil
}

func (d *Driver) DriverName() string {
	return "cloudsigma"
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_API_LOCATION",
			Name:   "cloudsigma-api-location",
			Usage:  "CloudSigma API location endpoint code",
		},
		mcnflag.IntFlag{
			EnvVar: "CLOUDSIGMA_CPU",
			Name:   "cloudsigma-cpu",
			Usage:  "CPU clock speed for the host in MHz",
			Value:  defaultCPU,
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_CPU_TYPE",
			Name:   "cloudsigma-cpu-type",
			Usage:  "CPU type",
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_CPU_EPC_SIZE",
			Name:   "cloudsigma-cpu-epc-size",
			Usage:  "Enclave Page Cache (EPC) size",
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_DRIVE_NAME",
			Name:   "cloudsigma-drive-name",
			Usage:  "CloudSigma drive name",
			Value:  defaultDriveName,
		},
		mcnflag.IntFlag{
			EnvVar: "CLOUDSIGMA_DRIVE_SIZE",
			Name:   "cloudsigma-drive-size",
			Usage:  "Drive size for the host in GiB",
			Value:  defaultDriveSize,
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_DRIVE_UUID",
			Name:   "cloudsigma-drive-uuid",
			Usage:  "CloudSigma drive uuid",
		},
		mcnflag.IntFlag{
			EnvVar: "CLOUDSIGMA_MEMORY",
			Name:   "cloudsigma-memory",
			Usage:  "Size of memory for the host in MB",
			Value:  defaultMemory,
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_PASSWORD",
			Name:   "cloudsigma-password",
			Usage:  "CloudSigma password",
		},
		mcnflag.IntFlag{
			EnvVar: "CLOUDSIGMA_SSH_PORT",
			Name:   "cloudsigma-ssh-port",
			Usage:  "SSH port to connect",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_SSH_USER",
			Name:   "cloudsigma-ssh-user",
			Usage:  "SSH username to connect.",
			Value:  defaultSSHUser,
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_STATIC_IP",
			Name:   "cloudsigma-static-ip",
			Usage:  "CloudSigma network adapter’s static IP address",
		},
		mcnflag.StringFlag{
			EnvVar: "CLOUDSIGMA_USERNAME",
			Name:   "cloudsigma-username",
			Usage:  "CloudSigma user email",
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
	if d.StaticIP != "" {
		if parsedIP := net.ParseIP(d.StaticIP); parsedIP == nil {
			return fmt.Errorf("%s is not a valid textual representation of an IP address", d.StaticIP)
		}
		_, _, err := d.getClient().IPs.Get(d.StaticIP)
		if err != nil {
			return err
		}
	}
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
	err = d.deleteSSHKey()
	if err != nil {
		return err
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
	d.APILocation = flags.String("cloudsigma-api-location")
	d.CPU = flags.Int("cloudsigma-cpu")
	d.CPUType = flags.String("cloudsigma-cpu-type")
	d.CPUEnclavePageCache = flags.String("cloudsigma-cpu-epc-size")
	d.DriveName = flags.String("cloudsigma-drive-name")
	d.DriveSize = flags.Int("cloudsigma-drive-size")
	d.DriveUUID = flags.String("cloudsigma-drive-uuid")
	d.Memory = flags.Int("cloudsigma-memory")
	d.Password = flags.String("cloudsigma-password")
	d.SSHPort = flags.Int("cloudsigma-ssh-port")
	d.SSHUser = flags.String("cloudsigma-ssh-user")
	d.StaticIP = flags.String("cloudsigma-static-ip")
	d.Username = flags.String("cloudsigma-username")

	if d.Username == "" {
		return fmt.Errorf("cloudsigma driver requires the --cloudsigma-username option")
	}

	if d.Password == "" {
		return fmt.Errorf("cloudsigma driver requires the --cloudsigma-password option")
	}

	if d.DriveName != "" && d.DriveUUID != "" {
		return fmt.Errorf("--cloudsigma-drive-name and --cloudsigma-drive-uuid are mutually exclusive")
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
	client := api.NewBasicAuthClient(d.Username, d.Password)
	if d.APILocation != "" {
		client.SetLocationForBaseURL(d.APILocation)
	}
	return client
}

func (d *Driver) getSDKClient() *cloudsigma.Client {
	client := cloudsigma.NewBasicAuthClient(d.Username, d.Password)
	if d.APILocation != "" {
		client.SetLocation(d.APILocation)
	}
	client.SetUserAgent("docker-machine-driver-cloudsigma")
	return client
}

func (d *Driver) createSSHKey() (*cloudsigma.Keypair, error) {
	d.SSHKeyPath = d.GetSSHKeyPath()

	if err := ssh.GenerateSSHKey(d.SSHKeyPath); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.SSHKeyPath + ".pub")
	if err != nil {
		return nil, err
	}

	keypairCreateRequest := &cloudsigma.KeypairCreateRequest{
		Keypairs: []cloudsigma.Keypair{
			{Name: d.MachineName, PublicKey: strings.TrimRight(string(publicKey), "\r\n")},
		},
	}

	keypairs, _, err := d.getSDKClient().Keypairs.Create(context.Background(), keypairCreateRequest)
	if err != nil {
		return nil, err
	}

	return &keypairs[0], nil
}

func (d *Driver) deleteSSHKey() error {
	sshKeyUUID := d.SSHKeyUUID

	resp, err := d.getSDKClient().Keypairs.Delete(context.Background(), sshKeyUUID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Info("SSH key doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}

	return nil
}

func (d *Driver) cloneLibraryDrive() (*cloudsigma.LibraryDrive, error) {
	client := d.getSDKClient()

	driveUUID := ""
	if d.DriveName != "" {
		listOptions := &cloudsigma.LibraryDriveListOptions{
			NamesContain: []string{d.DriveName},
			ListOptions:  cloudsigma.ListOptions{Limit: 0},
		}
		libraryDrives, _, err := client.LibraryDrives.List(context.Background(), listOptions)
		if err != nil {
			return nil, err
		}

		libdriveVersion := ""
		libdriveUUID := ""
		for _, libraryDrive := range libraryDrives {
			if libraryDrive.ImageType != "preinst" {
				continue
			}
			// exclude Ubuntu 20.10 because there is no stable repo with docker engine
			// wait until https://github.com/docker/machine/issues/4856 is fixed
			if strings.Contains(libraryDrive.Name, "Ubuntu 20.10") {
				log.Debugf("Skip library drive %s (Ubuntu 20.10), because provisioning scripts are not available in stable channel", libraryDrive.UUID)
				continue
			}

			if libdriveVersion == "" {
				libdriveVersion = libraryDrive.Version
				libdriveUUID = libraryDrive.UUID
			}
			if libdriveVersion < libraryDrive.Version {
				libdriveVersion = libraryDrive.Version
				libdriveUUID = libraryDrive.UUID
			}
		}

		if libdriveUUID == "" {
			return nil, fmt.Errorf("could not find any library drive with name %s", d.DriveName)
		}
		log.Debugf("Found library drive: %s, version: %s, UUID: %s", d.DriveName, libdriveVersion, libdriveUUID)

		driveUUID = libdriveUUID
	} else {
		driveUUID = d.DriveUUID
	}

	cloneRequest := &cloudsigma.LibraryDriveCloneRequest{
		LibraryDrive: &cloudsigma.LibraryDrive{
			Name:        d.MachineName,
			Size:        d.DriveSize * 1024 * 1024 * 1024,
			StorageType: "dssd",
		},
	}
	clonedDrive, _, err := client.LibraryDrives.Clone(context.Background(), driveUUID, cloneRequest)
	if err != nil {
		return clonedDrive, err
	}

	log.Debugf("Waiting until cloning process are done...")

	for {
		drive, _, err := client.Drives.Get(context.Background(), clonedDrive.UUID)
		if err != nil {
			return clonedDrive, nil
		}

		if drive.Status == "unmounted" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Debugf("Created drive UUID %v", clonedDrive.UUID)

	return clonedDrive, nil
}

func (d *Driver) createServer() (*api.Server, error) {
	serverCreateRequest := &api.ServerCreateRequest{
		CPU:    d.CPU,
		Memory: d.Memory * 1024 * 1024,
		Name:   d.MachineName,
		NICS: []api.NIC{
			{IPv4Configuration: api.IPConfiguration{Configuration: "dhcp"}, Model: "virtio"},
		},
		PublicKeys:  []string{d.SSHKeyUUID},
		VNCPassword: "cloudsigma",
	}

	if d.StaticIP != "" {
		log.Debugf("Static ip address is defined %s, use it for NIC configuration.", d.StaticIP)

		serverCreateRequest.NICS[0].IPv4Configuration = api.IPConfiguration{
			Configuration: "static",
			IP:            d.StaticIP,
		}
	}

	if d.CPUEnclavePageCache != "" {
		log.Debugf("CPU enclave page cache is defined %s, use it by server creation.", d.CPUEnclavePageCache)

		serverCreateRequest.CPUEnclavePageCache = d.CPUEnclavePageCache
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
			{BootOrder: 1, DevChannel: "0:0", Device: "virtio", DriveUUID: d.ClonedDriveUUID},
		},
		Memory:      server.Memory,
		Name:        server.Name,
		VNCPassword: server.VNCPassword,
	}

	log.Debug("Attaching existing drive to virtual server...")
	server, _, err = client.Servers.AttachDrive(server.UUID, attachDriveRequest)
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
