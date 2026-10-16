package cloudsigma

import (
	"strings"
	"testing"
)

// Characterize the driver's existing name-selection policy, not semantic-version
// ordering or provider acceptance. An intentional policy repair may supersede
// these legacy expectations. All calls use the real method and SDK transport.
func TestLibraryImageSelection(t *testing.T) {
	for _, tc := range []struct{ name, objects, selected, wantError string }{
		{"eligible-highest", `[{"uuid":"install","image_type":"install","version":"99"},{"uuid":"old","image_type":"preinst","name":"Ubuntu 22.04","version":"22.04"},{"uuid":"excluded","image_type":"preinst","name":"Ubuntu 20.10 Server","version":"99"},{"uuid":"new","image_type":"preinst","name":"Ubuntu 24.04","version":"24.04"}]`, "new", ""},
		{"reverse-order", `[{"uuid":"new","image_type":"preinst","version":"24.04"},{"uuid":"old","image_type":"preinst","version":"22.04"}]`, "new", ""},
		{"equal-version-first", `[{"uuid":"first","image_type":"preinst","version":"24.04"},{"uuid":"second","image_type":"preinst","version":"24.04"}]`, "first", ""},
		{"legacy-lexical-not-semver", `[{"uuid":"ten","image_type":"preinst","version":"10"},{"uuid":"nine","image_type":"preinst","version":"9"}]`, "nine", ""},
		{"empty", `[]`, "", "could not find any library drive with name Ubuntu & LTS"},
		{"ineligible-only", `[{"uuid":"install","image_type":"install","version":"99"},{"uuid":"excluded","image_type":"preinst","name":"Ubuntu 20.10 Server","version":"99"}]`, "", "could not find any library drive with name Ubuntu & LTS"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []step{{method: "GET", path: "libdrives/", query: "limit=0&name__icontains=Ubuntu+%26+LTS", status: 200, response: `{"objects":` + tc.objects + `}`}}
			if tc.selected != "" {
				steps = append(steps, step{method: "POST", path: "libdrives/" + tc.selected + "/action/", query: "do=clone", status: 202, payload: `{"name":"fixture","size":21474836480,"storage_type":"dssd","uuid":""}`, response: `{"objects":[{"uuid":"selected-clone"}]}`}, step{method: "GET", path: "drives/selected-clone/", status: 200, response: `{"uuid":"selected-clone","status":"unmounted"}`})
			}
			d := fixture(t, steps...)
			d.DriveName, d.DriveUUID, d.DriveSize = "Ubuntu & LTS", "must-not-win", 20
			got, err := d.cloneLibraryDrive()
			if tc.wantError != "" {
				if err == nil || err.Error() != tc.wantError || got != nil {
					t.Fatalf("drive=%v error=%v", got, err)
				}
			} else if err != nil || got == nil || got.UUID != "selected-clone" {
				t.Fatalf("drive=%v error=%v", got, err)
			}
			if d.DriveUUID != "must-not-win" || d.ClonedDriveUUID != "synthetic-drive" {
				t.Fatal("selection mutated driver-owned IDs")
			}
		})
	}
	t.Run("list-refusal", func(t *testing.T) {
		d := fixture(t, step{method: "GET", path: "libdrives/", query: "limit=0&name__icontains=Ubuntu+%26+LTS", status: 403, response: `[{"error_message":"synthetic selection refusal"}]`})
		d.DriveName = "Ubuntu & LTS"
		got, err := d.cloneLibraryDrive()
		if got != nil || err == nil || !strings.Contains(err.Error(), "403") {
			t.Fatalf("drive=%v error=%v", got, err)
		}
	})
}
