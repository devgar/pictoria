package main

import (
	"log"
	"os"
	"path"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
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

func init() {
	initStorage()
}
