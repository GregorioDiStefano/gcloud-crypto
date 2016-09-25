package main

import (
	"errors"
	"fmt"

	"github.com/ryanuber/go-glob"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
)

func doDeleteObject(bs *bucketService, keys simplecrypto.Keys, filepath string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	encToDecPaths := getEncryptedToDecryptedMap(objects, keys.EncryptionKey)
	for k, _ := range encToDecPaths {

		if glob.Glob(filepath, k) && k != "keycheck" {
			encryptedFilename := encToDecPaths[k]
			if encryptedFilename == "" {
				return fmt.Errorf("file: %s not found in bucket", filepath)
			}

			if err := bs.deleteObject(encryptedFilename); err != nil {
				return err
			} else {
				fmt.Println("deleted: " + k)
			}
		}
	}
	return nil
}
