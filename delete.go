package main

import (
	"errors"
	"fmt"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"github.com/Sirupsen/logrus"
	"github.com/ryanuber/go-glob"
)

func (bs *bucketService) doDeleteObject(keys *simplecrypto.Keys, filepath string, encrypted bool) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	if encrypted {
		filepath, err = decryptFilePath(filepath, keys)
	}

	if err != nil {
		return err
	}

	switch filepath {
	case "*", "/*", "*/*":
		return errors.New("not perform destructive delete")
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys)
	for plaintextFilename, _ := range decToEncPaths {

		if glob.Glob(filepath, plaintextFilename) && plaintextFilename != PASSWORD_CHECK_FILE {
			encryptedFilename := decToEncPaths[plaintextFilename]
			if encryptedFilename == "" {
				return fmt.Errorf("file: %s not found in bucket", filepath)
			}

			if err := bs.deleteObject(encryptedFilename); err != nil {
				return err
			} else {
				log.WithFields(logrus.Fields{"filename": plaintextFilename}).Debug("deleted file.")
			}
		}
	}
	return nil
}
