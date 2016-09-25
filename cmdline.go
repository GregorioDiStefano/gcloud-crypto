package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
)

func interactiveMode(bs *bucketService, cryptoKeys simplecrypto.Keys) {
	var workingDirectory string
	var returnedError error
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')

		switch {
		case strings.HasPrefix(line, "cd"):
			workingDirectory = strings.TrimSpace(strings.TrimLeft(line, "cd"))
		case strings.HasPrefix(line, "upload"):
			file := strings.TrimSpace(strings.TrimLeft(line, "upload"))
			returnedError = processUpload(bs, cryptoKeys, file, workingDirectory)
		case strings.HasPrefix(line, "list"):
			returnedError = printList(bs, cryptoKeys.EncryptionKey)
		case strings.HasPrefix(line, "delete"):
			filepath := strings.TrimSpace(strings.TrimLeft(line, "delete"))
			returnedError = doDeleteObject(bs, cryptoKeys, filepath)
		default:
			fmt.Println("Invalid command, try: 'cd', 'upload', 'list', 'delete'")
		}
		if returnedError != nil {
			fmt.Println(returnedError)
		}
	}
}

func parseCmdLine(bs *bucketService, cryptoKeys simplecrypto.Keys) {
	var returnedError error

	switch {
	case flag.Lookup("delete").Value.String() != "":
		returnedError = doDeleteObject(bs, cryptoKeys, flag.Lookup("delete").Value.String())
	case flag.Lookup("upload").Value.String() != "":
		path := flag.Lookup("upload").Value.String()
		returnedError = processUpload(bs, cryptoKeys, path, flag.Lookup("dir").Value.String())
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
