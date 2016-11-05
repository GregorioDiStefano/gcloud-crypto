package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirsListing(t *testing.T) {
	bs, keys := setupUp()
	cleanUp(bs)

	dirListTests := []struct {
		uploadFilepath       string
		destinationDirectory string
		searchGlob           string

		deleteAfterTest bool
		expectedError   interface{}
		expectedOutput  []string
	}{
		{"", "", "*", true, nil, []string{}},
		{"testdata/*", "", "*", true, nil, []string{
			"testdata",
			"testdata/nested_1",
			"testdata/nested_1/nested_nested_1",
			"testdata/nested_1/nested_nested_1/nested_nested_nested_1",
			"testdata/nested_2",
			"testdata/nested_3",
			"testdata/test_a",
			"testdata/test_b",
		}},
		{"testdata/*", "abc", "*", true, nil, []string{
			"abc/testdata",
			"abc/testdata/nested_1",
			"abc/testdata/nested_1/nested_nested_1",
			"abc/testdata/nested_1/nested_nested_1/nested_nested_nested_1",
			"abc/testdata/nested_2",
			"abc/testdata/nested_3",
			"abc/testdata/test_a",
			"abc/testdata/test_b",
		}},
		{"testdata/", "abc", "abc/*", true, nil, []string{
			"abc/testdata",
			"abc/testdata/nested_1",
			"abc/testdata/nested_1/nested_nested_1",
			"abc/testdata/nested_1/nested_nested_1/nested_nested_nested_1",
			"abc/testdata/nested_2",
			"abc/testdata/nested_3",
			"abc/testdata/test_a",
			"abc/testdata/test_b",
		}},
	}

	for _, e := range dirListTests {
		if e.uploadFilepath != "" {
			err := processUpload(bs, &keys, e.uploadFilepath, e.destinationDirectory)
			assert.Nil(t, err)
		}

		dirsInBucket, err := getDirList(bs, &keys, e.searchGlob)
		assert.Nil(t, err)
		assert.EqualValues(t, e.expectedOutput, dirsInBucket)
		cleanUp(bs)
	}

}

func TestFileListing(t *testing.T) {
	bs, keys := setupUp()
	cleanUp(bs)

	dirListTests := []struct {
		uploadFilepath       string
		destinationDirectory string

		deleteAfterTest bool
		expectedError   interface{}
		expectedOutput  []string
	}{
		{"", "", true, nil, nil},
		{"testdata/*", "", true, nil, []string{
			"testdata/nested_1/nested_nested_1/nested_nested_nested_1/testdata1",
			"testdata/nested_1/nested_nested_1/testdata1",
			"testdata/nested_1/testdata1",
			"testdata/nested_2/testdata2",
			"testdata/nested_3/testdata1",
			"testdata/nested_3/testdata2",
			"testdata/nested_3/testdata3",
			"testdata/nested_3/testdata4",
			"testdata/test_a/a",
			"testdata/test_b/b",
			"testdata/testdata1",
			"testdata/testdata2",
			"testdata/testdata3",
			"testdata/testdata4",
			"testdata/testdata5",
			"testdata/testdata6",
		}},
	}

	for _, e := range dirListTests {
		if e.uploadFilepath != "" {
			err := processUpload(bs, &keys, e.uploadFilepath, e.destinationDirectory)
			assert.Nil(t, err)
		}

		filesInBucket, err := getFileList(bs, &keys, "")
		assert.Nil(t, err)
		assert.EqualValues(t, e.expectedOutput, filesInBucket)
	}

}
