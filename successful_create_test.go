package cloudsigma

import (
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/docker/machine/libmachine/state"
)

// This is the actual complete Driver.Create, not libmachine provisioning. The
// existing-key filesystem seam avoids generation; Create has no SSH execution.
// Do not parallelize: the inherited fixture replaces http.DefaultTransport.
func TestSuccessfulCreateLifecycle(t *testing.T) {
	for _, static := range []bool{false, true} {
		name := "dhcp"
		nic := `{"conf":"dhcp"}`
		if static {
			name = "static"
			nic = `{"conf":"static","ip":{"uuid":"192.0.2.20"}}`
		}
		t.Run(name, func(t *testing.T) {
			steps := []step{
				{method: "POST", path: "keypairs/", status: 201, payload: `{"objects":[{"name":"fixture","public_key":"ssh-rsa SYNTHETIC-NOT-VALID fixture"}]}`, response: `{"objects":[{"uuid":"created-key"}]}`},
				{method: "POST", path: "libdrives/synthetic-library/action/", query: "do=clone", status: 202, payload: `{"name":"fixture","size":21474836480,"storage_type":"dssd","uuid":""}`, response: `{"objects":[{"uuid":"created-drive"}]}`},
				{method: "GET", path: "drives/created-drive/", status: 200, response: `{"uuid":"created-drive","status":"cloning"}`},
				{method: "GET", path: "drives/created-drive/", status: 200, response: `{"uuid":"created-drive","status":"unmounted"}`},
				{method: "POST", path: "servers/", status: 201, payload: `{"objects":[{"cpu":2000,"mem":1073741824,"name":"fixture","nics":[{"ip_v4_conf":` + nic + `,"model":"virtio"}],"pubkeys":[{"uuid":"created-key"}],"vnc_password":"cloudsigma"}]}`, response: `{"objects":[{"uuid":"created-server","cpu":2000,"mem":1073741824,"name":"fixture","vnc_password":"cloudsigma"}]}`},
				{method: "PUT", path: "servers/created-server/", status: 200, payload: `{"cpu":2000,"mem":1073741824,"name":"fixture","vnc_password":"cloudsigma","drives":[{"boot_order":1,"dev_channel":"0:0","device":"virtio","drive":{"uuid":"created-drive"}}]}`, response: `{"uuid":"created-server"}`},
				{method: "GET", path: "servers/created-server/", status: 200, response: `{"uuid":"created-server","status":"stopped"}`},
				{method: "POST", path: "servers/created-server/action/", query: "do=start", status: 202, payload: `{}`, response: `{}`},
				{method: "GET", path: "servers/created-server/", status: 200, response: `{"uuid":"created-server","status":"starting"}`},
				{method: "GET", path: "servers/created-server/", status: 200, response: `{"uuid":"created-server","status":"running","runtime":{"nics":[{"interface_type":"private","ip_v4":{"uuid":"192.0.2.99"}}]}}`},
				{method: "GET", path: "servers/created-server/", status: 200, response: `{"uuid":"created-server","status":"running","runtime":{"nics":[{"interface_type":"public","ip_v4":{"uuid":"192.0.2.20"}}]}}`},
				{method: "GET", path: "servers/created-server/", status: 200, response: `{"uuid":"created-server","status":"running"}`},
			}
			d := fixture(t, steps...)
			d.ServerUUID, d.ClonedDriveUUID, d.SSHKeyUUID = "", "", ""
			d.DriveUUID, d.DriveSize = "synthetic-library", 20
			d.IPAddress = "192.0.2.123" // Start must discard this stale address.
			if static {
				d.StaticIP = "192.0.2.20"
			}
			fixtureKey(t, d, true)
			keyPath := d.SSHKeyPath
			// Observe driver-owned IDs before each dispatch, not only final state.
			transport := http.DefaultTransport
			call := 0
			http.DefaultTransport = transportFunc(func(r *http.Request) (*http.Response, error) {
				want := [3]string{}
				if call >= 1 {
					want[0] = "created-key"
				}
				if call >= 4 {
					want[1] = "created-drive"
				}
				if call >= 6 {
					want[2] = "created-server"
				}
				got := [3]string{d.SSHKeyUUID, d.ClonedDriveUUID, d.ServerUUID}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("dispatch %d IDs %v want %v", call, got, want)
				}
				if call >= 8 && call <= 10 && d.IPAddress != "" {
					t.Fatalf("premature/stale IP at %d: %s", call, d.IPAddress)
				}
				call++
				return transport.RoundTrip(r)
			})
			if err := d.Create(); err != nil {
				t.Fatal(err)
			}
			if call != 11 || d.IPAddress != "192.0.2.20" || d.SSHKeyPath != keyPath {
				t.Fatalf("Create state calls=%d IP=%s key=%s", call, d.IPAddress, d.SSHKeyPath)
			}
			gotState, err := d.GetState()
			if err != nil || gotState != state.Running {
				t.Fatalf("state=%v error=%v", gotState, err)
			}
			raw, err := os.ReadFile(keyPath)
			if err != nil || string(raw) != "SYNTHETIC NON-KEY FIXTURE\n" {
				t.Fatal("private sentinel changed", err)
			}
			raw, err = os.ReadFile(keyPath + ".pub")
			if err != nil || string(raw) != "ssh-rsa SYNTHETIC-NOT-VALID fixture\r\n" {
				t.Fatal("public fixture changed", err)
			}
		})
	}
}
