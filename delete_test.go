package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoDeleteObject(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)

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
		if err := c.processUpload(e.uploadFilepath, e.destinationDirectory); err == nil {
			c.doDeleteObject(e.destinationDirectory+"/"+e.uploadFilepath, false)
			fileList, err := c.getFileList("")
			assert.Empty(t, fileList)
			assert.Empty(t, err)
		} else if err != nil {
			panic("failed to upload")
		}
	}
}
