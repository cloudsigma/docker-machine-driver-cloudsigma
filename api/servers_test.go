package api

import (
	"net/http"
	"testing"
)

func TestServers_Create_emptyPayload(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("servers/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
	})

	_, _, err := client.Servers.Create(nil)

	if err != ErrEmptyPayloadNotAllowed {
		t.Errorf(format("Server.Create should return error on empty payload", ErrEmptyPayloadNotAllowed, err))
	}
}
