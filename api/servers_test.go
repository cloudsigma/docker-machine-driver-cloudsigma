package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServers_Create(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	input := &ServerCreateRequest{
		CPU:    100,
		Memory: 200,
	}
	mux.HandleFunc("/servers/", func(w http.ResponseWriter, r *http.Request) {
		v := new(ServerCreateRequest)
		json.NewDecoder(r.Body).Decode(v)
		assert.Equal(t, input, v)

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, authorizationHeader, r.Header.Get("Authorization"))
		fmt.Fprint(w, `{"objects":[{"cpu":100,"mem":300}]}`)
	})
	expected := &Server{
		CPU:    100,
		Memory: 300,
	}

	server, _, err := client.Servers.Create(input)

	assert.NoError(t, err)
	assert.Equal(t, expected, server)
}

func TestServers_Create_emptyPayload(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Servers.Create(nil)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPayloadNotAllowed.Error(), err.Error())
}

func TestServers_Delete(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/servers/some-uuid/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
	})

	_, err := client.Servers.Delete("some-uuid")

	assert.NoError(t, err)
}

func TestServers_Delete_emptyUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, err := client.Servers.Delete("")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument.Error(), err.Error())
}

func TestServers_Delete_invalidUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, err := client.Servers.Delete("%")

	assert.Error(t, err)
}

func TestServers_Get(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/servers/long-uuid", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, authorizationHeader, r.Header.Get("Authorization"))
		fmt.Fprint(w, `{"cpu":100,"mem":200,"name":"test server","resource_uri":"1234-5678","uuid":"long-uuid"}`)
	})
	expected := &Server{
		CPU:         100,
		Memory:      200,
		Name:        "test server",
		ResourceURI: "1234-5678",
		UUID:        "long-uuid",
	}

	server, _, err := client.Servers.Get("long-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expected, server)
}

func TestServers_Get_emptyUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Servers.Get("")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument.Error(), err.Error())
}

func TestServers_Get_invalidUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Servers.Get("%")

	assert.Error(t, err)
}
