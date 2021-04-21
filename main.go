package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"runtime"
	"time"

	"github.com/chai2010/webp"
	"github.com/denisbrodbeck/sqip"
)

func serveFile(w http.ResponseWriter, r *http.Request) {
	var route = path.Join(storage, r.URL.Path[1:])
	log.Println("Using route", route)
	http.ServeFile(w, r, route)
}

func writeError(w http.ResponseWriter, err error, code ...int) {
	var c = 500
	if len(code) > 0 {
		c = code[0]
	}
	w.WriteHeader(c)
	fmt.Fprintf(w, err.Error())
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		if r.Method == "GET" {
			serveFile(w, r)
			return
		}
		if r.Method == "POST" {
			f := r.URL.Query().Get("f")
			if f == "" {
				f = "file"
			}
			r.ParseMultipartForm(10 << 20) // Set max size to 10MB
			file, _, err := r.FormFile(f)
			if err != nil {
				writeError(w, err)
				return
			}
			defer file.Close()

			img, text, err := image.Decode(file)
			if err != nil {
				writeError(w, err)
				return
			}
			log.Printf("Got string format %s", text)

			svg, width, height, err := sqip.RunLoaded(img, 256, 16, 1, 128, 0, runtime.NumCPU(), "")
			if err != nil {
				writeError(w, err)
				return
			}

			log.Printf("Got SVG of size %dx%d", width, height)

			var img64 = "data:image/svg+xml;base64," + sqip.Base64(svg)

			log.Printf("Got SVG Image in BASE64")

			var buf bytes.Buffer
			var ops = &webp.Options{Lossless: false, Quality: 80}

			if err := webp.Encode(&buf, img, ops); err != nil {
				writeError(w, err)
				return
			}

			log.Printf("Encoded WEBP")

			var dest = path.Join(storage, "out.webp")
			if err := ioutil.WriteFile(dest, buf.Bytes(), 0666); err != nil {
				writeError(w, err)
			}
			log.Printf("Writed WEBP at %s", dest)

			time.Sleep(1400 * time.Millisecond)

			var location = fmt.Sprintf("http://%s/%s", r.Host, dest)
			w.Header().Add("location", location)
			w.WriteHeader(201)
			fmt.Fprintf(w, "Successfully Uploaded File\n\n%s\n", img64)
			return
		}
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
