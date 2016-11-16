package main

import (
	"context"
	"encoding/base64"
	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	"github.com/Sirupsen/logrus"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	gcsProjectID = "gcloud-crypto-testing"
)

var testStartTime string

func init() {
	log.Level = logrus.DebugLevel
	testStartTime = strconv.FormatInt(time.Now().Unix(), 10)
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

	testingBucketPrefix := "gct-" + testStartTime + "-"
	testingBucket := testingBucketPrefix + strings.ToLower(base64.RawURLEncoding.EncodeToString(randomByte(4)))
	bs := NewBucketService(*service, testingBucket, gcsProjectID)

	existingBucketsObj, _ := service.Buckets.List(gcsProjectID).Do()
	for _, b := range existingBucketsObj.Items {
		if strings.HasPrefix(b.Name, testingBucketPrefix) {

			objs, _ := NewBucketService(*service, b.Name, gcsProjectID).getObjects()
			for _, e := range objs {
				NewBucketService(*service, b.Name, gcsProjectID).deleteObject(e)
			}

			log.Info("Removing old testing bucket: " + b.Name)
			service.Buckets.Delete(b.Name).Do()
		}
	}

	if _, err := (service.Buckets.Insert(gcsProjectID, &storage.Bucket{Name: testingBucket, Location: "eu"}).Do()); err != nil {
		panic(err)
	}

	keys, err := simplecrypto.GetKeyFromPassphrase([]byte("testing"), []byte("salt1234"), 4096, 16, 1)

	if err != nil {
		panic(err)
	}

	return bs, *keys
}

func brokenSetupUp() (*bucketService, simplecrypto.Keys) {
	client, err := google.DefaultClient(context.Background(), storage.DevstorageFullControlScope)

	if err != nil {
		log.Fatalf("Unable to get default client: %v", err)
	}
	service, err := storage.New(client)

	if err != nil {
		panic(err)
	}

	userData := parseConfig()
	userData.configFile.Set("bucket", "bad")
	userData.configFile.Set("project_id", "bad")

	bs := NewBucketService(*service, "bad", "bad")
	keys, err := simplecrypto.GetKeyFromPassphrase([]byte("testing"), []byte("salt1234"), 4096, 16, 1)

	if err != nil {
		panic(err)
	}

	return bs, *keys
}

func cleanUp(bs *bucketService) {
	objs, _ := bs.getObjects()
	for _, e := range objs {
		bs.deleteObject(e)
	}
	bs.bucketCache.seenFiles = make(map[string]string, 100)
}

func randomFile() string {
	tmpfile, _ := ioutil.TempFile(".", "test")
	tmpfile.Write([]byte("this is a test string"))
	return tmpfile.Name()
}

func searchForString(slice []string, s string) bool {
	for _, e := range slice {
		if e == s {
			return true
		}
	}
	return false
}
