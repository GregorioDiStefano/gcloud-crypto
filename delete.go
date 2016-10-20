package main

import (
	"errors"
	"fmt"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
)

func doDeleteObject(bs *bucketService, keys simplecrypto.Keys, filepath string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys.EncryptionKey)
	for plaintextFilename, _ := range decToEncPaths {
		if filepath == plaintextFilename && plaintextFilename != PASSWORD_CHECK_FILE {
			encryptedFilename := decToEncPaths[plaintextFilename]
			if encryptedFilename == "" {
				return fmt.Errorf("file: %s not found in bucket", filepath)
			}

			if err := bs.deleteObject(encryptedFilename); err != nil {
				return err
			} else {
				fmt.Println("deleted: " + plaintextFilename)
			}
		}
	}
	return nil
}
