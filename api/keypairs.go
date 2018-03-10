package api

import (
	"errors"
	"fmt"
	"net/http"
)

const keypairsBasePath = "keypairs"

// KeypairsService handles communication with the keypairs (SSH keys) related methods of the Cloudsigma API.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/keypairs.html
type KeypairsService service

// Keypair represents a CloudSigma keypair (ssh keys).
type Keypair struct {
	Fingerprint   string `json:"fingerprint,omitempty"`
	HasPrivateKey bool   `json:"has_private_key,omitempty"`
	Name          string `json:"name"`
	PrivateKey    string `json:"private_key,omitempty"`
	PublicKey     string `json:"public_key"`
	ResourceURI   string `json:"resource_key,omitempty"`
	UUID          string `json:"uuid,omitempty"`
}

type KeypairCreateRequest struct {
	Keypairs []Keypair `json:"objects"`
}

// Get provides information for keypair identified by uuid.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/keypairs.html#listing-getting-updating-deleting
func (s *KeypairsService) Get(uuid string) (*Keypair, *http.Response, error) {
	if uuid == "" {
		return nil, nil, ErrEmptyArgument
	}

	path := fmt.Sprintf("%v/%v", keypairsBasePath, uuid)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	keypair := new(Keypair)
	resp, err := s.client.Do(req, keypair)
	if err != nil {
		return nil, resp, err
	}

	return keypair, resp, err
}

// Create makes a keypair.
//
// CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/keypairs.html#listing-getting-updating-deleting
func (s *KeypairsService) Create(keypairCreateRequest *KeypairCreateRequest) (*Keypair, *http.Response, error) {
	if keypairCreateRequest == nil {
		return nil, nil, ErrEmptyPayloadNotAllowed
	}

	path := fmt.Sprintf("%v/", keypairsBasePath)

	req, err := s.client.NewRequest(http.MethodPost, path, keypairCreateRequest)
	if err != nil {
		return nil, nil, err
	}

	root := new(KeypairCreateRequest)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}

	if len(root.Keypairs) > 1 {
		return nil, resp, errors.New("root.Keypairs count cannot be more than 1")
	}

	return &root.Keypairs[0], resp, err
}

// Delete removes the keypair identified by uuid.
//
//CloudSigma API docs: https://cloudsigma-docs.readthedocs.io/en/latest/keypairs.html#listing-getting-updating-deleting
func (s *KeypairsService) Delete(uuid string) (*http.Response, error) {
	if uuid == "" {
		return nil, ErrEmptyArgument
	}

	path := fmt.Sprintf("%v/%v/", keypairsBasePath, uuid)

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
