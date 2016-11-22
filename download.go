package main

import (
	"errors"
	_ "fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	"github.com/ryanuber/go-glob"
)

const (
	fileNotFoundRemotelyError  = "File not found"
	destinationFileExistsError = "File already exists locally"
)

func moveDownload(source, destination string) error {
	log.Debugf("downloaded file to: %s", destination)
	if destStat, err := os.Stat(destination); err == nil {
		// destination file exists
		sourceStat, _ := os.Stat(source)
		if sourceStat.Size() == destStat.Size() {
			log.Error("downloaded file already exists in destination")
			return errors.New(destinationFileExistsError)
		}
	}

	log.Debugf("moving file from %s to %s", source, destination)
	os.MkdirAll(filepath.Dir(destination), 0777)
	return os.Rename(source, destination)
}

func (c *client) doDownload(downloadPath, destinationDir string) error {
	objects, err := c.bucket.List()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	if len(destinationDir) > 0 {
		if _, err := os.Stat(destinationDir); os.IsNotExist(err) {
			log.Infof("destination directory: %s does not exist, created.", destinationDir)
			os.Mkdir(destinationDir, 0777)
		}
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, c.keys)
	foundFile := false

	for remotePlaintextPath := range decToEncPaths {
		globMatched := glob.Glob(downloadPath, remotePlaintextPath)
		isDirectoryDownload := strings.HasSuffix(downloadPath, "/") && strings.HasPrefix(remotePlaintextPath, downloadPath)
		if globMatched || isDirectoryDownload {
			foundFile = true
			finalDownloadDestination := ""
			log.Infof("Downloading: %s", remotePlaintextPath)

			encryptedFilepath := decToEncPaths[remotePlaintextPath]
			decryptedFilePath, _ := decryptFilePath(decToEncPaths[remotePlaintextPath], c.keys)
			downloadedEncryptedFile, err := c.bucket.Download(encryptedFilepath)
			defer os.Remove(downloadedEncryptedFile)

			if err != nil {
				return err
			}

			downloadedPlaintextFile, err := simplecrypto.DecryptFile(downloadedEncryptedFile, c.keys)

			if err != nil {
				return err
			}

			_, tempDownloadFilename := path.Split(downloadedPlaintextFile)
			_, actualFilename := path.Split(decryptedFilePath)

			if isDirectoryDownload {
				finalDownloadDestination = filepath.Join(destinationDir, filepath.Dir(remotePlaintextPath), actualFilename)
			} else if downloadPath == remotePlaintextPath {
				// downloading a single file
				finalDownloadDestination = filepath.Join(destinationDir, actualFilename)
			} else {
				// downloading a matching glob
				relativeDownloadPath := relativePathFromGlob(downloadPath, remotePlaintextPath)
				finalDownloadDestination = filepath.Join(destinationDir, relativeDownloadPath)
			}

			//TODO: check if filesize matches
			if _, err := os.Stat(finalDownloadDestination); err == nil {
				log.Errorf("file already exists: %s exists locally, skipping", finalDownloadDestination)
				os.Remove(downloadedPlaintextFile)
				continue
			}

			moveDownload(tempDownloadFilename, finalDownloadDestination)
			os.Remove(downloadedEncryptedFile)
		}
	}
	if !foundFile {
		return errors.New(fileNotFoundRemotelyError)
	}
	return nil
}
