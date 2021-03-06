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

// Driver represents a CloudSigma driver type.
type Driver struct {
	*drivers.BaseDriver
	APILocation     string
	ClonedDriveUUID string
	CPU             int
	CPUType         string
	DriveName       string
	DriveSize       int
	DriveUUID       string
	Memory          int
	Password        string
	ServerUUID      string
	SSHKeyUUID      string
	StaticIP        string
	Username        string
}

// NewDriver creates a CloudSigma driver.
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// Create makes a new host using the driver's config.
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

// DriverName returns the CloudSigma driver's name.
func (d *Driver) DriverName() string {
	return "cloudsigma"
}

// GetCreateFlags returns the flags that can be set, their descriptions and defaults.
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

// GetSSHHostname returns hostname for use with ssh.
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetSSHKeyPath returns key path for use with ssh.
func (d *Driver) GetSSHKeyPath() string {
	if d.SSHKeyPath == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

// GetURL returns a Docker compatible host URL for connecting to this host.
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

// GetState returns the state that the host is in (running, stopped, etc).
func (d *Driver) GetState() (state.State, error) {
	server, _, err := d.getClient().Servers.Get(context.Background(), d.ServerUUID)
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

// Kill stops a host forcefully.
func (d *Driver) Kill() error {
	_, _, err := d.getClient().Servers.Stop(context.Background(), d.ServerUUID)
	return err
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation.
func (d *Driver) PreCreateCheck() error {
	if d.StaticIP != "" {
		if parsedIP := net.ParseIP(d.StaticIP); parsedIP == nil {
			return fmt.Errorf("%s is not a valid textual representation of an IP address", d.StaticIP)
		}
		_, resp, err := d.getClient().IPs.Get(context.Background(), d.StaticIP)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("IP address %s not found, check your active subscriptions", d.StaticIP)
			}
			return err
		}
	}

	if d.DriveUUID != "" {
		_, resp, err := d.getClient().LibraryDrives.Get(context.Background(), d.DriveUUID)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("drive with UUID %s not found", d.DriveUUID)
			}
			return err
		}
	}
	return nil
}

// Remove deletes a host.
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

	log.Info("Deleting CloudSigma server...")
	if resp, err := client.Servers.Delete(context.Background(), d.ServerUUID); err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Info("CloudSigma server doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}

	log.Info("Deleting CloudSigma drive...")
	if resp, err := client.Drives.Delete(context.Background(), d.ClonedDriveUUID); err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Info("CloudSigma drive doesn't exist, assuming it s already deleted")
		} else {
			return err
		}
	}

	return nil
}

// Restart calls Stop() and Start().
func (d *Driver) Restart() error {
	err := d.stopServer()
	if err != nil {
		return err
	}
	return d.startServer()
}

// SetConfigFromFlags configures the driver with the object returned by RegisterCreateFlags.
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APILocation = flags.String("cloudsigma-api-location")
	d.CPU = flags.Int("cloudsigma-cpu")
	d.CPUType = flags.String("cloudsigma-cpu-type")
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

// Start sends 'start' action to start a host.
func (d *Driver) Start() error {
	return d.startServer()
}

// Stop sends 'stop' action to stop a host gracefully.
func (d *Driver) Stop() error {
	return d.stopServer()
}

func (d *Driver) getClient() *cloudsigma.Client {
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

	keypairs, _, err := d.getClient().Keypairs.Create(context.Background(), keypairCreateRequest)
	if err != nil {
		return nil, err
	}

	keypair := &keypairs[0]
	log.Debugf("Created SSH key UUID: %v", keypair.UUID)

	return keypair, nil
}

func (d *Driver) deleteSSHKey() error {
	sshKeyUUID := d.SSHKeyUUID

	resp, err := d.getClient().Keypairs.Delete(context.Background(), sshKeyUUID)
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
	client := d.getClient()

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

func (d *Driver) createServer() (*cloudsigma.Server, error) {
	serverCreateRequest := &cloudsigma.ServerCreateRequest{
		Servers: []cloudsigma.Server{
			{
				CPU:    d.CPU,
				Memory: d.Memory * 1024 * 1024,
				Name:   d.MachineName,
				NICs: []cloudsigma.ServerNIC{
					{IP4Configuration: &cloudsigma.ServerIPConfiguration{Type: "dhcp"}, Model: "virtio"},
				},
				PublicKeys:  []cloudsigma.Keypair{{UUID: d.SSHKeyUUID}},
				VNCPassword: "cloudsigma",
			},
		},
	}

	if d.StaticIP != "" {
		log.Debugf("Static ip address is defined %s, use it for NIC configuration.", d.StaticIP)

		serverCreateRequest.Servers[0].NICs[0].IP4Configuration = &cloudsigma.ServerIPConfiguration{
			Type:      "static",
			IPAddress: &cloudsigma.IP{UUID: d.StaticIP},
		}
	}

	client := d.getClient()
	servers, _, err := client.Servers.Create(context.Background(), serverCreateRequest)
	if err != nil {
		return &servers[0], err
	}

	server := &servers[0]
	attachDriveRequest := &cloudsigma.ServerAttachDriveRequest{
		CPU: server.CPU,
		Drives: []cloudsigma.ServerDrive{
			{BootOrder: 1, DevChannel: "0:0", Device: "virtio", Drive: &cloudsigma.Drive{UUID: d.ClonedDriveUUID}},
		},
		Memory:      server.Memory,
		Name:        server.Name,
		VNCPassword: server.VNCPassword,
	}

	log.Debug("Attaching existing drive to virtual server...")
	server, _, err = client.Servers.AttachDrive(context.Background(), server.UUID, attachDriveRequest)
	if err != nil {
		return server, err
	}

	return server, nil
}

func (d *Driver) startServer() error {
	client := d.getClient()

	log.Debug("Checking server state...")
	server, _, err := client.Servers.Get(context.Background(), d.ServerUUID)
	if err != nil {
		return nil
	}
	if server.Status == "running" {
		log.Debug("Server is already running")
		return nil
	}

	log.Debug("Starting CloudSigma virtual server...")
	_, _, err = client.Servers.Start(context.Background(), d.ServerUUID)
	if err != nil {
		return err
	}

	d.IPAddress = ""
	log.Debug("Waiting for IP address to be assigned to the server...")
	for {
		time.Sleep(1 * time.Second)

		server, _, err = client.Servers.Get(context.Background(), d.ServerUUID)
		if err != nil {
			return nil
		}
		if server.Runtime == nil {
			continue
		}

		for _, nic := range server.Runtime.RuntimeNICs {
			if nic.InterfaceType == "public" {
				d.IPAddress = nic.IPv4.UUID
			}
		}
		if d.IPAddress != "" && server.Status == "running" {
			break
		}
	}

	return nil
}

func (d *Driver) stopServer() error {
	client := d.getClient()

	log.Debug("Checking server state...")
	server, _, err := client.Servers.Get(context.Background(), d.ServerUUID)
	if err != nil {
		return nil
	}
	if server.Status == "stopped" {
		log.Debug("Server is already stopped")
		return nil
	}

	log.Debug("Stopping CloudSigma server...")
	_, _, err = client.Servers.Shutdown(context.Background(), d.ServerUUID)
	if err != nil {
		return err
	}

	log.Debug("Waiting until server is stopped...")
	for {
		server, _, err := client.Servers.Get(context.Background(), d.ServerUUID)
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
