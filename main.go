package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"syscall"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
)

func init() {
	flag.Bool("list", false, "list folders/files")
	flag.String("download", "", "file to download to local disk")
	flag.String("upload", "", "file to upload to cloud")
}

func main() {
	flag.Parse()

	userData := parseConfig()

	fmt.Print("Password: ")

	password, err := terminal.ReadPassword(syscall.Stdin)
	fmt.Println()

	if err != nil {
		panic(err)
	}

	cryptoKeys := simplecrypto.GetKeyFromPassphrase(password, userData.salt)

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

	parseCmdLine(bs, *cryptoKeys)
	os.Exit(0)
}

func parseCmdLine(bs *bucketService, cryptoKeys simplecrypto.Keys) {

	switch {
	case flag.Lookup("upload").Value.String() != "":
		doUpload(bs, cryptoKeys, flag.Lookup("upload").Value.String())
	case flag.Lookup("download").Value.String() != "":
		doDownload(bs, cryptoKeys, flag.Lookup("download").Value.String())
	case flag.Lookup("list").Value.String() == "true":
		printList(bs, cryptoKeys.EncryptionKey)
	}

}

func doUpload(bs *bucketService, keys simplecrypto.Keys, filename string) {
	fmt.Println(filename)
	encrypedFile, iv, _ := simplecrypto.EncryptFile(filename, keys.EncryptionKey)
	encrypedPath := encryptFilePath(filename, keys.EncryptionKey)
	hmac := simplecrypto.CalculateHMAC(keys.HmacKey, iv, encrypedFile, false)
	addHMACToFile(encrypedFile, hmac)
	bs.uploadToBucket(encrypedFile, encrypedPath)
}

func printList(bs *bucketService, key []byte) {
	objects, err := bs.getObjects()
	if err != nil {
		panic("Failed getting objects:" + err.Error())
	}

	encToDecPaths := getEncryptedToDecryptedMap(objects, key)

	count := 0
	for i := range encToDecPaths {
		fmt.Println(count, ": ", i)
		count++
	}
}

func doDownload(bs *bucketService, keys simplecrypto.Keys, filename string) {
	objects, err := bs.getObjects()

	if err != nil {
		panic("Failed getting objects:" + err.Error())
	}

	encToDecPaths := getEncryptedToDecryptedMap(objects, keys.EncryptionKey)

	encryptedFilepath := encToDecPaths[filename]
	decryptedFilePath := decryptFilePath(encToDecPaths[filename], keys.EncryptionKey)
	downloadedEncryptedFile, _ := bs.downloadFromBucket(encryptedFilepath)

	iv := simplecrypto.GetIVFromEncryptedFile(downloadedEncryptedFile)
	actualHMAC := simplecrypto.CalculateHMAC(keys.HmacKey, iv, downloadedEncryptedFile, true)

	if expectedHMAC, err := getHMACFromFile(downloadedEncryptedFile); err == nil {
		if !bytes.Equal(actualHMAC, expectedHMAC) {
			panic("File has been tampered with!")
		}
	} else {
		panic(err.Error())
	}

	downloadedPlaintextFile, _ := simplecrypto.DecryptFile(downloadedEncryptedFile, keys.EncryptionKey)

	_, tempDownloadFilename := path.Split(downloadedPlaintextFile)
	_, actualFilename := path.Split(decryptedFilePath)

	os.Rename(tempDownloadFilename, actualFilename)
	truncateHMACSignature(actualFilename)

	os.Remove(downloadedEncryptedFile)
	os.Remove(downloadedPlaintextFile)

	fmt.Println("Downloaded: " + actualFilename)
}
