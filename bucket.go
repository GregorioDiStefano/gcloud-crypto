package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

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

func NewBucketService(service storage.Service, bucketName, projectName string) *bucketService {
	return &bucketService{service, bucket{bucketName, projectName}}
}

func (bs bucketService) uploadToBucket(fileToUpload, encryptedUploadPath string) error {
	object := &storage.Object{Name: encryptedUploadPath}
	file, err := os.Open(fileToUpload)
	defer file.Close()

	if err != nil {
		fmt.Printf("Error opening %q: %v", fileToUpload, err)
	}
	if res, err := bs.service.Objects.Insert(bs.bucket.name, object).Media(file).Do(); err == nil {
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
	defer download.Body.Close()

	if err != nil {
		fmt.Println(err)
	}

	if written, err := io.Copy(writeFile, download.Body); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Bytes written:", written)
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
