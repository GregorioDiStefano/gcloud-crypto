package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
)

const (
	fileNotFoundError      = "That file did not exist locally"
	fileAlreadyExistsError = "This file already exists at this location"
	fileUploadFailError    = "at least one file failed to upload"
)

func doUpload(bs *bucketService, keys simplecrypto.Keys, uploadFile, remoteDirectory string) error {
	encryptedFile, err := simplecrypto.EncryptFile(uploadFile, keys)
	if err != nil {
		return errors.New("failed to encrypt file: " + uploadFile)
	}
	fmt.Println("Encrypting complete.")
	var encryptedPath string

	if objects, err := bs.getObjects(); err == nil {
		fmt.Println("Uploading to: ", remoteDirectory)

		decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys.EncryptionKey)

		for e := range decToEncPaths {
			//instead of creating a new "directory", check if a directory already exists, and use it
			if len(remoteDirectory) > 0 {
				if filepath.Dir(e) == remoteDirectory {
					fmt.Println("using existing directory: ", filepath.Dir(decToEncPaths[e]))
					encryptedFilename, _ := simplecrypto.EncryptText(path.Base(uploadFile), keys.EncryptionKey)
					encryptedPath = filepath.Dir(decToEncPaths[e]) + "/" + encryptedFilename
					fmt.Println("using already existing encrypted path: " + encryptedPath)
					break
				} else {
					remoteDirectoryFilename := path.Clean(remoteDirectory + "/" + uploadFile)
					encryptedPath = encryptFilePath(remoteDirectoryFilename, keys.EncryptionKey)
				}
			} else {
				encryptedPath = encryptFilePath(uploadFile, keys.EncryptionKey)
			}
		}

		// we didnt loop above, since the bucket is empty
		if len(decToEncPaths) == 0 {
			if len(remoteDirectory) > 0 {
				//TODO: remove duplicated code
				remoteDirectoryFilename := path.Clean(remoteDirectory + "/" + uploadFile)
				encryptedPath = encryptFilePath(remoteDirectoryFilename, keys.EncryptionKey)
			} else {
				encryptedPath = encryptFilePath(uploadFile, keys.EncryptionKey)
			}
		}

		for e := range decToEncPaths {
			//stupid
			plainTextRemotePath := decryptFilePath(encryptedPath, keys.EncryptionKey)
			if e == plainTextRemotePath {
				return errors.New(fileAlreadyExistsError)
			}
		}
	} else {
		return err
	}

	bs.uploadToBucket(encryptedFile, encryptedPath)
	return nil
}

func processUpload(bs *bucketService, keys simplecrypto.Keys, uploadPath, remoteDirectory string) error {
	globMatch, err := filepath.Glob(uploadPath)
	var errorUploading bool
	fmt.Println(globMatch, err)
	if err != nil {
		return err
	}

	if len(globMatch) == 0 {
		return errors.New(fileNotFoundError)
	}

	for _, path := range globMatch {
		fmt.Println("Uploading: " + path + " to " + remoteDirectory)
		if isDir(path) {
			err = filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
				if !isDir(walkPath) {
					return doUpload(bs, keys, walkPath, remoteDirectory)
				}
				return nil
			})
		} else {
			err = doUpload(bs, keys, path, remoteDirectory)
		}

		if err != nil {
			errorUploading = true
			fmt.Println(fmt.Sprintf("failed with %s when uploading: %s", err.Error(), path))
			os.Remove(uploadPath + ".enc")
		}
	}

	if errorUploading {
		return errors.New(fileUploadFailError)
	}
	return nil
}
