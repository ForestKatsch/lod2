package storage

import (
	"lod2/config"
	"log"
	"os"
)

func Init() {
	if _, err := os.Stat(config.Config.StoragePath); os.IsNotExist(err) {
		log.Printf("warning: storage directory missing at %s. create this directory to enable file management.", config.Config.StoragePath)
	} else {
		log.Printf("storage ready")
	}
}
