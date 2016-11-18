package main

import (
	"errors"
	_ "fmt"
	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	"github.com/ryanuber/go-glob"
	_ "google.golang.org/appengine/log"
	"os"
	"path"
	"path/filepath"
	"strings"
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

func doDownload(bs *bucketService, keys *simplecrypto.Keys, downloadPath, destinationDir string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	if len(destinationDir) > 0 {
		if _, err := os.Stat(destinationDir); os.IsNotExist(err) {
			log.Infof("destination directory: %s does not exist, created.", destinationDir)
			os.Mkdir(destinationDir, 0777)
		}
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys)
	foundFile := false

	for remotePlaintextPath := range decToEncPaths {
		globMatched := glob.Glob(downloadPath, remotePlaintextPath)
		isDirectoryDownload := strings.HasSuffix(downloadPath, "/") && strings.HasPrefix(remotePlaintextPath, downloadPath)
		if globMatched || isDirectoryDownload {
			foundFile = true
			finalDownloadDestination := ""
			log.Infof("Downloading: %s", remotePlaintextPath)

			encryptedFilepath := decToEncPaths[remotePlaintextPath]
			decryptedFilePath, _ := decryptFilePath(decToEncPaths[remotePlaintextPath], keys)
			downloadedEncryptedFile, err := bs.downloadFromBucket(encryptedFilepath)
			defer os.Remove(downloadedEncryptedFile)

			if err != nil {
				return err
			}

			downloadedPlaintextFile, err := simplecrypto.DecryptFile(downloadedEncryptedFile, keys)

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
