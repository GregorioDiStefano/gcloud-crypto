package main

import (
	//"io/ioutil"

	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoDownload(t *testing.T) {
	bs, keys := setupUp()

	uploadPath := "testdata"
	processUpload(bs, &keys, uploadPath, "")

	downloadTests := []struct {
		downloadGlob                 string
		downloadDestinationDirectory string

		expectedError         error
		expectedStructureType string
	}{
		{"testdata/*", "dltest0", nil, "globdir"},
		{"testdata/", "dltest1", nil, "dir"},
		{"testdata/nested_1/nested_nested_1/nested_nested_nested_1/testdata1", "dltest2", nil, "file"},
		{"testdata/testdata1", "", nil, "file"},
		{"*", "dltest2", nil, "dir"},
		{"foo", "", errors.New(fileNotFoundRemotelyError), ""},
	}

	for _, e := range downloadTests {
		defer os.RemoveAll(e.downloadDestinationDirectory)
		err := doDownload(bs, &keys, e.downloadGlob, e.downloadDestinationDirectory)

		fmt.Println(getFileList(bs, &keys, ""))
		assert.Equal(t, e.expectedError, err)

		switch e.expectedStructureType {
		case "file":
			_, downloadFilename := filepath.Split(e.downloadGlob)
			fmt.Println("diff", "-r", "-q", e.downloadGlob, filepath.Join(e.downloadDestinationDirectory, downloadFilename))
			out, err := exec.Command("diff", "-r", "-q", e.downloadGlob, filepath.Join(e.downloadDestinationDirectory, downloadFilename)).Output()
			assert.Empty(t, out)
			assert.Nil(t, err)
		case "dir":
			out, err := exec.Command("diff", "-r", "-q", uploadPath, filepath.Join(e.downloadDestinationDirectory, uploadPath)).Output()
			assert.Empty(t, out)
			assert.Nil(t, err)
		case "globdir":
			out, err := exec.Command("diff", "-r", "-q", uploadPath, filepath.Join(e.downloadDestinationDirectory)).Output()
			assert.Empty(t, out)
			assert.Nil(t, err)
		}
	}
}
