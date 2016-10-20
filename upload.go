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

// doesFileExist checks that the remote does not have the exact same path and filename already
func doesFileExist(uploadPath string, existingFilesMap map[string]string, keys simplecrypto.Keys) error {
	plainTextUploadPath, _ := decryptFilePath(uploadPath, keys.EncryptionKey)
	for plainTextRemotePath := range existingFilesMap {
		if plainTextUploadPath == plainTextRemotePath {
			return errors.New(fileAlreadyExistsError)
		}
	}
	return nil
}

// findExistingPath is an optimization: reuse already existing encrypted paths instead of
// having the same encrypted path in different objects
func findExistingPath(bs bucketService, keys simplecrypto.Keys, uploadDirectoryPath string) string {
	if objects, err := bs.getObjects(); err == nil {

		decryptedToEncryptedFiles := getDecryptedToEncryptedFileMapping(objects, keys.EncryptionKey)

		for e := range decryptedToEncryptedFiles {
			if filepath.Dir(e) == uploadDirectoryPath {
				fmt.Println("found already existing directory")
				return decryptedToEncryptedFiles[e]
			}
		}
	}
	return ""
}

func doUpload(bs *bucketService, keys simplecrypto.Keys, uploadFile, remoteDirectory string) error {
	encryptedFile, err := simplecrypto.EncryptFile(uploadFile, keys)
	encryptedPath := ""

	if err != nil {
		panic(err)
	}

	fmt.Println("Encrypting complete.")

	objects, err := bs.getObjects()

	if err != nil {
		panic(err)
	}

	fmt.Println("Uploading to: ", remoteDirectory)

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys.EncryptionKey)

	if remoteDirectory == "" && filepath.Dir(uploadFile) == "." {
		encryptedPath = encryptFilePath(uploadFile, keys.EncryptionKey)
	} else {
		finalRemoteUploadPath := filepath.Clean(remoteDirectory + "/" + filepath.Dir(uploadFile) + "/" + filepath.Base(uploadFile))
		finalRemoteUploadDirectoryPath := filepath.Dir(finalRemoteUploadPath)
		fmt.Println("upload directory: ", finalRemoteUploadDirectoryPath, "final upload path: ", finalRemoteUploadPath)
		if matchingDirectory := findExistingPath(*bs, keys, finalRemoteUploadDirectoryPath); matchingDirectory != "" {
			encryptedFilename, _ := simplecrypto.EncryptText(path.Base(uploadFile), keys.EncryptionKey)
			encryptedPath = filepath.Dir(matchingDirectory) + "/" + encryptedFilename
		} else {
			encryptedPath = encryptFilePath(finalRemoteUploadPath, keys.EncryptionKey)
		}
	}

	//check if the file exists
	if err := doesFileExist(encryptedPath, decToEncPaths, keys); err != nil {
		return err
	}

	return bs.uploadToBucket(encryptedFile, encryptedPath)
}

func processUpload(bs *bucketService, keys simplecrypto.Keys, uploadPath, remoteDirectory string) error {
	globMatch, err := filepath.Glob(uploadPath)
	var errorUploading bool

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
