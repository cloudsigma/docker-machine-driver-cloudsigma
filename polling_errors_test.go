package cloudsigma

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	sdk "github.com/cloudsigma/cloudsigma-sdk-go/cloudsigma"
)

// Exercise public lifecycle methods through the real SDK. A failed observation
// is not successful completion, including after an action was acknowledged.
func TestLifecycleObservationErrors(t *testing.T) {
	for _, phase := range []string{"start-initial", "start-poll", "stop-initial", "stop-poll", "restart-stop-initial"} {
		for _, kind := range []string{"http", "decode", "deadline"} {
			t.Run(phase+"/"+kind, func(t *testing.T) {
				failed := step{method: "GET", path: "servers/synthetic-id/", status: 200}
				switch kind {
				case "http":
					failed.status = 403
					failed.response = `[{"error_message":"observation refused","error_type":"forbidden","error_point":"server"}]`
				case "decode":
					failed.response = `{`
				case "deadline":
					failed.failure = context.DeadlineExceeded
				}
				steps := []step{failed}
				if phase == "start-poll" {
					steps = []step{get("stopped"), {method: "POST", path: "servers/synthetic-id/action/", query: "do=start", payload: `{}`, status: 202, response: `{}`}, failed}
				}
				if phase == "stop-poll" {
					steps = []step{get("running"), {method: "POST", path: "servers/synthetic-id/action/", query: "do=shutdown", payload: `{}`, status: 202, response: `{}`}, failed}
				}
				d := fixture(t, steps...)
				d.IPAddress = "192.0.2.123"
				var err error
				switch phase {
				case "start-initial", "start-poll":
					err = d.Start()
				case "stop-initial", "stop-poll":
					err = d.Stop()
				case "restart-stop-initial":
					err = d.Restart()
				}
				if err == nil {
					t.Fatal("observation error swallowed")
				}
				switch kind {
				case "http":
					var api *sdk.ErrorResponse
					if !errors.As(err, &api) || api.Response.StatusCode != 403 || len(api.Errors) != 1 || api.Errors[0].Message != "observation refused" || api.Errors[0].Type != "forbidden" || api.Errors[0].Point != "server" {
						t.Fatalf("lost SDK error: %T %v", err, err)
					}
				case "decode":
					var syntax *json.SyntaxError
					// An incomplete JSON object is returned as unexpected EOF by encoding/json.
					if !errors.As(err, &syntax) && err.Error() != "unexpected EOF" {
						t.Fatalf("lost decode error: %T %v", err, err)
					}
				case "deadline":
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Fatalf("lost deadline: %v", err)
					}
				}
				wantIP := "192.0.2.123"
				if phase == "start-poll" {
					wantIP = ""
				}
				if d.IPAddress != wantIP || d.ServerUUID != "synthetic-id" || d.ClonedDriveUUID != "synthetic-drive" || d.SSHKeyUUID != "synthetic-key" {
					t.Fatalf("unexpected lifecycle state: %+v", d)
				}
			})
		}
	}
}
