package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoDeleteObject(t *testing.T) {
	bs, keys := setupUp()
	cleanUp(bs)

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
		if err := processUpload(bs, &keys, e.uploadFilepath, e.destinationDirectory); err == nil {
			bs.doDeleteObject(&keys, e.destinationDirectory+"/"+e.uploadFilepath, false)
			fileList, err := getFileList(bs, &keys, "")
			assert.Empty(t, fileList)
			assert.Empty(t, err)
		} else if err != nil {
			panic("failed to upload")
		}
	}
}
