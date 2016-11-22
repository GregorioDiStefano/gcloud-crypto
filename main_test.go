package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
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
	keys, err := simplecrypto.GetKeyFromPassphrase([]byte("testing"), []byte("salt1234"), 4096, 16, 1)

	bs := NewGoogleBucketService(service, keys, testingBucket, gcsProjectID)

	existingBucketsObj, _ := service.Buckets.List(gcsProjectID).Do()
	for _, b := range existingBucketsObj.Items {
		if strings.HasPrefix(b.Name, testingBucketPrefix) {

			objs, _ := NewGoogleBucketService(service, keys, b.Name, gcsProjectID).List()
			for _, e := range objs {
				NewGoogleBucketService(service, keys, b.Name, gcsProjectID).Delete(e)
			}

			log.Info("Removing old testing bucket: " + b.Name)
			service.Buckets.Delete(b.Name).Do()
		}
	}

	if _, err := (service.Buckets.Insert(gcsProjectID, &storage.Bucket{Name: testingBucket, Location: "eu"}).Do()); err != nil {
		panic(err)
	}

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

	keys, err := simplecrypto.GetKeyFromPassphrase([]byte("testing"), []byte("salt1234"), 4096, 16, 1)
	bs := NewGoogleBucketService(service, keys, "bad", "bad")

	if err != nil {
		panic(err)
	}

	return bs, *keys
}

func cleanUp(c *client) {
	objs, _ := c.bucket.List()
	for _, e := range objs {
		c.bucket.Delete(e)
	}

	c.bcache.seenFiles = make(map[string]string, 100)
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

func TestVerifyPassword(t *testing.T) {
	bs, keys := setupUp()
	c := client{&keys, bs, bucketCache{}}

	tf, _ := ioutil.TempFile("/tmp", "testing")
	defer os.Remove(tf.Name())

	// test with correct password
	tf.WriteString("oEFyYW0rbdcMOif8pzS8McO4tVRvHvX6uVNTkmYad4sPcQ4M")
	md5hex, err := hex.DecodeString("3483ba92a60078005e30a70200e0827b")
	assert.Nil(t, err)

	err = c.bucket.Upload(tf.Name(), PASSWORD_CHECK_FILE, md5hex)
	assert.Nil(t, err)

	err = verifyPassword(bs, &keys)
	assert.Nil(t, err)

	// test when keycontents doesn't match password
	tf2, _ := ioutil.TempFile("/tmp", "testing")
	tf2.WriteString("wrongpasssword")
	md5hex, err = hex.DecodeString("deadbeef")
	assert.Nil(t, err)

	err = c.bucket.Upload(tf2.Name(), PASSWORD_CHECK_FILE, md5hex)
	assert.Nil(t, err)

	err = verifyPassword(bs, &keys)
	assert.NotNil(t, err)
}
