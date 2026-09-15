package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"strconv"

	"github.com/shogo82148/go-imaging/bitmap"
	"github.com/shogo82148/qrcode"
	"github.com/shogo82148/qrcode/microqr"
	"github.com/shogo82148/ridgenative"
)

// quietZone is the number of modules of the margin around the QR Code.
const quietZone = 4

// microQuietZone is the number of modules required around a Micro QR Code.
const microQuietZone = 2

// maxSize limits memory and CPU consumption when rendering an image.
const maxSize = 4096

//go:embed index.html
var playgroundHTML []byte

func main() {
	g := NewGenerator()
	ridgenative.ListenAndServe(":8080", g)
}

type Generator struct {
	mux *http.ServeMux
}

func NewGenerator() *Generator {
	mux := http.NewServeMux()
	g := &Generator{
		mux: mux,
	}
	mux.HandleFunc("GET /{$}", g.getPlayground)
	mux.HandleFunc("GET /qr", g.getQR)
	mux.HandleFunc("GET /microqr", g.getMicroQR)
	return g
}

func (g *Generator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func (g *Generator) getPlayground(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(playgroundHTML)))
	_, _ = w.Write(playgroundHTML)
}

func (g *Generator) getQR(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	data := q.Get("data")
	opts := make([]qrcode.EncodeOptions, 0, 4)
	size := 0
	if q.Has("size") {
		var err error
		size, err = strconv.Atoi(q.Get("size"))
		if err != nil || size <= 0 || size > maxSize {
			http.Error(w, "invalid size", http.StatusBadRequest)
			return
		}
	}

	format := "png"
	if q.Has("format") {
		format = q.Get("format")
		switch format {
		case "png", "svg":
			// ok
		default:
			http.Error(w, "invalid format", http.StatusBadRequest)
			return
		}
	}

	if q.Has("level") {
		level := q.Get("level")
		switch level {
		case "L", "l":
			opts = append(opts, qrcode.WithLevel(qrcode.LevelL))
		case "M", "m":
			opts = append(opts, qrcode.WithLevel(qrcode.LevelM))
		case "Q", "q":
			opts = append(opts, qrcode.WithLevel(qrcode.LevelQ))
		case "H", "h":
			opts = append(opts, qrcode.WithLevel(qrcode.LevelH))
		default:
			http.Error(w, "invalid level", http.StatusBadRequest)
			return
		}
	}

	version := qrcode.Version(0)
	if q.Has("version") {
		v, err := strconv.Atoi(q.Get("version"))
		if err != nil {
			http.Error(w, "invalid version", http.StatusBadRequest)
			return
		}
		if v < 1 || v > 40 {
			http.Error(w, "invalid version", http.StatusBadRequest)
			return
		}
		version = qrcode.Version(v)
	}

	qr, err := qrcode.New([]byte(data), opts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if version > qr.Version {
		qr.Version = version
	}

	if format == "svg" {
		binimg, err := qr.EncodeToBitmap()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		g.writeSVG(w, binimg, size, quietZone)
		return
	}
	if size > 0 {
		binimg, err := qr.EncodeToBitmap()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		minimumSize := binimg.Bounds().Dx() + quietZone*2
		if size < minimumSize {
			http.Error(w, fmt.Sprintf("size must be at least %d for this QR code", minimumSize), http.StatusBadRequest)
			return
		}
		opts = append(opts, qrcode.WithWidth(size))
	}

	img, err := qr.Encode(opts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = w.Write(buf.Bytes())
}

func (g *Generator) getMicroQR(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	data := q.Get("data")
	opts := []microqr.EncodeOptions{microqr.WithQuietZone(microQuietZone)}

	size, err := parseSize(q.Get("size"), q.Has("size"))
	if err != nil {
		http.Error(w, "invalid size", http.StatusBadRequest)
		return
	}
	format, err := parseFormat(q.Get("format"), q.Has("format"))
	if err != nil {
		http.Error(w, "invalid format", http.StatusBadRequest)
		return
	}

	if q.Has("level") {
		switch q.Get("level") {
		case "check", "CHECK", "Check":
			opts = append(opts, microqr.WithLevel(microqr.LevelCheck))
		case "L", "l":
			opts = append(opts, microqr.WithLevel(microqr.LevelL))
		case "M", "m":
			opts = append(opts, microqr.WithLevel(microqr.LevelM))
		case "Q", "q":
			opts = append(opts, microqr.WithLevel(microqr.LevelQ))
		default:
			http.Error(w, "invalid level", http.StatusBadRequest)
			return
		}
	}

	version := microqr.Version(0)
	if q.Has("version") {
		v, err := strconv.Atoi(q.Get("version"))
		if err != nil || v < 1 || v > 4 {
			http.Error(w, "invalid version", http.StatusBadRequest)
			return
		}
		version = microqr.Version(v)
	}

	qr, err := microqr.New([]byte(data), opts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if version > qr.Version {
		qr.Version = version
	}

	binimg, err := qr.EncodeToBitmap()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if format == "svg" {
		g.writeSVG(w, binimg, size, microQuietZone)
		return
	}
	if size > 0 {
		minimumSize := binimg.Bounds().Dx() + microQuietZone*2
		if size < minimumSize {
			http.Error(w, fmt.Sprintf("size must be at least %d for this Micro QR code", minimumSize), http.StatusBadRequest)
			return
		}
		opts = append(opts, microqr.WithWidth(size))
	}
	img, err := qr.Encode(opts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writePNG(w, img)
}

func parseSize(value string, present bool) (int, error) {
	if !present {
		return 0, nil
	}
	size, err := strconv.Atoi(value)
	if err != nil || size <= 0 || size > maxSize {
		return 0, fmt.Errorf("invalid size")
	}
	return size, nil
}

func parseFormat(value string, present bool) (string, error) {
	if !present {
		return "png", nil
	}
	if value != "png" && value != "svg" {
		return "", fmt.Errorf("invalid format")
	}
	return value, nil
}

func writePNG(w http.ResponseWriter, img image.Image) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = w.Write(buf.Bytes())
}

type binaryImage interface {
	Bounds() image.Rectangle
	BinaryAt(x, y int) bitmap.Color
}

// writeSVG encodes the QR Code as an SVG image and writes it to w.
func (g *Generator) writeSVG(w http.ResponseWriter, binimg binaryImage, requestedSize, quietZone int) {
	bounds := binimg.Bounds()
	// size is the number of modules including the quiet zone on both sides.
	size := bounds.Dx() + quietZone*2

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	buf.WriteString(`<svg xmlns="http://www.w3.org/2000/svg"`)
	if requestedSize > 0 {
		fmt.Fprintf(&buf, ` width="%d" height="%d"`, requestedSize, requestedSize)
	}
	fmt.Fprintf(&buf, ` viewBox="0 0 %d %d" shape-rendering="crispEdges">`+"\n", size, size)
	fmt.Fprintf(&buf, `<rect width="%d" height="%d" fill="#ffffff"/>`+"\n", size, size)

	// Draw all dark modules as a single path for a compact output.
	buf.WriteString(`<path fill="#000000" d="`)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if binimg.BinaryAt(x, y) {
				fmt.Fprintf(&buf, "M%d %dh1v1h-1z",
					x-bounds.Min.X+quietZone, y-bounds.Min.Y+quietZone)
			}
		}
	}
	buf.WriteString(`"/>` + "\n")
	buf.WriteString(`</svg>` + "\n")

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = w.Write(buf.Bytes())
}
