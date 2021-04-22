package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
)

var MAX_UPLOAD_SIZE int64 = 10 << 20 // Set max size to 10MB

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%8s %s", r.Method, r.URL.Path)
		if r.Method == "GET" {
			serveFile(w, r)
			return
		}
		if r.Method == "POST" {

			if err := locationAvailable(r); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

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

			if err := mkDirAll(r); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			img, ext, err := image.Decode(file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Start async

			errCh := make(chan error, 3)
			imgCh := make(chan string, 1)
			defer close(errCh)
			defer close(imgCh)

			enpipe(errCh, func() error {
				return saveFile(file, fmt.Sprintf("%s.%s", r.URL.Path, ext))
			})
			repipe(imgCh, errCh, func() (string, error) {
				return parseB64SVG(img)
			})
			enpipe(errCh, func() error {
				return convertAndSaveWEBP(img, r)
			})
			if err := someErr(errCh); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Add("location", location(r, "webp"))
			w.Header().Add("location-webp", location(r, "webp"))
			if ext != "webp" {
				w.Header().Add(fmt.Sprintf("location-%s", ext), location(r, ext))
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, <-imgCh)
			return
		}
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
