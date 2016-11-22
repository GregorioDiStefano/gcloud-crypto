package main

import (
	"errors"

	_ "github.com/GregorioDiStefano/go-file-storage/log"
	"github.com/Sirupsen/logrus"
	"github.com/ryanuber/go-glob"
)

const (
	errDeleteFileNotFound = "Delete file not found"
)

func (c *client) doDeleteObject(filepath string, encrypted bool) error {
	fileFound := false
	objects, err := c.bucket.List()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	if encrypted {
		filepath, err = decryptFilePath(filepath, c.keys)
	}

	if err != nil {
		return err
	}

	switch filepath {
	case "*", "/*", "*/*":
		return errors.New("not perform destructive delete")
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, c.keys)
	for plaintextFilename := range decToEncPaths {
		if glob.Glob(filepath, plaintextFilename) && plaintextFilename != PASSWORD_CHECK_FILE {
			fileFound = true
			encryptedFilename := decToEncPaths[plaintextFilename]
			if encryptedFilename == "" {
				return errors.New(errDeleteFileNotFound)
			}

			if err := c.bucket.Delete(encryptedFilename); err != nil {
				return err
			}
			log.WithFields(logrus.Fields{"filename": plaintextFilename}).Debug("deleted file.")
		}
	}

	if !fileFound {
		return errors.New(errDeleteFileNotFound)
	}
	return nil
}
