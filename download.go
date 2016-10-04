package main

import (
	"errors"
	"fmt"
	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"github.com/ryanuber/go-glob"
	"os"
	"path"
)

func doDownload(bs *bucketService, keys simplecrypto.Keys, filename string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys.EncryptionKey)

	foundFile := false
	for k, _ := range decToEncPaths {
		if glob.Glob(filename, k) {
			foundFile = true

			encryptedFilepath := decToEncPaths[k]
			decryptedFilePath := decryptFilePath(decToEncPaths[k], keys.EncryptionKey)
			downloadedEncryptedFile, err := bs.downloadFromBucket(encryptedFilepath)

			if err != nil {
				return err
			}

			defer os.Remove(downloadedEncryptedFile)

			downloadedPlaintextFile, err := simplecrypto.DecryptFile(downloadedEncryptedFile, keys)
			defer os.Remove(downloadedPlaintextFile)

			if err != nil {
				return err
			}

			_, tempDownloadFilename := path.Split(downloadedPlaintextFile)
			_, actualFilename := path.Split(decryptedFilePath)

			os.Rename(tempDownloadFilename, actualFilename)
			fmt.Println("downloaded: " + actualFilename)
		}
	}
	if !foundFile {
		return errors.New("No files found")
	}
	return nil
}
