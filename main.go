package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/chai2010/webp"
)

func DecodeWebP() {
	var buf bytes.Buffer
	var width, height int
	var data []byte
	var err error
	if data, err = ioutil.ReadFile("image.webp"); err != nil {
		log.Println(err)
	}
	if width, height, _, err = webp.GetInfo(data); err != nil {
		log.Println(err)
	}

	fmt.Printf("width = %d, height = %d\n", width, height)

	// GetMetadata
	if metadata, err := webp.GetMetadata(data, "ICCP"); err != nil {
		fmt.Printf("Metadata: err = %v\n", err)
	} else {
		fmt.Printf("Metadata: %s\n", string(metadata))
	}

	// Decode webp
	m, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		log.Println(err)
	}

	// Encode lossless webp
	if err = webp.Encode(&buf, m, &webp.Options{Lossless: false}); err != nil {
		log.Println(err)
	}
	if err = ioutil.WriteFile("output.webp", buf.Bytes(), 0666); err != nil {
		log.Println(err)
	}

	fmt.Println("Save output.webp ok")
}

func main() {

	// DecodeWebP()
	// return

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		if r.Method == "GET" {
			http.ServeFile(w, r, "image.jpg")
		} else {
			r.ParseMultipartForm(10 << 20) // Set max size to 10MB
			if file, header, err := r.FormFile("file"); err != nil {
				fmt.Println("Error Retrieving the file")
				fmt.Println(err)
				w.WriteHeader(500)
				fmt.Fprintf(w, "Error Retrieving the file\n")
				return
			} else {
				defer file.Close()
				fmt.Printf("Uploaded File: %+v\n", header.Filename)
				fmt.Printf("File Size:     %+v\n", header.Size)
				fmt.Printf("MIME Header:   %+v\n", header.Header["Content-Type"])
				// fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))

				img, text, err := image.Decode(file)
				if err != nil {
					log.Println(err)
				}
				log.Printf("Got string format %s", text)

				var buf bytes.Buffer

				if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 80}); err != nil {
					log.Println(err)
				}

				if err := ioutil.WriteFile("out.webp", buf.Bytes(), 0666); err != nil {
					log.Println(err)
				}
				fmt.Fprintf(w, "Successfully Uploaded File\n")
			}
		}
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
