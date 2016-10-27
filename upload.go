package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	_ "github.com/GregorioDiStefano/go-file-storage/log"
	"github.com/Sirupsen/logrus"
)

const (
	fileNotFoundError      = "That file did not exist locally"
	fileAlreadyExistsError = "This file already exists at this location"
	fileUploadFailError    = "at least one file failed to upload"
)

// doesFileExist checks that the remote does not have the exact same path and filename already
func doesFileExist(uploadPath string, existingFilesMap map[string]string, keys *simplecrypto.Keys) error {
	plainTextUploadPath, _ := decryptFilePath(uploadPath, keys)
	for plainTextRemotePath := range existingFilesMap {
		if plainTextUploadPath == plainTextRemotePath {
			return errors.New(fileAlreadyExistsError)
		}
	}
	return nil
}

// findExistingPath is an optimization: reuse already existing encrypted paths instead of
// having the same encrypted path in different objects
func findExistingPath(bs bucketService, keys *simplecrypto.Keys, uploadDirectoryPath string) string {
	if objects, err := bs.getObjects(); err == nil {

		decryptedToEncryptedFiles := getDecryptedToEncryptedFileMapping(objects, keys)

		for e := range decryptedToEncryptedFiles {
			if filepath.Dir(e) == uploadDirectoryPath {
				return decryptedToEncryptedFiles[e]
			}
		}
	}
	return ""
}

func doUpload(bs *bucketService, keys *simplecrypto.Keys, uploadFile, remoteDirectory string) error {
	encryptedPath := ""
	objects, err := bs.getObjects()

	if err != nil {
		panic(err)
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys)

	if remoteDirectory == "" && filepath.Dir(uploadFile) == "." {
		encryptedPath = encryptFilePath(uploadFile, keys)
	} else {

		if strings.HasPrefix(remoteDirectory, "/") {
			remoteDirectory = remoteDirectory[1:]
		}

		finalRemoteUploadPath := filepath.Clean(remoteDirectory + "/" + filepath.Dir(uploadFile) + "/" + filepath.Base(uploadFile))
		finalRemoteUploadDirectoryPath := filepath.Dir(finalRemoteUploadPath)

		if matchingDirectory := findExistingPath(*bs, keys, finalRemoteUploadDirectoryPath); matchingDirectory != "" {
			if encryptedFilename, err := simplecrypto.EncryptText(path.Base(uploadFile), keys.EncryptionKey); err == nil {
				encryptedPath = filepath.Dir(matchingDirectory) + "/" + encryptedFilename
			} else {
				return err
			}
		} else {
			encryptedPath = encryptFilePath(finalRemoteUploadPath, keys)
		}
	}

	//check if the file exists
	if err := doesFileExist(encryptedPath, decToEncPaths, keys); err != nil {
		return err
	}

	log.WithFields(logrus.Fields{"filename": uploadFile}).Debug("Starting encryption of file.")
	encryptedFile, md5Hash, err := simplecrypto.EncryptFile(uploadFile, keys)

	if err != nil {
		panic(err)
	}

	log.WithFields(logrus.Fields{"filename": uploadFile}).Debug("Encryption of file complete.")

	log.WithFields(logrus.Fields{"filename": uploadFile, "remoteDirectory": remoteDirectory}).Info("Uploading file")
	return bs.uploadToBucket(encryptedFile, keys, md5Hash, encryptedPath)
}

func processUpload(bs *bucketService, keys *simplecrypto.Keys, uploadPath, remoteDirectory string) error {
	globMatch, err := filepath.Glob(uploadPath)
	errorOccured := false

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
					doUpload(bs, keys, walkPath, remoteDirectory)
				}
				return nil
			})
		} else {
			err = doUpload(bs, keys, path, remoteDirectory)
		}

		if err != nil {
			errorOccured = true
			switch err.Error() {
			case fileAlreadyExistsError:
				fmt.Println("file already exists, skipping upload.")
				os.Remove(uploadPath + ".enc")
			default:
				fmt.Println(fmt.Sprintf("failed with %s when uploading: %s", err.Error(), path))
				os.Remove(uploadPath + ".enc")
				return err
			}
		}
	}

	if errorOccured {
		return errors.New(fileUploadFailError)
	}

	return nil
}
