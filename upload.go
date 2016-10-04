package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
)

func doUpload(bs *bucketService, keys simplecrypto.Keys, filePath, remoteDirectory string) error {
	encryptedFile, _ := simplecrypto.EncryptFile(filePath, keys)
	fmt.Print("Encrypting complete.")
	var encryptedPath string

	if objects, err := bs.getObjects(); err == nil {
		fmt.Println("Uploading to: ", remoteDirectory)

		decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys.EncryptionKey)

		for e := range decToEncPaths {
			//instead of creating a new "directory", check if a directory already exists, and use it
			if len(remoteDirectory) > 0 {
				if filepath.Dir(e) == remoteDirectory {
					fmt.Println("using existing directory: ", filepath.Dir(decToEncPaths[e]))
					encryptedFilename, _ := simplecrypto.EncryptText(path.Base(filePath), keys.EncryptionKey)
					encryptedPath = filepath.Dir(decToEncPaths[e]) + "/" + encryptedFilename
					fmt.Println("using already existing encrypted path: " + encryptedPath)
					break
				} else {
					remoteDirectoryFilename := path.Clean(remoteDirectory + "/" + path.Base(filePath))
					encryptedPath = encryptFilePath(remoteDirectoryFilename, keys.EncryptionKey)
				}
			} else {
				encryptedPath = encryptFilePath(filePath, keys.EncryptionKey)
			}
		}

		for e := range decToEncPaths {
			//stupid
			plainTextRemotePath := decryptFilePath(encryptedPath, keys.EncryptionKey)
			if e == plainTextRemotePath {
				return errors.New("This file already exists, delete it first.")
			}
		}
	} else {
		return err
	}

	bs.uploadToBucket(encryptedFile, encryptedPath)
	return nil
}

func processUpload(bs *bucketService, keys simplecrypto.Keys, path, remoteDirectory string) error {
	globMatch, err := filepath.Glob(path)
	var errorUploading bool
	fmt.Println(globMatch, err)
	if err != nil {
		return err
	}

	for _, path := range globMatch {
		fmt.Println("Uploading: " + path + " to " + remoteDirectory)
		if isDir(path) {
			err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if !isDir(path) {
					fmt.Println("Uploading: " + path)
					return doUpload(bs, keys, path, remoteDirectory)
				}
				return nil
			})
		} else {
			err = doUpload(bs, keys, path, remoteDirectory)
		}

		if err != nil {
			errorUploading = true
			fmt.Println(fmt.Sprintf("failed with %s when uploading: %s", err.Error(), path))
		}
	}

	if errorUploading {
		return errors.New("at least one file failed to upload")
	}
	return nil
}
