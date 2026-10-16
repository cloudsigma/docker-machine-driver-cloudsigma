package cloudsigma

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	sdk "github.com/cloudsigma/cloudsigma-sdk-go/cloudsigma"
)

// These fixtures exercise the real SDK and driver, not cloud completion.
// An acknowledged clone is retained in the helper result on observation failure.
func TestCloneObservationErrors(t *testing.T) {
	for _, caller := range []string{"helper", "create"} {
		for _, phase := range []string{"first-poll", "after-cloning"} {
			for _, kind := range []string{"http", "decode", "deadline"} {
				t.Run(caller+"/"+phase+"/"+kind, func(t *testing.T) {
					failed := step{method: "GET", path: "drives/created-drive/", status: 200}
					switch kind {
					case "http":
						failed.status = 403
						failed.response = `[{"error_message":"clone observation refused","error_type":"forbidden","error_point":"drive"}]`
					case "decode":
						failed.response = `{`
					case "deadline":
						failed.failure = context.DeadlineExceeded
					}
					steps := []step{}
					if caller == "create" {
						steps = append(steps, step{method: "POST", path: "keypairs/", status: 201, payload: `{"objects":[{"name":"fixture","public_key":"ssh-rsa SYNTHETIC-NOT-VALID fixture"}]}`, response: `{"objects":[{"uuid":"created-key"}]}`})
					}
					steps = append(steps, step{method: "POST", path: "libdrives/synthetic-library/action/", query: "do=clone", status: 202, payload: `{"name":"fixture","size":21474836480,"storage_type":"dssd","uuid":""}`, response: `{"objects":[{"uuid":"created-drive","name":"acknowledged-clone"}]}`})
					if phase == "after-cloning" {
						steps = append(steps, step{method: "GET", path: "drives/created-drive/", status: 200, response: `{"uuid":"created-drive","status":"cloning"}`})
					}
					steps = append(steps, failed)
					d := fixture(t, steps...)
					d.DriveUUID, d.DriveSize = "synthetic-library", 20
					var err error
					if caller == "create" {
						d.ServerUUID, d.ClonedDriveUUID, d.SSHKeyUUID = "", "", ""
						fixtureKey(t, d, true)
						err = d.Create()
						// No server dispatch or implicit cleanup is permitted. Current ownership
						// only records the key on this early return; this is not a cleanup claim.
						if d.SSHKeyUUID != "created-key" || d.ClonedDriveUUID != "" || d.ServerUUID != "" {
							t.Fatalf("unexpected partial state: %+v", d)
						}
					} else {
						var got *sdk.LibraryDrive
						got, err = d.cloneLibraryDrive()
						if got == nil || got.UUID != "created-drive" || got.Name != "acknowledged-clone" {
							t.Fatalf("lost acknowledged clone: %+v", got)
						}
						if d.ClonedDriveUUID != "synthetic-drive" || d.ServerUUID != "synthetic-id" || d.SSHKeyUUID != "synthetic-key" {
							t.Fatal("helper mutated ownership")
						}
					}
					if err == nil {
						t.Fatal("clone observation error swallowed")
					}
					switch kind {
					case "http":
						var api *sdk.ErrorResponse
						if !errors.As(err, &api) || api.Response.StatusCode != 403 || len(api.Errors) != 1 || api.Errors[0].Message != "clone observation refused" || api.Errors[0].Type != "forbidden" || api.Errors[0].Point != "drive" {
							t.Fatalf("lost SDK error: %T %v", err, err)
						}
					case "decode":
						var syntax *json.SyntaxError
						if !errors.As(err, &syntax) && err.Error() != "unexpected EOF" {
							t.Fatalf("lost decoder error: %T %v", err, err)
						}
					case "deadline":
						if !errors.Is(err, context.DeadlineExceeded) {
							t.Fatalf("lost deadline identity: %v", err)
						}
					}
				})
			}
		}
	}
}
