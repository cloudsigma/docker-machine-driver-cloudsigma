package cloudsigma

import "testing"

// Preserve the SDK's private numeric-license decoder at both consumer entries.
// The Create branch stops at an intentional synthetic server rejection, not SSH.
func TestCloneNumericLicenseCompatibility(t *testing.T) {
	for _, caller := range []string{"helper", "create"} {
		t.Run(caller, func(t *testing.T) {
			steps := []step{}
			if caller == "create" {
				steps = append(steps, step{method: "POST", path: "keypairs/", status: 201, payload: `{"objects":[{"name":"fixture","public_key":"ssh-rsa SYNTHETIC-NOT-VALID fixture"}]}`, response: `{"objects":[{"uuid":"created-key"}]}`})
			}
			steps = append(steps,
				step{method: "POST", path: "libdrives/synthetic-library/action/", query: "do=clone", status: 202, payload: `{"name":"fixture","size":21474836480,"storage_type":"dssd","uuid":""}`, response: `{"objects":[{"uuid":"created-drive","licenses":[{"license":{"free":2}}]}]}`},
				step{method: "GET", path: "drives/created-drive/", status: 200, response: `{"uuid":"created-drive","status":"unmounted"}`},
			)
			if caller == "create" {
				steps = append(steps, step{method: "POST", path: "servers/", status: 500, payload: `{"objects":[{"cpu":2000,"mem":1073741824,"name":"fixture","nics":[{"ip_v4_conf":{"conf":"dhcp"},"model":"virtio"}],"pubkeys":[{"uuid":"created-key"}],"vnc_password":"cloudsigma"}]}`, response: `[{"error_message":"synthetic server rejection"}]`})
			}
			d := fixture(t, steps...)
			d.DriveUUID, d.DriveSize = "synthetic-library", 20
			if caller == "create" {
				d.ServerUUID, d.ClonedDriveUUID, d.SSHKeyUUID = "", "", ""
				fixtureKey(t, d, true)
				if err := d.Create(); err == nil {
					t.Fatal("expected synthetic server rejection")
				}
				if d.ClonedDriveUUID != "created-drive" || d.SSHKeyUUID != "created-key" || d.ServerUUID != "" {
					t.Fatal("numeric-license clone did not reach server creation with retained identity")
				}
				return
			}
			got, err := d.cloneLibraryDrive()
			if err != nil {
				t.Fatalf("existing numeric-license response rejected: %v", err)
			}
			if got == nil || len(got.Licenses) != 1 || got.Licenses[0].License.Free != "2" {
				t.Fatalf("lost SDK license decoding: %+v", got)
			}
		})
	}
}
