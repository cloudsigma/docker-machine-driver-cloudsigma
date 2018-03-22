package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrives_Get(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/drives/long-uuid", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, authorizationHeader, r.Header.Get("Authorization"))
		fmt.Fprint(w, `{"name":"my drive","size":1000,"uuid":"long-uuid"}`)
	})
	expected := &Drive{
		Name: "my drive",
		Size: 1000,
		UUID: "long-uuid",
	}

	drive, _, err := client.Drives.Get("long-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expected, drive)
}

func TestDrives_Get_emptyUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Drives.Get("")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument.Error(), err.Error())
}

func TestDrives_Get_invalidUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Drives.Get("%")

	assert.Error(t, err)
}
