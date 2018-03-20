package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServers_Create_emptyPayload(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("servers/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
	})

	_, _, err := client.Servers.Create(nil)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPayloadNotAllowed.Error(), err.Error())
}
