package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashMismatch(t *testing.T) {
	bs, keys := setupUp()
	defer tearDown(bs)

	fmt.Println(getFileList(bs, keys.EncryptionKey, ""))
	file1 := encryptFilePath("test0", keys.EncryptionKey)
	file2 := encryptFilePath("test1", keys.EncryptionKey)

	randomFileTestFilename1 := randomFile()
	randomFileTestFilename2 := randomFile()

	md5hash, _ := getFileMD5(randomFileTestFilename1)
	err := bs.uploadToBucket(randomFileTestFilename1, keys, md5hash, file1)
	fmt.Print("files in bucket : ")
	fmt.Println(getFileList(bs, keys.EncryptionKey, ""))

	err = bs.uploadToBucket(randomFileTestFilename2, keys, []byte{0x00}, file2)
	assert.Equal(t, err, errors.New(hashMismatchErr))
	filesInBucket, err := getFileList(bs, keys.EncryptionKey, "")
	assert.Equal(t, []string{"test0"}, filesInBucket)
}
