package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeypairs_Create(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	input := &KeypairCreateRequest{
		Keypairs: []Keypair{
			{Name: "uploaded key", PublicKey: "long-long-public-key"},
		},
	}
	mux.HandleFunc("/keypairs/", func(w http.ResponseWriter, r *http.Request) {
		v := new(KeypairCreateRequest)
		_ = json.NewDecoder(r.Body).Decode(v)
		assert.Equal(t, input, v)

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, authorizationHeader, r.Header.Get("Authorization"))
		_, _ = fmt.Fprint(w, `{"objects":[{"name":"uploaded key","public_key":"long-long-public-key"}]}`)
	})
	expected := &Keypair{
		Name:      "uploaded key",
		PublicKey: "long-long-public-key",
	}

	keypair, _, err := client.Keypairs.Create(input)

	assert.NoError(t, err)
	assert.Equal(t, expected, keypair)
}

func TestKeypairs_Create_emptyPayload(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Keypairs.Create(nil)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPayloadNotAllowed.Error(), err.Error())
}

func TestKeypairs_Delete(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/keypairs/long-uuid/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
	})

	_, err := client.Keypairs.Delete("long-uuid")

	assert.NoError(t, err)
}

func TestKeypairs_Delete_emptyUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, err := client.Keypairs.Delete("")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument.Error(), err.Error())
}

func TestKeypairs_Delete_invalidUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, err := client.Keypairs.Delete("%")

	assert.Error(t, err)
}

func TestKeypairs_Get(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/keypairs/long-uuid", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, authorizationHeader, r.Header.Get("Authorization"))
		_, _ = fmt.Fprint(w, `{"name":"generated ssh keypair","uuid":"long-uuid"}`)
	})
	expected := &Keypair{
		Name: "generated ssh keypair",
		UUID: "long-uuid",
	}

	keypair, _, err := client.Keypairs.Get("long-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expected, keypair)
}

func TestKeypairs_Get_emptyUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Keypairs.Get("")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument.Error(), err.Error())
}

func TestKeypairs_Get_invalidUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Keypairs.Get("%")

	assert.Error(t, err)
}
