package main

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/GregorioDiStefano/gcloud-crypto/progress"
	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	"github.com/Sirupsen/logrus"

	googleAPI "google.golang.org/api/googleapi"
	storage "google.golang.org/api/storage/v1"
)

const (
	hashMismatchErr = "hash mismatch of uploaded file"
)

type bucketService struct {
	service *storage.Service
	keys    *simplecrypto.Keys
	bucket
}

type bucket struct {
	name    string
	project string
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
	progress.DrawProgress("Downloading", pt.totalRead, pt.contentLength)
	return c, err
}

func NewGoogleBucketService(service *storage.Service, keys *simplecrypto.Keys, bucketName, projectName string) *bucketService {
	return &bucketService{service, keys, bucket{bucketName, projectName}}
}

func (bs bucketService) Delete(encryptedFilePath string) error {
	if err := bs.service.Objects.Delete(bs.bucket.name, encryptedFilePath).Do(); err == nil {
	} else {
		return errors.New(fmt.Sprintf("Failed to delete <%s>: %s", encryptedFilePath, err.Error()))
	}
	return nil
}

func (bs bucketService) Upload(fileToUpload, encryptedUploadPath string, expectedMD5Hash []byte) error {
	defer os.Remove(fileToUpload)
	fileSize := int64(0)

	object := &storage.Object{Name: encryptedUploadPath}
	file, err := os.Open(fileToUpload)

	if err != nil {
		return errors.New("Failed opening file: " + fileToUpload + ", error: " + err.Error())
	}

	if fileStat, err := os.Stat(fileToUpload); err == nil {
		fileSize = fileStat.Size()
	}

	defer file.Close()

	if err != nil {
		fmt.Printf("Error opening %q: %v", fileToUpload, err)
	}

	// TODO: what is total? why is it 0?
	var pu googleAPI.ProgressUpdater = func(current, total int64) {
		progress.DrawProgress("Uploading", current, fileSize)
	}

	if res, err := bs.service.Objects.Insert(bs.bucket.name, object).ProgressUpdater(pu).Media(file).Do(); err == nil {
		if actualMD5Hash, err := b64.URLEncoding.DecodeString(res.Md5Hash); err == nil {
			if string(expectedMD5Hash) != string(actualMD5Hash) {
				log.WithFields(logrus.Fields{"expected md5": expectedMD5Hash, "actual md5": actualMD5Hash}).Warn("Uploaded file is corrupted")
				bs.Delete(encryptedUploadPath)
				return errors.New(hashMismatchErr)
			}
		}
		log.WithFields(logrus.Fields{"filename": encryptedUploadPath}).Debug("Created object successfully.")
	} else {
		return err
	}

	return nil
}

func (bs bucketService) Download(encryptedFilePath string) (string, error) {
	writeFile, _ := ioutil.TempFile(".", "download")
	saveFilename := writeFile.Name()
	defer writeFile.Close()

	obj := bs.service.Objects.Get(bs.bucket.name, encryptedFilePath)
	download, err := obj.Download()

	if err != nil {
		return saveFilename, errors.New("Error trying to download file:" + err.Error())
	}

	defer download.Body.Close()

	if err != nil {
		fmt.Println(err)
	}

	pt := &PassThrough{Reader: download.Body, contentLength: download.ContentLength}

	if written, err := io.Copy(writeFile, pt); err != nil {
		log.Warn("error when downloading file: %s, %s", writeFile.Name(), err.Error())
	} else if written != download.ContentLength {
		return saveFilename, errors.New("Download failed, file was not entirely downloaded")
	}

	writeFile.Close()

	return saveFilename, nil
}

func (bs bucketService) List() ([]string, error) {
	var objects []string
	pageToken := ""

	for {
		call := bs.service.Objects.List(bs.bucket.name)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Do()
		if err != nil {
			log.Errorf("error while getting object list: " + err.Error())
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

func (bs bucketService) Move(src, dst string) error {
	if rr, err := bs.service.Objects.Rewrite(bs.bucket.name, src, bs.bucket.name, dst, nil).Do(); err == nil {

		for !rr.Done {
			log.Debug("Waiting for file to be rewritten to new destination")
			time.Sleep(1 * time.Second)
		}

	} else {
		return err
	}

	if err := bs.Delete(src); err != nil {
		return err
	}

	return nil
}
