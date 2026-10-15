package cloudsigma

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Existing private-file sentinel exercises GenerateSSHKey's documented reuse
// path. No private key is generated, parsed or used for SSH; both files are fixtures.
func fixtureKey(t *testing.T, d *Driver, public bool) {
	t.Helper()
	d.SSHKeyPath = filepath.Join(t.TempDir(), "synthetic-not-a-private-key")
	if err := os.WriteFile(d.SSHKeyPath, []byte("SYNTHETIC NON-KEY FIXTURE\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if public {
		if err := os.WriteFile(d.SSHKeyPath+".pub", []byte("ssh-rsa SYNTHETIC-NOT-VALID fixture\r\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

// Missing public file must abort Create before API dispatch or resource mutation.
func TestSyntheticCreateSSHReadFailure(t *testing.T) {
	d := fixture(t)
	d.ServerUUID, d.ClonedDriveUUID, d.SSHKeyUUID = "", "", ""
	fixtureKey(t, d, false)
	err := d.Create()
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("want missing public key, got %v", err)
	}
	if d.ServerUUID != "" || d.ClonedDriveUUID != "" || d.SSHKeyUUID != "" {
		t.Fatalf("unexpected resource state: %+v", d)
	}
}

// A clone rejection after successful key registration retains the exact key ID
// for explicit Remove; creation does not silently pretend to roll back.
func TestSyntheticCreateCloneFailureCleanup(t *testing.T) {
	d := fixture(t,
		step{method: "POST", path: "keypairs/", status: 201, payload: `{"objects":[{"name":"fixture","public_key":"ssh-rsa SYNTHETIC-NOT-VALID fixture"}]}`, response: `{"objects":[{"uuid":"synthetic-created-key"}]}`},
		step{method: "POST", path: "libdrives/synthetic-library/action/", query: "do=clone", status: 500, payload: `{"name":"fixture","size":21474836480,"storage_type":"dssd","uuid":""}`, response: `[{"error_message":"synthetic clone rejection"}]`},
	)
	d.ServerUUID, d.ClonedDriveUUID, d.SSHKeyUUID = "", "", ""
	d.DriveUUID, d.DriveSize = "synthetic-library", 20
	fixtureKey(t, d, true)
	if err := d.Create(); err == nil {
		t.Fatal("clone failure swallowed")
	}
	if d.SSHKeyUUID != "synthetic-created-key" || d.ServerUUID != "" || d.ClonedDriveUUID != "" {
		t.Fatalf("unexpected partial state: %+v", d)
	}
}

// Exercise only previously unqualified Remove error/idempotence branches.
func TestSyntheticRemoveBoundaries(t *testing.T) {
	t.Run("all-absent", func(t *testing.T) {
		d := fixture(t, get("stopped"), step{method: "DELETE", path: "keypairs/synthetic-key/", status: 404, response: `[{"error_message":"absent"}]`}, step{method: "DELETE", path: "servers/synthetic-id/", status: 404, response: `[{"error_message":"absent"}]`}, step{method: "DELETE", path: "drives/synthetic-drive/", status: 404, response: `[{"error_message":"absent"}]`})
		if err := d.Remove(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("key-forbidden-stops-deletion", func(t *testing.T) {
		d := fixture(t, get("stopped"), step{method: "DELETE", path: "keypairs/synthetic-key/", status: 403, response: `[{"error_message":"synthetic forbidden"}]`})
		if err := d.Remove(); err == nil || !strings.Contains(err.Error(), "403") {
			t.Fatalf("expected key refusal, got %v", err)
		}
	})
	t.Run("server-failure-preserves-drive", func(t *testing.T) {
		d := fixture(t, get("stopped"), step{method: "DELETE", path: "keypairs/synthetic-key/", status: 204}, step{method: "DELETE", path: "servers/synthetic-id/", status: 500, response: `[{"error_message":"synthetic failure"}]`})
		if err := d.Remove(); err == nil {
			t.Fatal("server failure swallowed")
		}
	})
}
