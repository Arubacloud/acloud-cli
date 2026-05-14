package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ─── httptest harness ─────────────────────────────────────────────────────────

// arubaTestServer is an httptest-backed fake Aruba API. Tests register routes
// with On*/jsonResponse, then call Client() to get a real aruba.Client whose
// adapters hydrate real wrapper types from the registered responses. (TD-022)
type arubaTestServer struct {
	t      *testing.T
	routes map[string]http.HandlerFunc // keyed by "METHOD /path"
	srv    *httptest.Server
	client aruba.Client
}

// newArubaTestServer stands up the fake API and a real aruba.Client pointed at
// it. The server is torn down automatically via t.Cleanup.
func newArubaTestServer(t *testing.T) *arubaTestServer {
	t.Helper()
	a := &arubaTestServer{
		t:      t,
		routes: map[string]http.HandlerFunc{},
	}
	a.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		if h, ok := a.routes[key]; ok {
			h(w, r)
			return
		}
		t.Errorf("arubaTestServer: unregistered route %q (have: %v)", key, a.routeKeys())
		http.Error(w, "unregistered route", http.StatusNotFound)
	}))
	t.Cleanup(a.srv.Close)

	opts := aruba.NewOptions().
		WithBaseURL(a.srv.URL).
		WithToken("test-token").
		WithNoLogs()
	client, err := aruba.NewClient(opts)
	if err != nil {
		t.Fatalf("newArubaTestServer: %v", err)
	}
	a.client = client
	return a
}

// Client returns the real aruba.Client pointed at the fake server.
func (a *arubaTestServer) Client() aruba.Client { return a.client }

// On registers a handler for an exact "METHOD /path" key.
func (a *arubaTestServer) On(method, path string, h http.HandlerFunc) *arubaTestServer {
	a.routes[method+" "+path] = h
	return a
}

// OnGet registers a GET handler.
func (a *arubaTestServer) OnGet(path string, h http.HandlerFunc) *arubaTestServer {
	return a.On(http.MethodGet, path, h)
}

// OnPost registers a POST handler.
func (a *arubaTestServer) OnPost(path string, h http.HandlerFunc) *arubaTestServer {
	return a.On(http.MethodPost, path, h)
}

// OnPut registers a PUT handler.
func (a *arubaTestServer) OnPut(path string, h http.HandlerFunc) *arubaTestServer {
	return a.On(http.MethodPut, path, h)
}

// OnDelete registers a DELETE handler.
func (a *arubaTestServer) OnDelete(path string, h http.HandlerFunc) *arubaTestServer {
	return a.On(http.MethodDelete, path, h)
}

// OnPatch registers a PATCH handler.
func (a *arubaTestServer) OnPatch(path string, h http.HandlerFunc) *arubaTestServer {
	return a.On(http.MethodPatch, path, h)
}

// routeKeys returns the registered route keys for diagnostic messages.
func (a *arubaTestServer) routeKeys() []string {
	keys := make([]string, 0, len(a.routes))
	for k := range a.routes {
		keys = append(keys, k)
	}
	return keys
}

// jsonResponse writes status + JSON-marshalled body. body is typically a
// types.*Response or types.*List value that the SDK adapter expects.
func jsonResponse(status int, body any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			if err := json.NewEncoder(w).Encode(body); err != nil {
				// This runs inside a live httptest handler; panic surfaces in the test.
				panic("jsonResponse: encode: " + err.Error())
			}
		}
	}
}

// errorResponse writes status + a types.ErrorResponse JSON body so the SDK
// adapter surfaces it as an *aruba.HTTPError (which apiErrFromV2 formats).
func errorResponse(status int, title, detail string) http.HandlerFunc {
	return jsonResponse(status, types.ErrorResponse{
		Title:  &title,
		Detail: &detail,
	})
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// strPtr returns a pointer to s. Convenience for building SDK response structs in tests.
func strPtr(s string) *string { return &s }

// resetCmdFlags walks the entire command tree and marks every flag as "not
// changed". This is required because cobra/pflag tracks a "changed" bit per
// flag and the global rootCmd is shared across test cases: without this reset,
// a flag set in test N remains "seen" in test N+1, causing MarkFlagRequired
// validation to pass even when the flag is absent.
func resetCmdFlags(cmd *cobra.Command) {
	resetFlagSet := func(f *pflag.Flag) {
		f.Changed = false
		_ = f.Value.Set(f.DefValue) //nolint:errcheck
	}
	cmd.Flags().VisitAll(resetFlagSet)
	cmd.PersistentFlags().VisitAll(resetFlagSet)
	for _, sub := range cmd.Commands() {
		resetCmdFlags(sub)
	}
}

// runCmd sets the mock client, executes rootCmd with the given args, then
// resets state. It returns the error from rootCmd.Execute().
func runCmd(mock aruba.Client, args []string) error {
	resetCmdFlags(rootCmd)
	setClientForTesting(mock)
	defer resetClientState()
	defer resetCmdFlags(rootCmd)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// runCmdCapture executes rootCmd with the given args (via runCmd) and returns
// everything written to os.Stdout during the call. Use this instead of runCmd
// when a test needs to assert output content.
func runCmdCapture(mock aruba.Client, args []string) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	runErr := runCmd(mock, args)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	r.Close()
	return buf.String(), runErr
}
