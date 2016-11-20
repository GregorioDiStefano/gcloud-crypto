package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMismatch(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)

	c.getFileList("")
	file1 := encryptFilePath("test0", &keys)
	file2 := encryptFilePath("test1", &keys)

	randomFileTestFilename1 := randomFile()
	randomFileTestFilename2 := randomFile()

	md5hash, _ := getFileMD5(randomFileTestFilename1)
	err := bs.Upload(randomFileTestFilename1, file1, md5hash)
	assert.Nil(t, err)
	err = bs.Upload(randomFileTestFilename2, file2, []byte{0x00})
	assert.Equal(t, err, errors.New(hashMismatchErr))
	filesInBucket, err := c.getFileList("")
	assert.Equal(t, []string{"test0"}, filesInBucket)
}

func TestMoveObject(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)

	srcFile := encryptFilePath("test0", &keys)
	dstFile := encryptFilePath("dst", &keys)

	randomFileTestFilename := randomFile()
	md5hash, _ := getFileMD5(randomFileTestFilename)

	bs.Upload(randomFileTestFilename, srcFile, md5hash)
	bs.Move(srcFile, dstFile)

	files, err := c.getFileList("")

	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Contains(t, files, "dst")
}
