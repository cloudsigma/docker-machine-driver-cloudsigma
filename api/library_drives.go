package api

import (
	"errors"
	"fmt"
	"net/http"
)

const libdriveBasePath = "libdrives"

// LibraryDrivesService handles communication with the library drives related
// methods of the CloudSigma API.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/latest/libdrives.html
type LibraryDrivesService service

// LibraryDrive represents a CloudSigma library drive.
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

// Get provides detailed information for library drive identified by uuid.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/latest/libdrives.html#list-single-drive
func (s *LibraryDrivesService) Get(uuid string) (*LibraryDrive, *http.Response, error) {
	if uuid == "" {
		return nil, nil, ErrEmptyArgument
	}

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

// Clone duplicates a drive. Request body is optional and any or all of the key/value pairs from the drive
// definition can be omitted. Size of the cloned drive can only be bigger or the same.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/latest/libdrives.html#cloning-library-drive
func (s *LibraryDrivesService) Clone(uuid string, driveCloneRequest *DriveCloneRequest) (*Drive, *http.Response, error) {
	if uuid == "" {
		return nil, nil, ErrEmptyArgument
	}
	if driveCloneRequest == nil {
		return nil, nil, ErrEmptyPayloadNotAllowed
	}

	path := fmt.Sprintf("%v/%v/action/?do=clone", libdriveBasePath, uuid)

	req, err := s.client.NewRequest(http.MethodPost, path, driveCloneRequest)
	if err != nil {
		return nil, nil, err
	}

	root := new(drivesRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}

	if len(root.Drives) > 1 {
		return nil, resp, errors.New("root.Drives count cannot be more than 1")
	}

	return &root.Drives[0], resp, err
}
