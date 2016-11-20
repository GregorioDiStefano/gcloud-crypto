package main

import (
	"errors"
	"fmt"

	_ "github.com/GregorioDiStefano/go-file-storage/log"
	"github.com/Sirupsen/logrus"
	"github.com/ryanuber/go-glob"
)

func (c *client) doDeleteObject(filepath string, encrypted bool) error {
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
	for plaintextFilename, _ := range decToEncPaths {

		if glob.Glob(filepath, plaintextFilename) && plaintextFilename != PASSWORD_CHECK_FILE {
			encryptedFilename := decToEncPaths[plaintextFilename]
			if encryptedFilename == "" {
				return fmt.Errorf("file: %s not found in bucket", filepath)
			}

			if err := c.bucket.Delete(encryptedFilename); err != nil {
				return err
			} else {
				log.WithFields(logrus.Fields{"filename": plaintextFilename}).Debug("deleted file.")
			}
		}
	}
	return nil
}
