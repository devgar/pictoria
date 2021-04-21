package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/chai2010/webp"
	"github.com/denisbrodbeck/sqip"
)

var MAX_UPLOAD_SIZE int64 = 10 << 20 // Set max size to 10MB

func serveFile(w http.ResponseWriter, r *http.Request) {
	var route = path.Join(storage, r.URL.Path[1:])
	log.Println("Using route", route)
	http.ServeFile(w, r, route)
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		if r.Method == "GET" {
			serveFile(w, r)
			return
		}
		if r.Method == "POST" {

			if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			f := r.URL.Query().Get("f")
			if f == "" {
				f = "file"
			}
			file, _, err := r.FormFile(f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()

			img, ext, err := image.Decode(file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("Got string format %s", ext)

			svg, _, _, err := sqip.RunLoaded(img, 256, 16, 1, 128, 0, runtime.NumCPU(), "")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var img64 = "data:image/svg+xml;base64," + sqip.Base64(svg)

			var buf bytes.Buffer
			var ops = &webp.Options{Lossless: false, Quality: 80}

			if err := webp.Encode(&buf, img, ops); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			filename := fmt.Sprintf("%s.webp", r.URL.Path)

			var dest = path.Join(storage, filename)
			if err := os.MkdirAll(path.Dir(dest), os.FileMode(int(0776))); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("Created %s", path.Dir(dest))
			if err := ioutil.WriteFile(dest, buf.Bytes(), 0666); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("Writed WEBP at %s", dest)

			var location = fmt.Sprintf("http://%s%s", r.Host, filename)
			log.Printf("Accessible by %s", location)
			w.Header().Add("location", location)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "%s", img64)
			return
		}
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
