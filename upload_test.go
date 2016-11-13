package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoUpload(t *testing.T) {
	bs, keys := setupUp()
	cleanUp(bs)

	randomFileTestFilename := randomFile()
	defer os.Remove(randomFileTestFilename)

	uploadTests := []struct {
		uploadFilepath       string
		destinationDirectory string

		deleteAfterTest   bool
		srcType           string
		expectedError     interface{}
		expectedStructure []string
	}{
		{"testdata/*", "test0", true, "globdir", nil, []string{
			"test0/nested_1/nested_nested_1/nested_nested_nested_1/testdata1",
			"test0/nested_1/nested_nested_1/testdata1",
			"test0/nested_1/testdata1",
			"test0/nested_2/testdata2",
			"test0/nested_3/testdata1",
			"test0/nested_3/testdata2",
			"test0/nested_3/testdata3",
			"test0/nested_3/testdata4",
			"test0/test_a/a",
			"test0/test_b/b",
			"test0/testdata1",
			"test0/testdata2",
			"test0/testdata3",
			"test0/testdata4",
			"test0/testdata5",
			"test0/testdata6"}},
		{"testdata/", "test1", true, "dir", nil, []string{
			"test1/testdata/nested_1/nested_nested_1/nested_nested_nested_1/testdata1",
			"test1/testdata/nested_1/nested_nested_1/testdata1",
			"test1/testdata/nested_1/testdata1",
			"test1/testdata/nested_2/testdata2",
			"test1/testdata/nested_3/testdata1",
			"test1/testdata/nested_3/testdata2",
			"test1/testdata/nested_3/testdata3",
			"test1/testdata/nested_3/testdata4",
			"test1/testdata/test_a/a",
			"test1/testdata/test_b/b",
			"test1/testdata/testdata1",
			"test1/testdata/testdata2",
			"test1/testdata/testdata3",
			"test1/testdata/testdata4",
			"test1/testdata/testdata5",
			"test1/testdata/testdata6"}},
		{"testdata/testdata1", "test2", false, "file", nil, []string{"test2/testdata/testdata1"}},
		{"testdata/file_that_doesnt_exist", "test3", true, "file", errors.New(fileNotFoundError), nil},
		{"testdata/nested_3/", "test3", true, "dir", nil, []string{
			"test3/testdata/nested_3/testdata1",
			"test3/testdata/nested_3/testdata2",
			"test3/testdata/nested_3/testdata3",
			"test3/testdata/nested_3/testdata4"}},
		{"testdata/nested_3/*1*", "test4", true, "", nil, []string{"test4/testdata1"}},
		{"testdata/test_*", "test5", true, "", nil, []string{"test5/test_a/a", "test5/test_b/b"}},
		{randomFileTestFilename, "", true, "", nil, []string{randomFileTestFilename}},
	}

	for c, e := range uploadTests {
		log.Info("Running #", c)

		path := e.uploadFilepath
		remoteDirectory := e.destinationDirectory

		err := processUpload(bs, &keys, path, remoteDirectory)

		if err != nil {
			log.Debug("Error uploading: ", err)
		}

		assert.Equal(t, err, e.expectedError)

		if e.expectedError == nil {
			filesInBucket, err := getFileList(bs, &keys, "")
			assert.Nil(t, err)
			assert.EqualValues(t, filesInBucket, e.expectedStructure)
		}

		if e.expectedStructure != nil {
			cwd, _ := os.Getwd()
			tempDir, _ := ioutil.TempDir(cwd, "testrun")
			err := doDownload(bs, &keys, "*", tempDir)
			assert.Nil(t, err)
			switch e.srcType {
			case "file":
				out, err := exec.Command("diff", "-r", "-q", e.uploadFilepath, tempDir+"/"+e.destinationDirectory+"/"+filepath.Dir(e.uploadFilepath)).Output()
				assert.Empty(t, out)
				assert.Nil(t, err)
			case "dir":
				out, err := exec.Command("diff", "-r", "-q", filepath.Dir(e.uploadFilepath), tempDir+"/"+e.destinationDirectory+"/"+filepath.Dir(e.uploadFilepath)).Output()
				assert.Empty(t, out)
				assert.Nil(t, err)
			case "globdir":
				out, err := exec.Command("diff", "-r", "-q", filepath.Join(tempDir, e.destinationDirectory), filepath.Dir(e.uploadFilepath)).Output()
				assert.Empty(t, out)
				assert.Nil(t, err)
			}
			os.RemoveAll(tempDir)
		}

		if e.deleteAfterTest {
			cleanUp(bs)
			objectsAfterDelete, _ := bs.getObjects()
			assert.Empty(t, objectsAfterDelete, "Looks like objects still exist after deleting them all")
		}

		bs.bucketCache.seenFiles = make(map[string]string, 100)
	}
}

func TestDoUploadResume(t *testing.T) {
	bs, keys := setupUp()
	defer cleanUp(bs)

	err := processUpload(bs, &keys, "testdata/testdata1", "")
	filesInBucket, err := getFileList(bs, &keys, "")
	assert.Nil(t, err)
	assert.Equal(t, []string{"testdata/testdata1"}, filesInBucket)

	// how to actually check the file was not reuploaded?
	err = processUpload(bs, &keys, "testdata/testdata*", "testdata/")
	assert.Equal(t, err.Error(), fileUploadFailError)

	filesInBucket, err = getFileList(bs, &keys, "")

	assert.Nil(t, err)
	assert.Equal(t, []string{
		"testdata/testdata1",
		"testdata/testdata2",
		"testdata/testdata3",
		"testdata/testdata4",
		"testdata/testdata5",
		"testdata/testdata6"}, filesInBucket)
}

func TestDoUploadDirectoryAndResume(t *testing.T) {
	bs, keys := setupUp()
	defer cleanUp(bs)

	expectedOutput := []string{
		"testdata/testdata1",
		"testdata/testdata2",
		"testdata/testdata3",
		"testdata/testdata4",
		"testdata/testdata5",
		"testdata/testdata6",
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
	}

	sort.Strings(expectedOutput)

	err := processUpload(bs, &keys, "testdata/testdata1", "")
	assert.Nil(t, err)

	// how to actually check the file was not reuploaded?
	err = processUpload(bs, &keys, "testdata", "")

	filesInBucket, err := getFileList(bs, &keys, "")

	sort.Strings(filesInBucket)
	assert.EqualValues(t, expectedOutput, filesInBucket)
}

func TestExistingDirectoriesReused(t *testing.T) {
	bs, keys := setupUp()
	defer cleanUp(bs)

	identicalRemoteDirectories := []string{}
	identicalRemoteEncryptedDirectories := []string{}

	processUpload(bs, &keys, "testdata", "testing-directories")

	filesInBucket, err := getFileList(bs, &keys, "")
	for _, e := range filesInBucket {
		if !searchForString(identicalRemoteDirectories, filepath.Dir(e)) {
			identicalRemoteDirectories = append(identicalRemoteDirectories, filepath.Dir(e))
		}
	}

	encryptedFilesInBucket, err := bs.getObjects()
	assert.Empty(t, err)

	for _, e := range encryptedFilesInBucket {
		if !searchForString(identicalRemoteEncryptedDirectories, filepath.Dir(e)) {
			identicalRemoteEncryptedDirectories = append(identicalRemoteEncryptedDirectories, filepath.Dir(e))
		}
	}

	assert.Equal(t, len(identicalRemoteEncryptedDirectories), len(identicalRemoteDirectories))
}
