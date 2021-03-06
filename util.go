package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/chai2010/webp"
	"github.com/denisbrodbeck/sqip"
)

func someErr(errors chan error) error {
	for i := 0; i < len(errors); i++ {
		if err := <-errors; err != nil {
			return err
		}
	}
	return nil
}

func saveFile(file multipart.File, dest string) error {
	file.Seek(0, io.SeekStart)
	dst, err := os.Create(path.Join(storage, dest))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}

func parseB64SVG(img image.Image) (string, error) {
	svg, _, _, err := sqip.RunLoaded(img, 256, 16, 1, 128, 0, runtime.NumCPU(), "")
	return string("data:image/svg+xml;base64," + sqip.Base64(svg)), err
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	var route = path.Join(storage, r.URL.Path[1:])
	http.ServeFile(w, r, route)
}

func mkDirAll(r *http.Request) error {
	dest := path.Dir(path.Join(storage, r.URL.Path))
	return os.MkdirAll(dest, os.FileMode(int(0776)))
}

func convertAndSaveWEBP(img image.Image, r *http.Request) error {
	var buf bytes.Buffer
	var ops = &webp.Options{Lossless: false, Quality: 100}

	if err := webp.Encode(&buf, img, ops); err != nil {
		return err
	}

	filenameWebp := fmt.Sprintf("%s.webp", r.URL.Path)

	var dest = path.Join(storage, filenameWebp)

	return ioutil.WriteFile(dest, buf.Bytes(), 0666)
}

func location(r *http.Request, ext string) string {
	return fmt.Sprintf("http://%s%s.%s", r.Host, r.URL.Path, ext)
}

func locationAvailable(r *http.Request) error {
	file := path.Join(storage, fmt.Sprintf("%s.webp", r.URL.Path))
	if _, err := os.Stat(file); err == nil {
		return os.ErrExist
	}
	return nil
}

func enpipe(c chan error, f func() error) {
	go func() {
		c <- f()
	}()
}

func repipe(aChan chan string, eChan chan error, f func() (string, error)) {
	go func() {
		a, err := f()
		aChan <- a
		eChan <- err
	}()
}
