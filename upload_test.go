package main

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	_ "strings"
	"testing"
	"time"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
)

func init() {
	log.Level = logrus.DebugLevel
}

func setupUp() (*bucketService, simplecrypto.Keys) {
	client, err := google.DefaultClient(context.Background(), storage.DevstorageFullControlScope)

	if err != nil {
		log.Fatalf("Unable to get default client: %v", err)
	}
	service, err := storage.New(client)

	if err != nil {
		panic(err)
	}

	userData := parseConfig()
	userData.configFile.Set("bucket", "go-testing")
	userData.configFile.Set("project_id", "stuff-141918")

	bs := NewBucketService(*service, "go-testing", "stuff-141918")
	keys, err := simplecrypto.GetKeyFromPassphrase([]byte("testing"), []byte("salt1234"), 4096, 16, 1)

	if err != nil {
		panic(err)
	}

	return bs, *keys
}

func tearDown(bs *bucketService) {
	objs, _ := bs.getObjects()
	for _, e := range objs {
		bs.deleteObject(e)
	}

	// changes are sometimes not reflected
	time.Sleep(3 * time.Second)
}

func randomFile() string {
	tmpfile, _ := ioutil.TempFile(".", "test")
	tmpfile.Write([]byte("this is a test string"))
	return tmpfile.Name()
}

func TestDoUpload(t *testing.T) {
	bs, keys := setupUp()

	defer tearDown(bs)

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
		{"testdata/*", "test1", true, "dir", nil, []string{
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
		{"testdata/nested_3/*1*", "test4", true, "", nil, []string{"test4/testdata/nested_3/testdata1"}},
		{"testdata/test_*", "test5", true, "", nil, []string{"test5/testdata/test_a/a", "test5/testdata/test_b/b"}},
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

		time.Sleep(3 * time.Second)

		if e.expectedError == nil {
			filesInBucket, err := getFileList(bs, &keys, "")
			assert.Nil(t, err)
			assert.Equal(t, filesInBucket, e.expectedStructure)
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
			case "glob":

			}
			os.RemoveAll(tempDir)
		}

		if e.deleteAfterTest {
			objs, _ := bs.getObjects()
			for _, e := range objs {
				bs.deleteObject(e)
			}
		}

	}
}

func TestDoUploadResume(t *testing.T) {
	bs, keys := setupUp()

	defer tearDown(bs)

	err := processUpload(bs, &keys, "testdata/testdata1", "")
	assert.Nil(t, err)

	// how to actually check the file was not reuploaded?
	err = processUpload(bs, &keys, "testdata/testdata*", "")
	assert.Equal(t, err.Error(), fileUploadFailError)

	filesInBucket, err := getFileList(bs, &keys, "")
	assert.Equal(t, []string{
		"/testdata/testdata1",
		"/testdata/testdata2",
		"/testdata/testdata3",
		"/testdata/testdata4",
		"/testdata/testdata5",
		"/testdata/testdata6"}, filesInBucket)
}
