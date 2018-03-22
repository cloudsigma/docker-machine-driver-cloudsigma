package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLibraryDrives_Clone(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/libdrives/long-uuid/action/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, authorizationHeader, r.Header.Get("Authorization"))
		fmt.Fprint(w, `{"objects":[{"name":"cloned drive","size":1000,"uuid":"generated-uuid"}]}`)
	})
	expected := &Drive{
		Name: "cloned drive",
		Size: 1000,
		UUID: "generated-uuid",
	}

	drive, _, err := client.LibraryDrives.Clone("long-uuid", &DriveCloneRequest{})

	assert.NoError(t, err)
	assert.Equal(t, expected, drive)
}

func TestLibraryDrives_Clone_emptyUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.LibraryDrives.Clone("", &DriveCloneRequest{})

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument.Error(), err.Error())
}

func TestLibraryDrives_Clone_emptyPayload(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.LibraryDrives.Clone("long-uuid", nil)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPayloadNotAllowed.Error(), err.Error())
}

func TestLibraryDrives_Clone_invalidUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.LibraryDrives.Clone("%", &DriveCloneRequest{})

	assert.Error(t, err)
}

func TestLibraryDrives_Get(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/libdrives/long-uuid", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, authorizationHeader, r.Header.Get("Authorization"))
		fmt.Fprint(w, `{"arch":"64","name":"long-uuid"}`)
	})
	expected := &LibraryDrive{
		Arch: "64",
		Name: "long-uuid",
	}

	libdrive, _, err := client.LibraryDrives.Get("long-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expected, libdrive)
}

func TestLibraryDrives_Get_emptyUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.LibraryDrives.Get("")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument.Error(), err.Error())
}

func TestLibraryDrives_Get_invalidUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.LibraryDrives.Get("%")

	assert.Error(t, err)
}
