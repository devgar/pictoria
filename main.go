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
	"syscall"

	"github.com/chai2010/webp"
)

var storage string

func initStorage() {
	storage = os.Getenv("STORAGE")
	if storage == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			storage = path.Join(homedir, ".pictoria")
		} else {
			log.Panic("Can't choose storage", err)
		}
	}
	if err := os.MkdirAll(storage, os.ModeSetuid); err != nil {
		log.Panic("Error creating storage", err)
	}

	if err := syscall.Access(storage, syscall.O_RDWR); err != nil {
		log.Panic("Can't create on desired storage ", err)
	}

	log.Println("Storage in", storage)
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	var route = path.Join(storage, r.URL.Path[1:])
	log.Println("Using route", route)
	http.ServeFile(w, r, route)
}

func main() {

	initStorage()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		if r.Method == "GET" {
			serveFile(w, r)
		} else {
			r.ParseMultipartForm(10 << 20) // Set max size to 10MB
			if file, _, err := r.FormFile("file"); err != nil {
				fmt.Println("Error Retrieving the file")
				fmt.Println(err)
				w.WriteHeader(500)
				fmt.Fprintf(w, "Error Retrieving the file\n")
				return
			} else {
				defer file.Close()
				// fmt.Printf("Uploaded File: %+v\n", header.Filename)
				// fmt.Printf("File Size:     %+v\n", header.Size)
				// fmt.Printf("MIME Header:   %+v\n", header.Header["Content-Type"])
				// fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))

				img, text, err := image.Decode(file)
				if err != nil {
					log.Println(err)
				}
				log.Printf("Got string format %s", text)

				var buf bytes.Buffer
				var ops = &webp.Options{Lossless: false, Quality: 80}

				if err := webp.Encode(&buf, img, ops); err != nil {
					log.Println(err)
				}

				var dest = path.Join(storage, "out.webp")
				if err := ioutil.WriteFile(dest, buf.Bytes(), 0666); err != nil {
					log.Println(err)
				}
				var location = fmt.Sprintf("http://%s/%s", r.Host, dest)
				w.Header().Add("location", location)
				w.WriteHeader(201)
				fmt.Fprintf(w, "Successfully Uploaded File\n")
			}
		}
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
