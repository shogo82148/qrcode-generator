package main

import (
	"bytes"
	"image/png"
	"net/http"
	"strconv"

	"github.com/shogo82148/qrcode"
	"github.com/shogo82148/ridgenative"
)

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
	mux.HandleFunc("GET /qr", g.getQR)
	return g
}

func (g *Generator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func (g *Generator) getQR(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	data := q.Get("data")
	opts := make([]qrcode.EncodeOptions, 0, 4)

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
	w.Write(buf.Bytes())
}
