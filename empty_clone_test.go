package cloudsigma

import "testing"

// Malformed synthetic acknowledgments must stop before polling or server creation.
func TestEmptyCloneAcknowledgment(t *testing.T) {
	for _, caller := range []string{"helper", "create"} {
		for name, body := range map[string]string{"empty": `{"objects":[]}`, "absent": `{}`, "null": `{"objects":null}`, "root-null": `null`, "wrong-type": `{"objects":{}}`, "malformed": `{`} {
			t.Run(caller+"/"+name, func(t *testing.T) {
				steps := []step{}
				if caller == "create" {
					steps = append(steps, step{method: "POST", path: "keypairs/", status: 201, payload: `{"objects":[{"name":"fixture","public_key":"ssh-rsa SYNTHETIC-NOT-VALID fixture"}]}`, response: `{"objects":[{"uuid":"created-key"}]}`})
				}
				steps = append(steps, step{method: "POST", path: "libdrives/synthetic-library/action/", query: "do=clone", status: 202, payload: `{"name":"fixture","size":21474836480,"storage_type":"dssd","uuid":""}`, response: body})
				d := fixture(t, steps...)
				d.DriveUUID, d.DriveSize = "synthetic-library", 20
				// Only the test catches panic, to retain all baseline regression outcomes.
				defer func() {
					if got := recover(); got != nil {
						t.Errorf("clone acknowledgment panicked: %v", got)
					}
				}()
				if caller == "create" {
					d.ServerUUID, d.ClonedDriveUUID, d.SSHKeyUUID = "", "", ""
					fixtureKey(t, d, true)
					if err := d.Create(); err == nil {
						t.Fatal("missing acknowledgment error")
					}
					if d.ServerUUID != "" || d.ClonedDriveUUID != "" || d.SSHKeyUUID != "created-key" {
						t.Fatal("unexpected partial ownership")
					}
				} else {
					got, err := d.cloneLibraryDrive()
					if err == nil || got != nil {
						t.Fatalf("got %+v, %v; want nil, error", got, err)
					}
				}
			})
		}
	}
}
