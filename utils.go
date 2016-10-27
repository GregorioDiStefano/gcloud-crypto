package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
)

type decryptedToEncryptedFilePath map[string]string

func isStringInSlice(s string, list []string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

func getDecryptedToEncryptedFileMapping(encryptedFilePaths []string, key *simplecrypto.Keys) decryptedToEncryptedFilePath {
	m := make(decryptedToEncryptedFilePath, len(encryptedFilePaths))
	for _, e := range encryptedFilePaths {
		plainTextFilepath, err := decryptFilePath(e, key)

		if err != nil {
			fmt.Println(err)
		}

		m[plainTextFilepath] = e
	}

	return m
}

func encryptFilePath(path string, key *simplecrypto.Keys) string {
	splitPath := strings.Split(path, "/")
	var encryptedPath []string

	for _, e := range splitPath {
		encText, _ := simplecrypto.EncryptText(e, key.EncryptionKey)
		encryptedPath = append(encryptedPath, encText)
	}

	entireEncryptedPath := strings.Join(encryptedPath, "/")
	return entireEncryptedPath
}

func decryptFilePath(encryptedPath string, key *simplecrypto.Keys) (string, error) {
	splitPath := strings.Split(encryptedPath, "/")
	decryptedPath := []string{}

	for _, e := range splitPath {
		if e == PASSWORD_CHECK_FILE {
			continue
		} else if t, err := simplecrypto.DecryptText(e, key.EncryptionKey); err == nil {
			decryptedPath = append(decryptedPath, t)
		} else {
			return "", errors.New("failed to decrypt file: " + encryptedPath)
		}
	}

	entireDecryptedPath := strings.Join(decryptedPath, "/")
	return entireDecryptedPath, nil
}

func enumeratePrint(items []string) {
	if len(items) > 0 {
		count := 0
		for _, e := range items {
			if e != "" {
				fmt.Println(fmt.Sprintf("%d:\t%s", count, e))
				count++
			}
		}
	}
}

func isDir(filepath string) bool {
	if fileStat, err := os.Stat(filepath); err == nil {
		return fileStat.IsDir()
	}
	return false
}

func getFileMD5(filePath string) ([]byte, error) {
	var result []byte

	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}

	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

/*
func findCommonPath(path1, path2 string) string {
	path1 = filepath.Clean(path1)
	path2 = filepath.Clean(path2)

	splitPath1 := strings.Split(path1, "/")
	splitPath2 := strings.Split(path2, "/")
	commonPath := ""

	for i, e := range splitPath1 {
		if splitPath2[i] == e {
			if len(e) > 0 {
				commonPath += e + "/"
			}

		} else {
			break
		}
	}
	commonPath = filepath.Clean(commonPath)
	if commonPath == "." {
		return ""
	}
	return filepath.Clean(commonPath) + "/"
}
*/
