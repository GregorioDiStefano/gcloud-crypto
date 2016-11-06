package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMismatch(t *testing.T) {
	bs, keys := setupUp()
	cleanUp(bs)

	getFileList(bs, &keys, "")
	file1 := encryptFilePath("test0", &keys)
	file2 := encryptFilePath("test1", &keys)

	randomFileTestFilename1 := randomFile()
	randomFileTestFilename2 := randomFile()

	md5hash, _ := getFileMD5(randomFileTestFilename1)
	err := bs.uploadToBucket(randomFileTestFilename1, &keys, md5hash, file1)

	err = bs.uploadToBucket(randomFileTestFilename2, &keys, []byte{0x00}, file2)
	assert.Equal(t, err, errors.New(hashMismatchErr))
	filesInBucket, err := getFileList(bs, &keys, "")
	assert.Equal(t, []string{"test0"}, filesInBucket)
}

func TestMoveObject(t *testing.T) {
	bs, keys := setupUp()
	cleanUp(bs)

	srcFile := encryptFilePath("test0", &keys)
	dstFile := encryptFilePath("dst", &keys)

	randomFileTestFilename := randomFile()
	md5hash, _ := getFileMD5(randomFileTestFilename)

	bs.uploadToBucket(randomFileTestFilename, &keys, md5hash, srcFile)
	bs.moveObject(srcFile, dstFile)

	files, err := getFileList(bs, &keys, "")

	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Contains(t, files, "dst")
}
