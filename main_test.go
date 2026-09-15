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

func TestGetQR_PNGWithSize(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&size=256")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	img, err := png.Decode(rec.Body)
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 256 || got.Y != 256 {
		t.Errorf("image size = %v, want 256x256", got)
	}
}

func TestGetQR_InvalidSize(t *testing.T) {
	for _, size := range []string{"", "abc", "0", "-1", "4097"} {
		t.Run(size, func(t *testing.T) {
			rec := doRequest(t, "/qr?data=hello&size="+size)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestGetQR_PNGSizeBelowMinimum(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&size=1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if got, want := rec.Body.String(), "size must be at least 29 for this QR code\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestGetQR_PNGMinimumSize(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&size=29")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	img, err := png.Decode(rec.Body)
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 29 || got.Y != 29 {
		t.Errorf("image size = %v, want 29x29", got)
	}
}

func TestGetQR_MaxSize(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&format=svg&size=4096")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); !bytes.Contains([]byte(got), []byte(`width="4096" height="4096"`)) {
		t.Errorf("SVG does not contain maximum dimensions: %s", got)
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

func TestGetQR_SVGWithSize(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&format=svg&size=256")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); !bytes.Contains([]byte(got), []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256"`)) {
		t.Errorf("SVG does not contain requested dimensions: %s", got)
	}
}

func TestGetQR_SVGAllowsSizeBelowPNGMinimum(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&format=svg&size=1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); !bytes.Contains([]byte(got), []byte(`width="1" height="1"`)) {
		t.Errorf("SVG does not contain requested dimensions: %s", got)
	}
}
