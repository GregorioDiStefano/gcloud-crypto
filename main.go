package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"syscall"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
)

func init() {
	flag.Bool("i", false, "interactive mode")
	flag.Bool("list", false, "list folders/files")
	flag.String("delete", "", "delete object")
	flag.String("download", "", "file to download to local disk")
	flag.String("upload", "", "file to upload to cloud")

	flag.String("dir", "", "directory to store uploaded file to")
}

func main() {
	flag.Parse()
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

	cryptoKeys, _ := simplecrypto.GetKeyFromPassphrase(password, userData.salt)
	client, err := google.DefaultClient(context.Background(), storage.DevstorageFullControlScope)

	if err != nil {
		log.Fatalf("Unable to get default client: %v", err)
	}

	service, err := storage.New(client)

	if err != nil {
		log.Fatalf("Unable to create storage service: %v", err)
	}

	if _, err := service.Buckets.Get(userData.configFile.GetString("bucket")).Do(); err == nil {
	}

	bs := NewBucketService(*service, userData.configFile.GetString("bucket"), userData.configFile.GetString("project_id"))

	if err := verifyPassword(bs, *cryptoKeys); err != nil {
		fmt.Println(err)
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
	const testString = "keyCheck"
	testdata, err := simplecrypto.EncryptText(testString, cryptoKeys.EncryptionKey)

	if err != nil {
		return errors.New("unable to encrypt test string: " + err.Error())
	}

	testfile, err := bs.downloadFromBucket("keycheck")
	defer os.Remove(testfile)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to find a \"keycheck\" file, if this is a new bucket, create a file called, \"keycheck\" containing: %s", testdata))
	} else {
		testfileBytes, _ := ioutil.ReadFile(testfile)
		if plainText, err := simplecrypto.DecryptText(string(testfileBytes), cryptoKeys.EncryptionKey); err != nil || plainText != testString {
			return errors.New("failed to verify bucket is using specified password: " + err.Error())
		}
	}
	return nil
}

func getFileList(bs *bucketService, key []byte) ([]string, error) {
	objects, err := bs.getObjects()
	if err != nil {
		return nil, err
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, key)

	var keys []string
	for k := range decToEncPaths {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys, nil
}
