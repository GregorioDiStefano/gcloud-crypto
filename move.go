package main

import (
	"errors"
	"path/filepath"
	"strings"

	_ "github.com/GregorioDiStefano/go-file-storage/log"
	"github.com/Sirupsen/logrus"
	"github.com/ryanuber/go-glob"
)

func (c *client) doMoveObject(src, dst string) error {
	objects, err := c.bucket.List()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	if len(objects) == 0 {
		return errors.New("no objects exist remotely, nothing to move")
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, c.keys)
	isGlob := strings.Contains(src, "*")

	for plaintextFilename := range decToEncPaths {
		var finalDst string

		// this is a single file rename
		if !strings.HasSuffix(src, "/") && !strings.HasSuffix(dst, "/") {
			encryptedFilename := decToEncPaths[plaintextFilename]
			finalDstEncrypted := encryptFilePath(dst, c.keys)

			log.WithFields(logrus.Fields{"original": plaintextFilename, "new location": finalDst}).Debug("file moved")
			return c.bucket.Move(encryptedFilename, finalDstEncrypted)
		}

		// this is a directory rename
		if strings.HasSuffix(src, "/") && strings.HasSuffix(dst, "/") && strings.HasPrefix(plaintextFilename, src) {
			encryptedFilename := decToEncPaths[plaintextFilename]
			finalDstEncrypted := encryptFilePath(filepath.Clean(filepath.Join(dst, plaintextFilename)), c.keys)
			if err := c.bucket.Move(encryptedFilename, finalDstEncrypted); err != nil {
				return err
			}
			continue
		}

		if isGlob && glob.Glob(src, plaintextFilename) {
			encryptedFilename := decToEncPaths[plaintextFilename]

			// copy all the files to the dst 'folder'
			srcWithoutWildcard := strings.Trim(src, "*")
			if strings.HasPrefix(plaintextFilename, filepath.Dir(srcWithoutWildcard)) {
				finalDst = filepath.Clean(filepath.Join(dst, relativePathFromGlob(src, plaintextFilename)))
				finalDst = strings.TrimPrefix(finalDst, "/")
			} else {
				finalDst = filepath.Clean(filepath.Join(dst, plaintextFilename))
			}

			finalDstEncrypted := encryptFilePath(finalDst, c.keys)
			if err := c.bucket.Move(encryptedFilename, finalDstEncrypted); err != nil {
				return err
			}
			log.WithFields(logrus.Fields{"original": plaintextFilename, "new location": finalDst}).Debug("file moved")
		}
	}

	return nil
}
