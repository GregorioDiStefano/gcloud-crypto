package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"github.com/Sirupsen/logrus"
	"github.com/ryanuber/go-glob"
)

func (bs *bucketService) doMoveObject(keys simplecrypto.Keys, src, dst string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	isGlob := strings.Contains(src, "*")
	isDestinationFolder := strings.HasSuffix(dst, "/")

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys.EncryptionKey)
	for plaintextFilename := range decToEncPaths {
		var finalDst string
		if glob.Glob(src, plaintextFilename) {
			encryptedFilename := decToEncPaths[plaintextFilename]

			if encryptedFilename == "" {
				return fmt.Errorf("file: %s not found in bucket", src)
			}

			if isGlob || isDestinationFolder {
				// copy all the files to the dst 'folder'
				finalDst = filepath.Clean(dst + "/" + plaintextFilename)
				fmt.Println(finalDst)
			}

			finalDstEncrypted := encryptFilePath(finalDst, keys.EncryptionKey)
			if err := bs.moveObject(encryptedFilename, finalDstEncrypted); err != nil {
				return err
			}
			log.WithFields(logrus.Fields{"original": plaintextFilename, "new location": finalDst}).Debug("file moved")
		}
	}
	return nil
}
