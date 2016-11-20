package main

import (
	"errors"
	"path"
	"path/filepath"
	"strings"

	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	_ "github.com/GregorioDiStefano/go-file-storage/log"
)

const (
	fileNotFoundError      = "That file did not exist locally"
	fileAlreadyExistsError = "This file already exists at this location"
	fileUploadFailError    = "at least one file failed to upload"
)

// reuseExistingEncryptedPath is an optimization: reuse already existing encrypted paths instead of
// having the same encrypted path in different objects
func (c *client) reuseExistingEncryptedPath(fullPlaintextRemoteUploadPath string) string {
	for encryptedPath, decryptedPath := range c.bcache.seenFiles {
		if filepath.Dir(decryptedPath) == filepath.Dir(fullPlaintextRemoteUploadPath) {
			if encryptedFilename, err := simplecrypto.EncryptText(path.Base(fullPlaintextRemoteUploadPath), c.keys.EncryptionKey); err != nil {
				return ""
			} else {
				return filepath.Clean(filepath.Dir(encryptedPath) + "/" + encryptedFilename)
			}
		}
	}
	return ""
}

func (c *client) prepareAndDoUpload(uploadFile, remoteUploadPath string) error {
	finalEncryptedUploadPath := ""

	for e := range c.bcache.seenFiles {
		plaintextFilepath, err := decryptFilePath(e, c.keys)
		if err != nil {
			log.Errorf("Unable to decrypt filepath: %s", e)
			continue
		}
		if plaintextFilepath == remoteUploadPath {
			log.Infof("this file already exists: %s", plaintextFilepath)
			return errors.New(fileAlreadyExistsError)
		}
	}

	encryptedFile, md5Hash, err := simplecrypto.EncryptFile(uploadFile, c.keys)

	if err != nil {
		return err
	}

	if finalEncryptedUploadPath = c.reuseExistingEncryptedPath(remoteUploadPath); finalEncryptedUploadPath == "" {
		finalEncryptedUploadPath = encryptFilePath(remoteUploadPath, c.keys)
	}

	if err := c.bucket.Upload(encryptedFile, finalEncryptedUploadPath, md5Hash); err != nil {
		return err
	}

	c.bcache.addFile(finalEncryptedUploadPath, remoteUploadPath)
	return nil
}

func (c *client) processUpload(uploadPath, remoteDirectory string) error {
	globMatches := globMatchWithDirectories(uploadPath)
	errorOccuredWhileUploading := false

	if len(globMatches) == 0 {
		return errors.New(fileNotFoundError)
	}

	// cache the list of all files before we start uploading
	objects, err := c.bucket.List()

	if err != nil {
		log.Fatal("Unable to load remote objects")
		return err
	}

	for _, encryptedFile := range objects {
		decryptedFile, _ := decryptFilePath(encryptedFile, c.keys)
		c.bcache.addFile(encryptedFile, decryptedFile)
	}

	for _, fileToUpload := range globMatches {
		newUploadDirectory := ""
		if strings.Contains(uploadPath, "*") {
			newUploadDirectory = filepath.Join(remoteDirectory, relativePathFromGlob(uploadPath, fileToUpload))
		} else {
			newUploadDirectory = filepath.Join(remoteDirectory, fileToUpload)
		}
		if err := c.prepareAndDoUpload(fileToUpload, newUploadDirectory); err != nil {
			if err != nil {
				errorOccuredWhileUploading = true
				switch err.Error() {
				case fileAlreadyExistsError:
					log.Info("file already exists, skipping upload.")
				default:
					log.Infof("failed with %s when uploading: %s", err.Error(), fileToUpload)
					return err
				}
			}
		}
	}

	if errorOccuredWhileUploading {
		return errors.New(fileUploadFailError)
	}

	return nil
}
