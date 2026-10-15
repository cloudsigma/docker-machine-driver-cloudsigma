package cloudsigma

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	gossh "golang.org/x/crypto/ssh"
)

// Actual Create with real in-process libmachine RSA generation. These keys are
// disposable local fixtures, never installed or used to authenticate anywhere.
// Filesystem outcomes below characterize pinned libmachine, not API guarantees.
func TestGeneratedKeyCreate(t *testing.T) {
	for _, scenario := range []string{"register-error", "clone-error", "missing-parent", "parent-is-file", "public-is-directory", "existing-public"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Chmod(root, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "key")
			steps := []step{}
			if scenario == "register-error" || scenario == "existing-public" {
				steps = append(steps, step{method: "POST", path: "keypairs/", status: 403, response: `{"error_message":"fixture refusal"}`, payload: `{}`})
			}
			if scenario == "clone-error" {
				steps = append(steps, step{method: "POST", path: "keypairs/", status: 201, response: `{"objects":[{"uuid":"generated-key"}]}`, payload: `{}`}, step{method: "POST", path: "libdrives/synthetic-library/action/", query: "do=clone", status: 403, response: `{"error_message":"fixture refusal"}`, payload: `{"name":"fixture","size":21474836480,"storage_type":"dssd","uuid":""}`})
			}
			switch scenario {
			case "missing-parent":
				path = filepath.Join(root, "missing", "key")
			case "parent-is-file":
				if err := os.WriteFile(path, []byte("external sentinel"), 0600); err != nil {
					t.Fatal(err)
				}
				path = filepath.Join(path, "key")
			case "public-is-directory":
				if err := os.Mkdir(path+".pub", 0700); err != nil {
					t.Fatal(err)
				}
			case "existing-public":
				if err := os.WriteFile(path+".pub", []byte("external public sentinel"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			d := fixture(t, steps...)
			d.SSHKeyPath = path
			d.SSHKeyUUID, d.ClonedDriveUUID, d.ServerUUID = "", "", ""
			d.DriveUUID, d.DriveSize = "synthetic-library", 20
			original := http.DefaultTransport
			calls := 0
			var private, public []byte
			http.DefaultTransport = transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					private, public = generatedPair(t, path)
					raw, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}
					var got struct {
						Objects []struct {
							Name      string `json:"name"`
							PublicKey string `json:"public_key"`
						} `json:"objects"`
					}
					if err := json.Unmarshal(raw, &got); err != nil {
						t.Fatal(err)
					}
					var shape map[string]json.RawMessage
					if err := json.Unmarshal(raw, &shape); err != nil {
						t.Fatal(err)
					}
					if len(shape) != 1 || len(got.Objects) != 1 || got.Objects[0].Name != "fixture" || got.Objects[0].PublicKey != strings.TrimRight(string(public), "\r\n") {
						t.Fatal("generated public request mismatch")
					}
					var objects []map[string]json.RawMessage
					if err := json.Unmarshal(shape["objects"], &objects); err != nil || len(objects[0]) != 2 {
						t.Fatal("unexpected key fields", err)
					}
					// The inherited strict transport checks every other request property.
					r.Body = io.NopCloser(strings.NewReader(`{}`))
				}
				return original.RoundTrip(r)
			})
			err := d.Create()
			if err == nil {
				t.Fatal("expected bounded failure")
			}
			wantCalls := len(steps)
			if calls != wantCalls || d.ServerUUID != "" || d.ClonedDriveUUID != "" || d.SSHKeyPath != path {
				t.Fatal("unexpected mutation or dispatch")
			}
			wantID := ""
			if scenario == "clone-error" {
				wantID = "generated-key"
			}
			if d.SSHKeyUUID != wantID {
				t.Fatal("ownership marker", d.SSHKeyUUID)
			}
			switch scenario {
			case "missing-parent", "parent-is-file":
				wantError := "Error writing keys to file(s)"
				if scenario == "parent-is-file" {
					wantError = "Desired directory for SSH keys does not exist"
				}
				if !strings.Contains(err.Error(), wantError) {
					t.Fatal(err)
				}
				if _, e := os.Stat(path); e == nil {
					t.Fatal("unexpected private file")
				}
				if scenario == "missing-parent" {
					if _, e := os.Stat(filepath.Dir(path)); !errors.Is(e, os.ErrNotExist) {
						t.Fatal("implicit parent creation", e)
					}
				} else {
					raw, e := os.ReadFile(filepath.Dir(path))
					if e != nil || string(raw) != "external sentinel" {
						t.Fatal("external file changed", e)
					}
				}
			case "public-is-directory":
				if !strings.Contains(err.Error(), "Error writing keys to file(s)") {
					t.Fatal(err)
				}
				raw, e := os.ReadFile(path)
				if e != nil || len(raw) == 0 {
					t.Fatal("partial private file unexpectedly cleaned", e)
				}
				info, e := os.Stat(path + ".pub")
				if e != nil || !info.IsDir() {
					t.Fatal("external directory changed", e)
				}
			default:
				afterPrivate, afterPublic := generatedPair(t, path)
				if !bytes.Equal(private, afterPrivate) || !bytes.Equal(public, afterPublic) {
					t.Fatal("failure unexpectedly changed or cleaned generated files")
				}
			}
		})
	}
}

func generatedPair(t *testing.T, path string) ([]byte, []byte) {
	t.Helper()
	private, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	block, rest := pem.Decode(private)
	if block == nil || block.Type != "RSA PRIVATE KEY" || len(rest) != 0 {
		t.Fatal("private encoding")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if err := key.Validate(); err != nil {
		t.Fatal(err)
	}
	if key.N.BitLen() != 2048 {
		t.Fatal("RSA size")
	}
	public, err := os.ReadFile(path + ".pub")
	if err != nil {
		t.Fatal(err)
	}
	pub, _, _, rest, err := gossh.ParseAuthorizedKey(public)
	if err != nil || len(rest) != 0 {
		t.Fatal("public encoding", err)
	}
	actual, ok := pub.(gossh.CryptoPublicKey)
	if !ok {
		t.Fatal("public crypto type")
	}
	rsaPub, ok := actual.CryptoPublicKey().(*rsa.PublicKey)
	if !ok || rsaPub.E != key.E || rsaPub.N.Cmp(key.N) != 0 {
		t.Fatal("key pair mismatch")
	}
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "openbsd":
		for _, f := range []string{path, path + ".pub"} {
			info, e := os.Stat(f)
			if e != nil || info.Mode().Perm() != 0600 {
				t.Fatal("pinned library mode must be 0600", e)
			}
		}
	}
	return private, public
}
