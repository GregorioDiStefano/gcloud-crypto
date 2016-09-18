package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	googleAPI "google.golang.org/api/googleapi"
	storage "google.golang.org/api/storage/v1"
)

type bucketService struct {
	service storage.Service
	bucket
}

type bucket struct {
	name    string
	project string
}

type BucketInteraction interface {
	uploadToBucket(string, string) *bucketService
	getObjects() ([]string, error)
}

type PassThrough struct {
	io.Reader
	totalRead     int64
	contentLength int64
}

type Progress struct {
	total      int64
	transfered int64
}

var progresses []Progress

func (pt *PassThrough) Read(b []byte) (int, error) {
	c, err := pt.Reader.Read(b)
	pt.totalRead += int64(c)

	fmt.Printf("%.2f complete.\r", 100*float64(pt.totalRead)/float64(pt.contentLength))

	return c, err
}

func NewBucketService(service storage.Service, bucketName, projectName string) *bucketService {
	return &bucketService{service, bucket{bucketName, projectName}}
}

func (bs bucketService) deleteObject(encryptedFilePath string) error {
	if err := bs.service.Objects.Delete(bs.bucket.name, encryptedFilePath).Do(); err == nil {
	} else {
		return errors.New(fmt.Sprintf("Failed to delete <%s>: %s", encryptedFilePath, err.Error()))
	}
	return nil
}

func (bs bucketService) uploadToBucket(fileToUpload, encryptedUploadPath string) error {
	var fileSize int64
	object := &storage.Object{Name: encryptedUploadPath}

	file, err := os.Open(fileToUpload)
	defer os.Remove(fileToUpload)

	if err != nil {
		return errors.New("Failed opening file: " + fileToUpload)
	}

	if fileStat, err := os.Stat(fileToUpload); err == nil {
		fileSize = fileStat.Size()
	}

	defer file.Close()

	if err != nil {
		fmt.Printf("Error opening %q: %v", fileToUpload, err)
	}

	var pu googleAPI.ProgressUpdater = func(current, total int64) {
		fmt.Printf("Uploaded: %.2f%%\r", (float64(current) / float64(fileSize) * float64(100)))
	}

	if res, err := bs.service.Objects.Insert(bs.bucket.name, object).ProgressUpdater(pu).Media(file).Do(); err == nil {
		fmt.Printf("Created object %v at location %v\n\n", res.Name, res.SelfLink)
	} else {
		return errors.New("Failed to upload")
	}

	return nil
}

func (bs bucketService) downloadFromBucket(encryptedFilePath string) (string, error) {
	writeFile, _ := ioutil.TempFile(".", "download")
	saveFilename := writeFile.Name()
	defer writeFile.Close()

	obj := bs.service.Objects.Get(bs.bucket.name, encryptedFilePath)
	download, err := obj.Download()

	if err != nil {
		return "", errors.New("Error trying to download file:" + err.Error())
	}

	defer download.Body.Close()

	if err != nil {
		fmt.Println(err)
	}

	pt := &PassThrough{Reader: download.Body, contentLength: download.ContentLength}

	if written, err := io.Copy(writeFile, pt); err != nil {
		fmt.Println(err)
	} else if written != download.ContentLength {
		return "", errors.New("Download failed, file was not entirely downloaded")
	}

	writeFile.Close()

	return saveFilename, nil
}

func (bs bucketService) getObjects() ([]string, error) {
	var objects []string
	pageToken := ""

	for {
		call := bs.service.Objects.List(bs.bucket.name)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Do()
		if err != nil {
			return nil, errors.New("failed to get objects in bucket")
		}
		for _, object := range res.Items {
			objects = append(objects, object.Name)
		}
		if pageToken = res.NextPageToken; pageToken == "" {
			break
		}
	}

	return objects, nil
}
