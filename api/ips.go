package api

import (
	"fmt"
	"net/http"
)

const ipBasePath = "ips"

// IPsService handles communication with the IPs related methods of the CloudSigma API.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/latest/networking.html#ips
type IPsService service

// IP represents a CloudSigma ip address.
type IP struct {
	Gateway     string   `json:"gateway,omitempty"`
	Nameservers []string `json:"nameservers,omitempty"`
	Netmask     int      `json:"netmask,omitempty"`
	UUID        string   `json:"uuid"`
}

// Get provides detailed information for IP address identified by uuid.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/latest/networking.html#get-single-ip
func (s *IPsService) Get(uuid string) (*IP, *http.Response, error) {
	if uuid == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf("%v/%v", ipBasePath, uuid)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	ip := new(IP)
	resp, err := s.client.Do(req, ip)
	if err != nil {
		return nil, resp, err
	}

	return ip, resp, nil
}
