package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	"github.com/Sirupsen/logrus"
	"github.com/ryanuber/go-glob"
)

func (bs *bucketService) doMoveObject(keys *simplecrypto.Keys, src, dst string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	isGlob := strings.HasSuffix(src, "*")

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, keys)
	for plaintextFilename := range decToEncPaths {
		var finalDst string

		if glob.Glob(src, plaintextFilename) {
			encryptedFilename := decToEncPaths[plaintextFilename]

			if encryptedFilename == "" {
				return fmt.Errorf("file: %s not found in bucket", src)
			}

			// destination is always a folder when you are copy a set of files
			if isGlob {
				// copy all the files to the dst 'folder'
				srcWithoutWildcard := strings.Trim(src, "*")
				if strings.HasPrefix(plaintextFilename, filepath.Dir(srcWithoutWildcard)) {
					finalDst = filepath.Clean(filepath.Join(dst, strings.TrimPrefix(plaintextFilename, filepath.Dir(srcWithoutWildcard))))
					finalDst = strings.TrimPrefix(finalDst, "/")
				}
			} else {
				finalDst = filepath.Clean(filepath.Join(dst, plaintextFilename))
			}

			finalDstEncrypted := encryptFilePath(finalDst, keys)
			if err := bs.moveObject(encryptedFilename, finalDstEncrypted); err != nil {
				return err
			}
			log.WithFields(logrus.Fields{"original": plaintextFilename, "new location": finalDst}).Debug("file moved")
		}
	}
	return nil
}
