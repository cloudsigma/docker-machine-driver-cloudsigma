package api

import (
	"errors"
	"fmt"
	"net/http"
)

const serverBasePath = "servers"

// ServersService handles communication with the servers related methods of the CloudSigma API.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/2.14/servers.html
type ServersService service

// Server represents a CloudSigma server.
type Server struct {
	CPU         int    `json:"cpu"`
	CPUType     string `json:"cpu_type,omitempty"`
	Memory      int    `json:"mem"`
	Name        string `json:"name"`
	Owner       Owner  `json:"owner"`
	ResourceURI string `json:"resource_uri"`
	UUID        string `json:"uuid"`
	VNCPassword string `json:"vnc_password"`
}

type Owner struct {
	ResourceURI string `json:"resource_uri,omitempty"`
	UUID        string `json:"uuid"`
}

type ServerCreateRequest struct {
	CPU         int    `json:"cpu"`
	Memory      int    `json:"mem"`
	Name        string `json:"name"`
	VNCPassword string `json:"vnc_password"`
	NICS        []NIC  `json:"nics,omitempty"`
}

type NIC struct {
	IPv4Configuration IPConfiguration `json:"ip_v4_conf,omitempty"`
	Model             string          `json:"model,omitempty"`
}

type IPConfiguration struct {
	Configuration string `json:"conf,omitempty"`
}

type serversRoot struct {
	Servers []Server `json:"objects"`
}

// Get provides detailed information for server identified by uuid.
func (s *ServersService) Get(uuid string) (*Server, *http.Response, error) {
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
		return nil, resp, errors.New("root.Servers count cannot be more then 1")
	}

	return &root.Servers[0], resp, err
}
