package main

import (
	"errors"
	"os"
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
func reuseExistingEncryptedPath(bs bucketService, keys *simplecrypto.Keys, fullPlaintextRemoteUploadPath string) string {
	for encryptedPath, decryptedPath := range bs.bucketCache.seenFiles {
		if filepath.Dir(decryptedPath) == filepath.Dir(fullPlaintextRemoteUploadPath) {
			if encryptedFilename, err := simplecrypto.EncryptText(path.Base(fullPlaintextRemoteUploadPath), keys.EncryptionKey); err != nil {
				return ""
			} else {
				return filepath.Clean(filepath.Dir(encryptedPath) + "/" + encryptedFilename)
			}
		}
	}
	return ""
}

func globMatchWithDirectories(path string) []string {
	globMatch, _ := filepath.Glob(path)
	matches := []string{}

	for _, matchedPath := range globMatch {
		filepath.Walk(matchedPath, func(matchedPath string, info os.FileInfo, err error) error {
			if !isDir(matchedPath) {
				matches = append(matches, matchedPath)
			}
			return nil
		})
	}
	return matches
}

func prepareAndDoUpload(bs *bucketService, uploadFile, remoteUploadPath string, keys *simplecrypto.Keys) error {
	finalEncryptedUploadPath := ""

	for e := range bs.bucketCache.seenFiles {
		plaintextFilepath, err := decryptFilePath(e, keys)
		if err != nil {
			log.Errorf("Unable to decrypt filepath: %s", e)
			continue
		}
		if plaintextFilepath == remoteUploadPath {
			log.Infof("this file already exists: %s", plaintextFilepath)
			return errors.New(fileAlreadyExistsError)
		}
	}

	encryptedFile, md5Hash, err := simplecrypto.EncryptFile(uploadFile, keys)

	if err != nil {
		log.Error(err)
	}

	if finalEncryptedUploadPath = reuseExistingEncryptedPath(*bs, keys, remoteUploadPath); finalEncryptedUploadPath == "" {
		finalEncryptedUploadPath = encryptFilePath(remoteUploadPath, keys)
	}

	if err := bs.uploadToBucket(encryptedFile, keys, md5Hash, finalEncryptedUploadPath); err != nil {
		return err
	}

	bs.bucketCache.addFile(finalEncryptedUploadPath, remoteUploadPath)
	return nil
}

func processUpload(bs *bucketService, keys *simplecrypto.Keys, uploadPath, remoteDirectory string) error {
	globMatches := globMatchWithDirectories(uploadPath)
	errorOccuredWhileUploading := false

	if len(globMatches) == 0 {
		return errors.New(fileNotFoundError)
	}

	// cache the list of all files before we start uploading
	objects, err := bs.getObjects()

	if err != nil {
		log.Fatal("Unable to load remote objects")
		return err
	}

	for _, encryptedFile := range objects {
		decryptedFile, _ := decryptFilePath(encryptedFile, keys)
		bs.bucketCache.addFile(encryptedFile, decryptedFile)
	}

	for _, fileToUpload := range globMatches {
		newUploadDirectory := ""
		if strings.Contains(uploadPath, "*") {
			newUploadDirectory = filepath.Join(remoteDirectory, relativePathFromGlob(uploadPath, fileToUpload))
		} else {
			newUploadDirectory = filepath.Join(remoteDirectory, fileToUpload)
		}
		if err := prepareAndDoUpload(bs, fileToUpload, newUploadDirectory, keys); err != nil {
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
