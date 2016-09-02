package main

import (
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
			decryptedPath = []string{"error decrypting filepath"}
		}
	}

	entireDecryptedPath := strings.Join(decryptedPath, "/")
	return entireDecryptedPath
}

func addHMACToFile(filepath string, hmac []byte) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0600)
	defer f.Close()

	if err != nil {
		return err
	}

	f.Write(hmac)

	return nil
}

func getHMACFromFile(filepath string) ([]byte, error) {
	f, err := os.Open(filepath)
	defer f.Close()
	hmacBytes := make([]byte, 32)

	if err != nil {
		return nil, err
	}

	if fileStat, err := f.Stat(); err == nil {
		fileSize := fileStat.Size()
		if _, err := f.ReadAt(hmacBytes, fileSize-32); err == nil {
			return hmacBytes, nil
		}
		return nil, err
	} else {
		return nil, err
	}
}

func truncateHMACSignature(filepath string) error {
	if fileStat, err := os.Stat(filepath); err == nil {
		return os.Truncate(filepath, fileStat.Size()-32)
	} else {
		return err
	}
}
