package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"github.com/ryanuber/go-glob"
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
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("> ")
			reader.ReadString('\n')
		}
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

func parseCmdLine(bs *bucketService, cryptoKeys simplecrypto.Keys) {
	var returnedError error

	switch {
	case flag.Lookup("delete").Value.String() != "":
		returnedError = doDeleteObject(bs, cryptoKeys, flag.Lookup("delete").Value.String())
	case flag.Lookup("upload").Value.String() != "":
		path := flag.Lookup("upload").Value.String()
		returnedError = processUpload(bs, path, cryptoKeys)
	case flag.Lookup("download").Value.String() != "":
		returnedError = doDownload(bs, cryptoKeys, flag.Lookup("download").Value.String())
	case flag.Lookup("list").Value.String() == "true":
		printList(bs, cryptoKeys.EncryptionKey)
	}

	if returnedError != nil {
		fmt.Println("Action returned error: " + returnedError.Error())
		os.Exit(1)
	}
}

func processUpload(bs *bucketService, path string, cryptoKeys simplecrypto.Keys) error {
	globMatch, err := filepath.Glob(path)
	var errorUploading bool
	fmt.Println(globMatch, err)
	if err != nil {
		return err
	}

	for _, path := range globMatch {
		fmt.Println("Uploading: " + path)
		if isDir(path) {
			err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if !isDir(path) {
					fmt.Println("Uploading: " + path)
					return doUpload(bs, cryptoKeys, path, flag.Lookup("dir").Value.String())
				}
				return nil
			})
		} else {
			err = doUpload(bs, cryptoKeys, path, flag.Lookup("dir").Value.String())
		}

		if err != nil {
			errorUploading = true
			fmt.Println(fmt.Sprintf("failed with %s when uploading: %s", err.Error(), path))
		}
	}

	if errorUploading {
		return errors.New("at least one file failed to upload")
	}
	return nil
}

func doUpload(bs *bucketService, keys simplecrypto.Keys, filePath, remoteDirectory string) error {
	encryptedFile, iv, _ := simplecrypto.EncryptFile(filePath, keys.EncryptionKey)
	var encryptedPath string

	if len(remoteDirectory) > 0 {
		remoteDirectoryFilename := path.Clean(remoteDirectory + "/" + path.Base(filePath))
		encryptedPath = encryptFilePath(remoteDirectoryFilename, keys.EncryptionKey)
	} else {
		encryptedPath = encryptFilePath(filePath, keys.EncryptionKey)
	}

	if objects, err := bs.getObjects(); err == nil {
		encToDecPaths := getEncryptedToDecryptedMap(objects, keys.EncryptionKey)
		for e := range encToDecPaths {
			//stupid
			plainTextRemotePath := decryptFilePath(encryptedPath, keys.EncryptionKey)
			if e == plainTextRemotePath {
				return errors.New("This file already exists, delete it first.")
			}
		}
	} else {
		fmt.Println(err)
	}

	hmac, err := simplecrypto.CalculateHMAC(keys.HMACKey, iv, encryptedFile, false)

	if err != nil {
		return err
	}

	simplecrypto.AddHMACToFile(encryptedFile, hmac)
	bs.uploadToBucket(encryptedFile, encryptedPath)
	return nil
}

func printList(bs *bucketService, key []byte) {
	objects, err := bs.getObjects()
	if err != nil {
		panic("failed getting objects:" + err.Error())
	}

	encToDecPaths := getEncryptedToDecryptedMap(objects, key)

	count := 0
	for i := range encToDecPaths {
		fmt.Println(count, ": ", i)
		count++
	}
}

func doDownload(bs *bucketService, keys simplecrypto.Keys, filename string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	encToDecPaths := getEncryptedToDecryptedMap(objects, keys.EncryptionKey)

	for k, _ := range encToDecPaths {
		//TODO: should be decToEncPaths
		fmt.Println(filename, k)
		if glob.Glob(filename, k) {
			fmt.Println("downloading: " + k)
			encryptedFilepath := encToDecPaths[k]
			decryptedFilePath := decryptFilePath(encToDecPaths[k], keys.EncryptionKey)
			downloadedEncryptedFile, _ := bs.downloadFromBucket(encryptedFilepath)
			defer os.Remove(downloadedEncryptedFile)

			iv, err := simplecrypto.GetIVFromEncryptedFile(downloadedEncryptedFile)

			if err != nil {
				panic(err)
			}

			actualHMAC, err := simplecrypto.CalculateHMAC(keys.HMACKey, iv, downloadedEncryptedFile, true)

			if err != nil {
				return err
			}

			if expectedHMAC, err := simplecrypto.GetHMACFromFile(downloadedEncryptedFile); err == nil {
				if !bytes.Equal(actualHMAC, expectedHMAC) {
					panic("file has been tampered with!")
				}
			} else {
				panic(err.Error())
			}

			downloadedPlaintextFile, _ := simplecrypto.DecryptFile(downloadedEncryptedFile, keys.EncryptionKey)
			defer os.Remove(downloadedPlaintextFile)

			_, tempDownloadFilename := path.Split(downloadedPlaintextFile)
			_, actualFilename := path.Split(decryptedFilePath)

			os.Rename(tempDownloadFilename, actualFilename)
			simplecrypto.TruncateHMACSignature(actualFilename)
			fmt.Println("downloaded: " + actualFilename)
		}
	}
	return nil
}

func doDeleteObject(bs *bucketService, keys simplecrypto.Keys, filename string) error {
	objects, err := bs.getObjects()

	if err != nil {
		return errors.New("failed getting objects: " + err.Error())
	}

	encToDecPaths := getEncryptedToDecryptedMap(objects, keys.EncryptionKey)

	encryptedFilename := encToDecPaths[filename]

	if encryptedFilename == "" {
		return fmt.Errorf("file: %s not found in bucket", filename)
	}

	return bs.deleteObject(encryptedFilename)
}
