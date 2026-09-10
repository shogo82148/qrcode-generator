package main

import (
	"bytes"
	"flag"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func doRequest(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()
	g := NewGenerator()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)
	return rec
}

func TestGetQR_PNG(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
	if _, err := png.Decode(rec.Body); err != nil {
		t.Errorf("failed to decode PNG: %v", err)
	}
}

func TestGetQR_InvalidFormat(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&format=gif")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetQR_SVG(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&format=svg")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/svg+xml" {
		t.Errorf("Content-Type = %q, want image/svg+xml", got)
	}

	// The SVG output is deterministic for a given input, so compare it against
	// a golden file. Regenerate it with `go test -update` and eyeball the
	// result to confirm it is still a scannable QR Code.
	got := rec.Body.Bytes()
	golden := filepath.Join("testdata", "hello.svg")
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("failed to read golden file (run `go test -update` to create it): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("SVG output does not match %s; run `go test -update` to regenerate if the change is intentional", golden)
	}
}
