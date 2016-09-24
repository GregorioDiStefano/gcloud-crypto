package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
)

type decryptedToEncryptedFilePath map[string]string

func getEncryptedToDecryptedMap(encryptedFilePaths []string, key []byte) decryptedToEncryptedFilePath {
	m := make(decryptedToEncryptedFilePath, len(encryptedFilePaths))
	for _, e := range encryptedFilePaths {
		plainTextFilepath := decryptFilePath(e, key)
		m[plainTextFilepath] = e
	}

	return m
}

func encryptFilePath(path string, key []byte) string {
	splitPath := strings.Split(path, "/")
	var encryptedPath []string

	for _, e := range splitPath {
		encText, _ := simplecrypto.EncryptText(e, key)
		encryptedPath = append(encryptedPath, encText)
	}

	entireEncryptedPath := strings.Join(encryptedPath, "/")
	return entireEncryptedPath
}

func decryptFilePath(encryptedPath string, key []byte) string {
	splitPath := strings.Split(encryptedPath, "/")
	var decryptedPath []string

	for _, e := range splitPath {
		if t, err := simplecrypto.DecryptText(e, key); err == nil {
			decryptedPath = append(decryptedPath, t)
		} else {
			decryptedPath = []string{fmt.Sprintf("(error decrypting filepath) %s", encryptedPath)}
		}
	}

	entireDecryptedPath := strings.Join(decryptedPath, "/")
	return entireDecryptedPath
}

func isDir(filepath string) bool {
	if fileStat, err := os.Stat(filepath); err == nil {
		return fileStat.IsDir()
	}
	return false
}
