package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
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
			log.Debugf("this file already exists remotely: %s", plainTextRemotePath)
			return errors.New(fileAlreadyExistsError)
		}
	}
	return nil
}

// findExistingPath is an optimization: reuse already existing encrypted paths instead of
// having the same encrypted path in different objects
func findExistingPath(bs bucketService, keys *simplecrypto.Keys, uploadDirectoryPath string) string {
	for encryptedPath, decryptedPath := range bs.bucketCache.seenFiles {
		if filepath.Dir(decryptedPath) == uploadDirectoryPath {
			return encryptedPath
		}
	}

	return ""
}

func doUpload(bs *bucketService, keys *simplecrypto.Keys, uploadFile, remoteDirectory string) error {
	encryptedRemoteFilePath := ""
	remoteUploadPath := ""
	finalUploadPath := ""

	remoteUploadPath = filepath.Clean(filepath.Join(remoteDirectory, filepath.Dir(uploadFile), filepath.Base(uploadFile)))
	remoteUploadDirectoryPath := filepath.Dir(remoteUploadPath)

	if matchingDirectory := findExistingPath(*bs, keys, remoteUploadDirectoryPath); matchingDirectory != "" {
		if encryptedFilename, err := simplecrypto.EncryptText(path.Base(uploadFile), keys.EncryptionKey); err != nil {
			return err
		} else {
			encryptedRemoteFilePath = filepath.Clean(filepath.Dir(matchingDirectory) + "/" + encryptedFilename)
		}
	} else {
		finalUploadPath = remoteUploadPath
		encryptedRemoteFilePath = encryptFilePath(finalUploadPath, keys)
	}

	for e := range bs.bucketCache.seenFiles {
		plaintextFilepath, err := decryptFilePath(e, keys)
		if err != nil {
			log.Infof("Unable to decrypt filepath: %s", e)
			continue
		}
		if plaintextFilepath == remoteUploadPath {
			log.Debugf("this file already exists: %s", plaintextFilepath)
			return errors.New(fileAlreadyExistsError)
		}
	}

	log.WithFields(logrus.Fields{"filename": uploadFile}).Debug("Starting encryption of file.")
	encryptedFile, md5Hash, err := simplecrypto.EncryptFile(uploadFile, keys)
	defer os.Remove(encryptedFile)

	if err != nil {
		panic(err)
	}

	log.WithFields(logrus.Fields{"filename": uploadFile}).Debug("Encryption of file complete.")
	log.WithFields(logrus.Fields{"filename": uploadFile, "remoteDirectory": remoteDirectory}).Info("Uploading file")

	if err := bs.uploadToBucket(encryptedFile, keys, md5Hash, encryptedRemoteFilePath); err != nil {
		return err
	}

	bs.bucketCache.addFile(encryptedRemoteFilePath, remoteUploadPath)
	return nil
}

func processUpload(bs *bucketService, keys *simplecrypto.Keys, uploadPath, remoteDirectory string) error {
	defer os.Remove(uploadPath + ".enc")
	defer bs.bucketCache.empty()
	errorOccured := false

	globMatch, err := filepath.Glob(uploadPath)

	if err != nil {
		return err
	}

	objects, err := bs.getObjects()

	if err != nil {
		return err
	}

	if len(globMatch) == 0 {
		return errors.New(fileNotFoundError)
	}

	// cache the list of all files before we start uploading
	for _, encryptedFile := range objects {
		decryptedFile, _ := decryptFilePath(encryptedFile, keys)
		bs.bucketCache.addFile(encryptedFile, decryptedFile)
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
				log.Info("file already exists, skipping upload.")
			default:
				log.Infof("failed with %s when uploading: %s", err.Error(), path)
				return err
			}
		}
	}

	if errorOccured {
		return errors.New(fileUploadFailError)
	}

	return nil
}
