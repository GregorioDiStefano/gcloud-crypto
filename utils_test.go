package main

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
)

func randomByte(length int) []byte {
	b := make([]byte, length)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}

	return b
}

func TestEncryptDecryptFilePath(t *testing.T) {
	t.Parallel()

	pathTest := []struct {
		path string
	}{
		{"root/a/abc/def/a.txt"},
		{"abc"},
		{"/abc"},
	}

	keys, err := simplecrypto.GetKeyFromPassphrase([]byte("testing"), []byte("salt1234"), 4096, 16, 1)
	assert.Nil(t, err)	
	
	for _, e := range pathTest {
			encryptedPath := encryptFilePath(e.path, keys)
			decryptedPath, err := decryptFilePath(encryptedPath, keys)

			assert.Nil(t, err)
			assert.Equal(t, decryptedPath, e.path)
	}	
}