package api

import (
	"errors"
	"fmt"
	"net/http"
)

const serverBasePath = "servers"

// ServersService handles communication with the servers related methods of the CloudSigma API.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/latest/servers.html
type ServersService service

// Server represents a CloudSigma server.
type Server struct {
	CPU         int           `json:"cpu"`
	CPUType     string        `json:"cpu_type"`
	Drives      []ServerDrive `json:"drive,omitempty"`
	Memory      int           `json:"mem"`
	Name        string        `json:"name"`
	Owner       Owner         `json:"owner"`
	PublicKeys  []PublicKey   `json:"pubkeys,omitempty"`
	ResourceURI string        `json:"resource_uri"`
	Runtime     Runtime       `json:"runtime,omitempty"`
	Status      string        `json:"status,omitempty"`
	UUID        string        `json:"uuid"`
	VNCPassword string        `json:"vnc_password"`
}

type ServerDrive struct {
	BootOrder  int    `json:"boot_order"`
	DevChannel string `json:"dev_channel"`
	Device     string `json:"device"`
	DriveUUID  string `json:"drive"`
}

type Owner struct {
	ResourceURI string `json:"resource_uri,omitempty"`
	UUID        string `json:"uuid"`
}

type PublicKey struct {
	ResourceURI string `json:"resource_uri,omitempty"`
	UUID        string `json:"uuid"`
}

type Runtime struct {
	RuntimeNICS []RuntimeNIC `json:"nics,omitempty"`
}

type RuntimeNIC struct {
	InterfaceType string      `json:"interface_type,omitempty"`
	IPv4          RuntimeIPv4 `json:"ip_v4,omitempty"`
}

type RuntimeIPv4 struct {
	ResourceURI string `json:"resource_uri,omitempty"`
	UUID        string `json:"uuid,omitempty"`
}

type ServerAction struct {
	Action string `json:"action,omitempty"`
	Result string `json:"result,omitempty"`
	UUID   string `json:"uuid,omitempty"`
}

type AttachDriveRequest struct {
	CPU         int           `json:"cpu"`
	CPUType     string        `json:"cpu_type"`
	Drives      []ServerDrive `json:"drives"`
	Memory      int           `json:"mem"`
	Name        string        `json:"name"`
	VNCPassword string        `json:"vnc_password"`
}

type ServerCreateRequest struct {
	CPU                 int      `json:"cpu"`
	CPUType             string   `json:"cpu_type"`
	CPUEnclavePageCache string   `json:"cpu_epc,omitempty"`
	Memory              int      `json:"mem"`
	Name                string   `json:"name"`
	VNCPassword         string   `json:"vnc_password"`
	NICS                []NIC    `json:"nics,omitempty"`
	PublicKeys          []string `json:"pubkeys,omitempty"`
}

type NIC struct {
	IPv4Configuration IPConfiguration `json:"ip_v4_conf,omitempty"`
	Model             string          `json:"model,omitempty"`
}

type IPConfiguration struct {
	Configuration string `json:"conf,omitempty"`
	IP            string `json:"ip,omitempty"`
}

type serversRoot struct {
	Servers []Server `json:"objects"`
}

// Get provides detailed information for server identified by uuid.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/servers.html#server-runtime-and-server-details
func (s *ServersService) Get(uuid string) (*Server, *http.Response, error) {
	if uuid == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf("%v/%v", serverBasePath, uuid)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	server := new(Server)
	resp, err := s.client.Do(req, server)
	if err != nil {
		return nil, resp, err
	}

	return server, resp, nil
}

// Create makes a new virtual server with given payload.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/servers.html#creating
func (s *ServersService) Create(serverCreateRequest *ServerCreateRequest) (*Server, *http.Response, error) {
	if serverCreateRequest == nil {
		return nil, nil, ErrEmptyPayloadNotAllowed
	}

	path := fmt.Sprintf("%v/", serverBasePath)

	req, err := s.client.NewRequest(http.MethodPost, path, serverCreateRequest)
	if err != nil {
		return nil, nil, err
	}

	root := new(serversRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}

	if len(root.Servers) > 1 {
		return nil, resp, errors.New("root.Servers count cannot be more than 1")
	}

	return &root.Servers[0], resp, err
}

// Delete removes a server together with it's all attached drives (recursive delete).
//
//CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/servers.html#delete-server-together-with-attached-drives-recursive-delete
func (s *ServersService) Delete(uuid string) (*http.Response, error) {
	if uuid == "" {
		return nil, ErrEmptyArgument
	}

	path := fmt.Sprintf("%v/%v/?recurse=all_drives", serverBasePath, uuid)

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ServerDrive edits a server with attaching drives to it.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/servers.html#attach-a-drive
func (s *ServersService) AttachDrive(serverUUID string, attachDriveRequest *AttachDriveRequest) (*Server, *http.Response, error) {
	if serverUUID == "" {
		return nil, nil, ErrEmptyArgument
	}
	if attachDriveRequest == nil {
		return nil, nil, ErrEmptyPayloadNotAllowed
	}

	path := fmt.Sprintf("%v/%v/", serverBasePath, serverUUID)

	req, err := s.client.NewRequest(http.MethodPut, path, attachDriveRequest)
	if err != nil {
		return nil, nil, err
	}

	serverWithDrives := new(Server)
	resp, err := s.client.Do(req, serverWithDrives)
	if err != nil {
		return nil, resp, err
	}

	return serverWithDrives, resp, nil
}

// Start starts a server with specific UUID.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/servers.html#start
func (s *ServersService) Start(uuid string) (*ServerAction, *http.Response, error) {
	return s.doAction(uuid, "start")
}

// Stop stops a server with specific UUID. This action is equivalent to pulling the power cord of a physical server.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/servers.html#stop
func (s *ServersService) Stop(uuid string) (*ServerAction, *http.Response, error) {
	return s.doAction(uuid, "stop")
}

// Stop Sends an ACPI shutdowns to a server with specific UUID for a minute.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/servers.html#acpi-shutdown
func (s *ServersService) Shutdown(uuid string) (*ServerAction, *http.Response, error) {
	return s.doAction(uuid, "shutdown")
}

func (s *ServersService) doAction(uuid, action string) (*ServerAction, *http.Response, error) {
	if uuid == "" || action == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf("%v/%v/action/?do=%v", serverBasePath, uuid, action)

	req, err := s.client.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, nil, err
	}

	serverAction := new(ServerAction)
	resp, err := s.client.Do(req, serverAction)
	if err != nil {
		return nil, resp, err
	}

	return serverAction, resp, nil
}
