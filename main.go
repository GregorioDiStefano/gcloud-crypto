package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"

	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

func init() {
	flag.Bool("i", false, "interactive mode")
	flag.Bool("D", false, "debug mode")
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

	userData := parseConfig()

	if flag.Lookup("D").Value.String() == "true" {
		log.Level = logrus.DebugLevel
		log.Debug("Debug logging enabled")
	}

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

	cryptoKeys, err := simplecrypto.GetKeyFromPassphrase(password, userData.salt, 8192, 16, 128)

	if err != nil {
		panic(err)
	}

	client, err := google.DefaultClient(context.Background(), storage.DevstorageFullControlScope)

	if err != nil {
		panic(fmt.Sprintf("Unable to get default client: %v", err))
	}

	service, err := storage.New(client)

	if err != nil {
		panic(fmt.Sprintf("Unable to create storage service: %v", err))
	}

	bs := NewBucketService(*service, userData.configFile.GetString("bucket"), userData.configFile.GetString("project_id"))

	if err := verifyPassword(bs, *cryptoKeys); err != nil {
		log.Warn(err)
		os.Exit(1)
	}

	if flag.Lookup("i").Value.String() == "true" {
		interactiveMode(rl, bs, *cryptoKeys)
	} else {
		parseCmdLine(bs, *cryptoKeys)
	}

	os.Exit(0)
}

func verifyPassword(bs *bucketService, cryptoKeys simplecrypto.Keys) error {
	const testString = PASSWORD_CHECK_STRING
	testdata, err := simplecrypto.EncryptText(testString, cryptoKeys.EncryptionKey)

	if err != nil {
		return errors.New("unable to encrypt test string: " + err.Error())
	}

	testfile, err := bs.downloadFromBucket(PASSWORD_CHECK_FILE)
	defer os.Remove(testfile)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to find a '%s' file, if this is a new bucket, create a file called '%s' containing: %s", PASSWORD_CHECK_FILE, PASSWORD_CHECK_FILE, testdata))
	} else {
		testfileBytes, _ := ioutil.ReadFile(testfile)
		if plainText, err := simplecrypto.DecryptText(string(testfileBytes), cryptoKeys.EncryptionKey); err != nil || plainText != testString {
			return errors.New("failed to verify password: " + err.Error())
		}
	}
	return nil
}
