package main

import (
	"errors"
	"testing"

	_ "github.com/GregorioDiStefano/go-file-storage/log"
	"github.com/stretchr/testify/assert"
)

func TestDoDeleteObject(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}

	uploadTests := []struct {
		uploadFilepath string
		deletePath     string

		expectedError                error
		expectedStructureAfterDelete []string
	}{
		{"testdata/testdata1", "testdata/testdata1", nil, nil},
		{"foo", "bar", errors.New(errDeleteFileNotFound), nil},
		{"testdata/testdata1", "foo", errors.New(errDeleteFileNotFound), []string{"testdata/testdata1"}},
		{"testdata/", "testdata/*", nil, nil},
		{"testdata/*", "testdata*", nil, []string{"nested_1/nested_nested_1/nested_nested_nested_1/testdata1",
			"nested_1/nested_nested_1/testdata1",
			"nested_1/testdata1",
			"nested_2/testdata2",
			"nested_3/testdata1",
			"nested_3/testdata2",
			"nested_3/testdata3",
			"nested_3/testdata4",
			"test_a/a",
			"test_b/b"}},
	}

	for _, e := range uploadTests {
		cleanUp(c)
		if err := c.processUpload(e.uploadFilepath, ""); err == nil {
			err := c.doDeleteObject(e.deletePath, false)
			assert.Equal(t, err, e.expectedError)
			fileList, _ := c.getFileList("")
			assert.EqualValues(t, e.expectedStructureAfterDelete, fileList)
		} else if err != nil {
			log.Error("failed to upload: " + e.uploadFilepath)
		}
	}
}
