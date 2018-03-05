package api

import (
	"fmt"
	"net/http"
)

const driveBasePath = "drives"

// DrivesService handles communication with the drives related methods of the CloudSigma API.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/2.14/drives.html
type DrivesService service

// Drive represents a CloudSigma drive.
type Drive struct {
	Media       string `json:"media"`
	Name        string `json:"name"`
	ResourceURI string `json:"resource_uri"`
	Size        int    `json:"size"`
	Status      string `json:"status"`
	StorageType string `json:"storage_type"`
	UUID        string `json:"uuid"`
}

type DriveCloneRequest struct {
	Media       string `json:"media,omitempty"`
	Name        string `json:"name,omitempty"`
	Size        int    `json:"size,omitempty"`
	StorageType string `json:"storage_type,omitempty"`
}

type drivesRoot struct {
	Drives []Drive `json:"objects"`
}

// Get detailed information for drive identified by uuid.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/2.14/drives.html#list-single-drive
func (s *DrivesService) Get(uuid string) (*Drive, *http.Response, error) {
	path := fmt.Sprintf("%v/%v", driveBasePath, uuid)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	drive := new(Drive)
	resp, err := s.client.Do(req, drive)
	if err != nil {
		return nil, resp, err
	}

	return drive, resp, nil
}
