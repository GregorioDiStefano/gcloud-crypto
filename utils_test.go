package main

import (
	"crypto/aes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randomByte(length int) []byte {
	b := make([]byte, length)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}

	return b
}

func TestEncryptDecryptFilePath_1(t *testing.T) {
	t.Parallel()

	fakePath := "root/a/abc/def/a.txt"
	key := randomByte(aes.BlockSize)

	encryptedPath := encryptFilePath(fakePath, key)
	decryptedPath := decryptFilePath(encryptedPath, key)

	assert.Equal(t, decryptedPath, fakePath)
}

func TestEncryptDecryptFilePath_2(t *testing.T) {
	t.Parallel()

	fakePath := "abc"
	key := randomByte(aes.BlockSize)

	encryptedPath := encryptFilePath(fakePath, key)
	decryptedPath := decryptFilePath(encryptedPath, key)

	assert.Equal(t, decryptedPath, fakePath)
}

func TestEncryptDecryptFilePath_3(t *testing.T) {
	t.Parallel()

	fakePath := "/abc"
	key := randomByte(aes.BlockSize)

	encryptedPath := encryptFilePath(fakePath, key)
	decryptedPath := decryptFilePath(encryptedPath, key)

	assert.Equal(t, decryptedPath, fakePath)
}
