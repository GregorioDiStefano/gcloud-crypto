package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoDeleteObject(t *testing.T) {
	bs, keys := setupUp()

	defer tearDown(bs)

	uploadTests := []struct {
		uploadFilepath       string
		destinationDirectory string

		deleteAfterTest   bool
		srcType           string
		expectedError     error
		expectedStructure []string
	}{
		{"testdata/testdata1", "test2", false, "file", nil, []string{"test2/testdata/testdata1"}},
	}

	for _, e := range uploadTests {
		if err := processUpload(bs, keys, e.uploadFilepath, e.destinationDirectory); err == nil {
			fmt.Println(getFileList(bs, keys.EncryptionKey))
			doDeleteObject(bs, keys, e.destinationDirectory+"/"+e.uploadFilepath)
			time.Sleep(30 * time.Second)
			fileList, err := getFileList(bs, keys.EncryptionKey)
			assert.Empty(t, fileList)
			assert.Empty(t, err)
		} else if err != nil {
			panic("failed to upload")
		}
	}
}
