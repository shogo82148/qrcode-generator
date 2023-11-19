package main

import (
	"bytes"
	"image/png"
	"net/http"

	"github.com/shogo82148/qrcode"
	"github.com/shogo82148/ridgenative"
)

func main() {
	http.HandleFunc("/qr", func(w http.ResponseWriter, r *http.Request) {
		data := []byte(r.URL.Query().Get("data"))
		img, err := qrcode.Encode(data, qrcode.WithLevel(qrcode.LevelL))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		err = png.Encode(&buf, img)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(buf.Bytes())
	})
	ridgenative.ListenAndServe(":8080", nil)
}
