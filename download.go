package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"github.com/ryanuber/go-glob"
)

func doDownload(bs *bucketService, keys simplecrypto.Keys, filename, destinationDir string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	if len(destinationDir) > 0 {
		if _, err := os.Stat(destinationDir); os.IsNotExist(err) {
			return fmt.Errorf("destination directory: %s does not exist", destinationDir)
		}
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys.EncryptionKey)

	foundFile := false
	for plaintextFilename, _ := range decToEncPaths {
		if glob.Glob(filename, plaintextFilename) {
			foundFile = true

			encryptedFilepath := decToEncPaths[plaintextFilename]
			decryptedFilePath, _ := decryptFilePath(decToEncPaths[plaintextFilename], keys.EncryptionKey)
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

			os.MkdirAll(destinationDir+"/"+filepath.Dir(plaintextFilename), 0777)
			fmt.Println("create dir: ", destinationDir+"/"+filepath.Dir(plaintextFilename))

			if len(destinationDir) == 0 {
				os.Rename(tempDownloadFilename, actualFilename)
				fmt.Println("downloaded file to: " + actualFilename)
			} else {
				os.Rename(tempDownloadFilename, destinationDir+"/"+filepath.Dir(plaintextFilename)+"/"+actualFilename)
				fmt.Println("downloaded file to: " + destinationDir + "/" + filepath.Dir(plaintextFilename) + "/" + actualFilename)
			}

		}
	}
	if !foundFile {
		return errors.New("No files found")
	}
	return nil
}
