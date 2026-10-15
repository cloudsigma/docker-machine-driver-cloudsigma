package cloudsigma

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cloudsigma/cloudsigma-sdk-go/cloudsigma"
	"github.com/docker/machine/libmachine/state"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestMain(m *testing.M) {
	http.DefaultTransport = transportFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("nonfixture network denied") })
	os.Exit(m.Run())
}

type body struct {
	io.Reader
	closed int
}

func (b *body) Close() error { b.closed++; return nil }

type step struct {
	method, path, query, payload, response string
	status                                 int
	failure                                error
}

func fixture(t *testing.T, steps ...step) *Driver {
	t.Helper()
	old := http.DefaultTransport
	i := 0
	bodies := []*body{}
	http.DefaultTransport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if i >= len(steps) {
			t.Fatalf("unexpected dispatch %s %s", r.Method, r.URL)
		}
		s := steps[i]
		i++
		if r.URL.Scheme != "https" || r.URL.Host != "next.cloudsigma.com" || r.Method != s.method || r.URL.Path != "/api/2.0/"+s.path || r.URL.RawQuery != s.query {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		}
		u, p, ok := r.BasicAuth()
		if !ok || u != "fixture" || p != "synthetic-secret" || r.UserAgent() != "docker-machine-driver-cloudsigma" {
			t.Fatal("credentials/user agent")
		}
		if _, ok := r.Context().Deadline(); !ok {
			t.Fatal("missing HTTP timeout deadline")
		}
		var raw []byte
		if r.Body != nil {
			raw, _ = io.ReadAll(r.Body)
		}
		if s.payload != "" {
			var got, want interface{}
			gotDecoder := json.NewDecoder(strings.NewReader(string(raw)))
			gotDecoder.UseNumber()
			wantDecoder := json.NewDecoder(strings.NewReader(s.payload))
			wantDecoder.UseNumber()
			if gotDecoder.Decode(&got) != nil || wantDecoder.Decode(&want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("payload got %s want %s", raw, s.payload)
			}
		} else if len(raw) != 0 {
			t.Fatalf("unexpected body %s", raw)
		}
		if s.failure != nil {
			return nil, s.failure
		}
		b := &body{Reader: strings.NewReader(s.response)}
		bodies = append(bodies, b)
		return &http.Response{StatusCode: s.status, Header: make(http.Header), Body: b, Request: r}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = old
		if i != len(steps) {
			t.Errorf("calls %d want %d", i, len(steps))
		}
		for _, b := range bodies {
			if b.closed != 1 {
				t.Errorf("body close %d", b.closed)
			}
		}
	})
	d := NewDriver("fixture", "")
	d.Username = "fixture"
	d.Password = "synthetic-secret"
	d.APILocation = "next"
	d.ServerUUID = "synthetic-id"
	d.ClonedDriveUUID = "synthetic-drive"
	d.SSHKeyUUID = "synthetic-key"
	d.CPU = 2000
	d.Memory = 1024
	return d
}
func get(status string) step {
	return step{method: "GET", path: "servers/synthetic-id/", status: 200, response: `{"uuid":"synthetic-id","status":"` + status + `"}`}
}
func TestModernGetState(t *testing.T) {
	for _, tc := range []struct {
		s    string
		want state.State
	}{{"running", state.Running}, {"stopped", state.Stopped}, {"paused", state.Paused}, {"starting", state.Starting}, {"stopping", state.Stopping}, {"unknown", state.None}} {
		t.Run(tc.s, func(t *testing.T) {
			d := fixture(t, get(tc.s))
			got, e := d.GetState()
			if e != nil || got != tc.want {
				t.Fatal(got, e)
			}
		})
	}
}
func TestModernGetStateErrors(t *testing.T) {
	for _, s := range []step{{method: "GET", path: "servers/synthetic-id/", status: 500, response: `{"error":"fixture"}`}, {method: "GET", path: "servers/synthetic-id/", status: 200, response: `{`}, {method: "GET", path: "servers/synthetic-id/", failure: context.DeadlineExceeded}} {
		t.Run(s.response, func(t *testing.T) {
			d := fixture(t, s)
			got, e := d.GetState()
			if e == nil || got != state.Error {
				t.Fatal(got, e)
			}
		})
	}
}
func TestModernCreateAndAttach(t *testing.T) {
	d := fixture(t, step{method: "POST", path: "servers/", status: 201, payload: `{"objects":[{"cpu":2000,"mem":1073741824,"name":"fixture","nics":[{"ip_v4_conf":{"conf":"dhcp"},"model":"virtio"}],"pubkeys":[{"uuid":"synthetic-key"}],"vnc_password":"cloudsigma"}]}`, response: `{"objects":[{"uuid":"synthetic-id","cpu":2000,"mem":1073741824,"name":"fixture","vnc_password":"cloudsigma"}]}`}, step{method: "PUT", path: "servers/synthetic-id/", status: 200, payload: `{"cpu":2000,"mem":1073741824,"name":"fixture","vnc_password":"cloudsigma","drives":[{"boot_order":1,"dev_channel":"0:0","device":"virtio","drive":{"uuid":"synthetic-drive"}}]}`, response: `{"uuid":"synthetic-id"}`})
	s, e := d.createServer()
	if e != nil || s.UUID != "synthetic-id" {
		t.Fatal(s, e)
	}
}
func TestModernCreateFailure(t *testing.T) {
	for _, tc := range []struct {
		status   int
		response string
	}{{500, `{"error":"fixture"}`}, {201, `{"objects":[]}`}, {201, `{`}} {
		t.Run(tc.response, func(t *testing.T) {
			d := fixture(t, step{method: "POST", path: "servers/", status: tc.status, response: tc.response, payload: `{"objects":[{"cpu":2000,"mem":1073741824,"name":"fixture","nics":[{"ip_v4_conf":{"conf":"dhcp"},"model":"virtio"}],"pubkeys":[{"uuid":"synthetic-key"}],"vnc_password":"cloudsigma"}]}`})
			s, e := d.createServer()
			if e == nil || s != nil {
				t.Fatal(s, e)
			}
		})
	}
}
func TestModernStartStopRemove(t *testing.T) {
	t.Run("start", func(t *testing.T) {
		last := get("running")
		last.response = `{"status":"running","runtime":{"nics":[{"interface_type":"public","ip_v4":{"uuid":"192.0.2.10"}}]}}`
		d := fixture(t, get("stopped"), step{method: "POST", path: "servers/synthetic-id/action/", payload: `{}`, query: "do=start", status: 202, response: `{}`}, last)
		if e := d.Start(); e != nil || d.IPAddress != "192.0.2.10" {
			t.Fatal(e, d.IPAddress)
		}
	})
	t.Run("stop", func(t *testing.T) {
		d := fixture(t, get("running"), step{method: "POST", path: "servers/synthetic-id/action/", payload: `{}`, query: "do=shutdown", status: 202, response: `{}`}, get("stopped"))
		if e := d.Stop(); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("remove", func(t *testing.T) {
		d := fixture(t, get("stopped"), step{method: "DELETE", path: "keypairs/synthetic-key/", status: 204}, step{method: "DELETE", path: "servers/synthetic-id/", status: 204}, step{method: "DELETE", path: "drives/synthetic-drive/", status: 204})
		if e := d.Remove(); e != nil {
			t.Fatal(e)
		}
	})
}
func TestModernClientRefusals(t *testing.T) {
	for _, kind := range []string{"location", "credentials"} {
		t.Run(kind, func(t *testing.T) {
			d := fixture(t)
			if kind == "location" {
				d.APILocation = "../evil"
			} else {
				d.Password = ""
			}
			if _, e := d.GetState(); e == nil {
				t.Fatal("expected local refusal")
			}
		})
	}
}
func TestModernClientCancellation(t *testing.T) {
	d := fixture(t, step{method: "GET", path: "servers/synthetic-id/", failure: context.DeadlineExceeded})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, e := d.getClient().Servers.Get(ctx, d.ServerUUID)
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
}

var _ *cloudsigma.Client

func TestModernAttachmentPresence(t *testing.T) {
	for _, tc := range []struct{ name, fields, want string }{
		{"explicit-zero-empty", `,"cpu":0,"mem":0,"name":"","vnc_password":""`, `"cpu":0,"mem":0,"name":"","vnc_password":""`},
		{"omitted", "", `"cpu":0,"mem":0,"name":"","vnc_password":""`},
		{"nonzero-large-integer", `,"cpu":2000,"mem":9007199254740993,"name":"quoted\"name","vnc_password":"\\secret"`, `"cpu":2000,"mem":9007199254740993,"name":"quoted\"name","vnc_password":"\\secret"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := fixture(t, step{method: "POST", path: "servers/", status: 201, payload: `{"objects":[{"cpu":2000,"mem":1073741824,"name":"fixture","nics":[{"ip_v4_conf":{"conf":"dhcp"},"model":"virtio"}],"pubkeys":[{"uuid":"synthetic-key"}],"vnc_password":"cloudsigma"}]}`, response: `{"objects":[{"uuid":"synthetic-id"` + tc.fields + `,"status":"stopped","meta":{"not":"echoed"}}]}`}, step{method: "PUT", path: "servers/synthetic-id/", status: 200, payload: `{` + tc.want + `,"drives":[{"boot_order":1,"dev_channel":"0:0","device":"virtio","drive":{"uuid":"synthetic-drive"}}]}`, response: `{"uuid":"synthetic-id"}`})
			s, e := d.createServer()
			if e != nil || s.UUID != "synthetic-id" {
				t.Fatal(s, e)
			}
		})
	}
}
