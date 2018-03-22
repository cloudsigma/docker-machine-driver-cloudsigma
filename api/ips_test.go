package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPs_Get(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/ips/long-uuid", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, authorizationHeader, r.Header.Get("Authorization"))
		fmt.Fprint(w, `{"gateway":"185.12.6.1","uuid":"long-uuid"}`)
	})
	expected := &IP{
		Gateway: "185.12.6.1",
		UUID:    "long-uuid",
	}

	ip, _, err := client.IPs.Get("long-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expected, ip)
}

func TestIPs_Get_emptyUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.IPs.Get("")

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArgument.Error(), err.Error())
}

func TestIPs_Get_invalidUUID(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.IPs.Get("%")

	assert.Error(t, err)
}
