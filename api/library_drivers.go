package api

import (
	"fmt"
	"net/http"
)

const libdriveBasePath = "libdrives"

// LibraryDrivesService handles communication with the library drives related
// methods of the CloudSigma API.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/2.14/libdrives.html
type LibraryDrivesService service

// LibraryDrive. TODO: enhance structure with mandatory fields
type LibraryDrive struct {
	Arch        string `json:"arch"`
	Description string `json:"description,omitempty"`
	Favourite   bool   `json:"favourite"`
	ImageType   string `json:"image_type"`
	Media       string `json:"media"`
	Name        string `json:"name"`
	OS          string `json:"os"`
	Paid        bool   `json:"paid"`
	ResourceURI string `json:"resource_uri"`
	Size        int    `json:"size"`
	Status      string `json:"status"`
	StorageType string `json:"storage_type"`
	UUID        string `json:"uuid"`
}

// Get detailed information for library drive identified by uuid.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/2.14/libdrives.html#list-single-drive
func (s *LibraryDrivesService) Get(uuid string) (*LibraryDrive, *http.Response, error) {
	path := fmt.Sprintf("%v/%v", libdriveBasePath, uuid)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	libdrive := new(LibraryDrive)
	resp, err := s.client.Do(req, libdrive)
	if err != nil {
		return nil, resp, err
	}

	return libdrive, resp, nil
}
