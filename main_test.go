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

func TestGetPlayground(t *testing.T) {
	rec := doRequest(t, "/qr/playground")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("QR Studio")) {
		t.Error("response does not contain the playground")
	}
}

func TestGetRoot_LandingPage(t *testing.T) {
	rec := doRequest(t, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	body := rec.Body.Bytes()
	for _, link := range []string{"/qr/playground", "/microqr/playground", "/rmqr/playground"} {
		if !bytes.Contains(body, []byte(link)) {
			t.Errorf("landing page does not link to %q", link)
		}
	}
}

func TestGetMicroPlayground(t *testing.T) {
	rec := doRequest(t, "/microqr/playground")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Micro QR Studio")) {
		t.Error("response does not contain the Micro QR playground")
	}
}

func TestGetRMQRPlayground(t *testing.T) {
	rec := doRequest(t, "/rmqr/playground")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("rMQR Studio")) {
		t.Error("response does not contain the rMQR playground")
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

func TestGetQR_PNGSizeBelowLargerMinimum(t *testing.T) {
	rec := doRequest(t, "/qr?data=hello&version=40&size=184")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if got, want := rec.Body.String(), "size must be at least 185 for this QR code\n"; got != want {
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

func TestGetMicroQR_PNG(t *testing.T) {
	rec := doRequest(t, "/microqr?data=12345&level=check")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
	img, err := png.Decode(rec.Body)
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}
	if got := img.Bounds().Dx(); got != 15 {
		t.Errorf("image width = %d, want 15", got)
	}
}

func TestGetMicroQR_SVG(t *testing.T) {
	rec := doRequest(t, "/microqr?data=hello&format=svg&size=128&level=L")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/svg+xml" {
		t.Errorf("Content-Type = %q, want image/svg+xml", got)
	}
	if got := rec.Body.Bytes(); !bytes.Contains(got, []byte(`width="128" height="128"`)) {
		t.Errorf("SVG does not contain requested dimensions: %s", got)
	}
}

func TestGetMicroQR_Version(t *testing.T) {
	rec := doRequest(t, "/microqr?data=hello&version=4&level=L")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}
	img, err := png.Decode(rec.Body)
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}
	if got := img.Bounds().Dx(); got != 21 {
		t.Errorf("image width = %d, want 21", got)
	}
}

func TestGetMicroQR_InvalidParameters(t *testing.T) {
	for _, target := range []string{
		"/microqr?data=hello&size=0",
		"/microqr?data=hello&format=gif",
		"/microqr?data=hello&level=H",
		"/microqr?data=hello&version=5",
	} {
		t.Run(target, func(t *testing.T) {
			rec := doRequest(t, target)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestGetMicroQR_PNGSizeBelowMinimum(t *testing.T) {
	rec := doRequest(t, "/microqr?data=12345&level=check&size=14")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if got, want := rec.Body.String(), "size must be at least 15 for this Micro QR code\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestGetRMQR_PNG(t *testing.T) {
	rec := doRequest(t, "/rmqr?data=123456789012")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
	img, err := png.Decode(rec.Body)
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 47 || got.Y != 11 {
		t.Errorf("image size = %v, want 47x11", got)
	}
}

func TestGetRMQR_SVGWithSize(t *testing.T) {
	rec := doRequest(t, "/rmqr?data=hello&format=svg&size=94")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/svg+xml" {
		t.Errorf("Content-Type = %q, want image/svg+xml", got)
	}
	if got := rec.Body.Bytes(); !bytes.Contains(got, []byte(`width="94" height="22" viewBox="0 0 47 11"`)) {
		t.Errorf("SVG does not contain proportional dimensions: %s", got)
	}
}

func TestGetRMQR_Version(t *testing.T) {
	rec := doRequest(t, "/rmqr?data=1&version=R11x27&level=H")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}
	img, err := png.Decode(rec.Body)
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 31 || got.Y != 15 {
		t.Errorf("image size = %v, want 31x15", got)
	}
}

func TestGetRMQR_Priority(t *testing.T) {
	for _, tt := range []struct {
		priority string
		width    int
		height   int
	}{
		{priority: "height", width: 63, height: 11},
		{priority: "width", width: 31, height: 15},
	} {
		t.Run(tt.priority, func(t *testing.T) {
			rec := doRequest(t, "/rmqr?data=1234567890123&priority="+tt.priority)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
			}
			img, err := png.Decode(rec.Body)
			if err != nil {
				t.Fatalf("failed to decode PNG: %v", err)
			}
			if got := img.Bounds().Size(); got.X != tt.width || got.Y != tt.height {
				t.Errorf("image size = %v, want %dx%d", got, tt.width, tt.height)
			}
		})
	}
}

func TestGetRMQR_InvalidParameters(t *testing.T) {
	for _, target := range []string{
		"/rmqr?data=hello&size=0",
		"/rmqr?data=hello&format=gif",
		"/rmqr?data=hello&level=L",
		"/rmqr?data=hello&version=R8x43",
		"/rmqr?data=hello&priority=depth",
	} {
		t.Run(target, func(t *testing.T) {
			rec := doRequest(t, target)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestGetRMQR_PNGSizeBelowMinimum(t *testing.T) {
	rec := doRequest(t, "/rmqr?data=123456789012&size=46")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if got, want := rec.Body.String(), "size must be at least 47 for this rMQR code\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

// Verify that the support and policy routes are reachable from every entry page.
func TestSupportNavigation(t *testing.T) {
	for _, path := range []string{"/", "/qr/playground", "/microqr/playground", "/rmqr/playground", "/docs", "/contact", "/terms", "/privacy"} {
		t.Run(path, func(t *testing.T) {
			rec := doRequest(t, path)
			if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "text/html; charset=utf-8" {
				t.Fatalf("unexpected page response: %d %s", rec.Code, rec.Header().Get("Content-Type"))
			}
			for _, link := range []string{`href="/contact"`, `href="/terms"`, `href="/privacy"`} {
				if !bytes.Contains(rec.Body.Bytes(), []byte(link)) {
					t.Errorf("missing navigation link %s", link)
				}
			}
		})
	}
}
