package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"

	"github.com/Sirupsen/logrus"
)

var (
	Version         string
	BuildTime       string
	CompilerVersion string
)

var log = logrus.New()

type client struct {
	keys   *simplecrypto.Keys
	bucket Bucket
	bcache bucketCache
}

func init() {
	flag.Bool("version", false, "debug mode")
	flag.Bool("debug", false, "debug mode")
	flag.Bool("list", false, "list folders/files")
	flag.String("delete", "", "delete object")
	flag.String("download", "", "file to download to local disk")
	flag.String("upload", "", "file to upload to cloud")
	flag.String("dir", "", "directory to store uploaded file to")
}

const (
	PASSWORD_CHECK_STRING = "keyCheck"
	PASSWORD_CHECK_FILE   = "keycheck"
)

func main() {
	flag.Parse()

	if flag.Lookup("debug").Value.String() == "true" {
		log.Level = logrus.DebugLevel
		log.Debug("Debug logging enabled")
	}

	if flag.Lookup("version").Value.String() == "true" {
		log.Infof("Version: %s", Version)
		log.Infof("Build date: %s", BuildTime)
		log.Infof("Compiler version: %s", CompilerVersion)
		os.Exit(0)
	}

	userData := parseConfig()

	fmt.Print("Password: ")

	password, err := terminal.ReadPassword(syscall.Stdin)

	fmt.Println()
	fmt.Println()

	if err != nil {
		panic(err)
	}

	rl, err := setupReadline()

	if err != nil {
		panic(err)
	}

	keys, err := simplecrypto.GetKeyFromPassphrase(password, userData.salt, 8192, 16, 128)

	if err != nil {
		panic(err)
	}

	googleClient, err := google.DefaultClient(context.Background(), storage.DevstorageFullControlScope)

	if err != nil {
		panic(fmt.Sprintf("Unable to get default client: %v", err))
	}

	service, err := storage.New(googleClient)

	if err != nil {
		panic(fmt.Sprintf("Unable to create storage service: %v", err))
	}

	bucket := NewGoogleBucketService(service, keys, userData.configFile.GetString("bucket"), userData.configFile.GetString("project_id"))

	if err := verifyPassword(bucket, keys); err != nil {
		log.Warn(err)
		os.Exit(1)
	}

	c := &client{keys, bucket, bucketCache{}}
	interactiveMode(c, rl)
	os.Exit(0)
}

func verifyPassword(bucket Bucket, keys *simplecrypto.Keys) error {
	testdata, err := simplecrypto.EncryptText(PASSWORD_CHECK_STRING, keys.EncryptionKey)

	if err != nil {
		return errors.New("unable to encrypt test string: " + err.Error())
	}

	testfile, err := bucket.Download(PASSWORD_CHECK_FILE)
	defer os.Remove(testfile)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to find a '%s' file, if this is a new bucket, create a file called '%s' containing: %s", PASSWORD_CHECK_FILE, PASSWORD_CHECK_FILE, testdata))
	} else {
		testfileBytes, _ := ioutil.ReadFile(testfile)
		if plainText, err := simplecrypto.DecryptText(string(testfileBytes), keys.EncryptionKey); err != nil || plainText != PASSWORD_CHECK_STRING {
			return errors.New("failed to verify password: " + err.Error())
		}
	}
	return nil
}
