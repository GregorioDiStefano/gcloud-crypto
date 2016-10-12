package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"github.com/chzyer/readline"
	"github.com/mattn/go-shellwords"
)

const (
	invalidFormat   = "invalid command line"
	invalidUpload   = "invalid upload request; try using 'upload <file>' or 'upload <file> <destination directroy>'"
	invalidDownload = "invalid download request; try using 'download <file>' or 'download <file> <destination folder>'"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("upload"),
	readline.PcItem("download"),
	readline.PcItem("delete"),
	readline.PcItem("list"),
	readline.PcItem("exit"),
)

func setupReadline() (*readline.Instance, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline-gcloud-enc.tmp",
		InterruptPrompt: "^C",
		AutoComplete:    completer,
		EOFPrompt:       "exit",
	})

	return l, err
}

// getSrcDestString seperates a line to src and dst path
func getSrcDestString(line string) (src, dst string, err error) {
	parsedLine, _ := shellwords.Parse(line)

	srcFile := ""
	destinationDirectory := ""

	if len(parsedLine) == 1 {
		srcFile = parsedLine[0]
	} else if len(parsedLine) == 2 {
		srcFile = parsedLine[0]
		destinationDirectory = parsedLine[1]
	} else {
		return "", "", errors.New(invalidFormat)
	}

	return srcFile, destinationDirectory, nil
}

func interactiveMode(rl *readline.Instance, bs *bucketService, cryptoKeys simplecrypto.Keys) {
	var returnedError error

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "upload"):
			cleanLine := (strings.TrimSpace(strings.TrimLeft(line, "upload")))
			if srcFile, destinationDirectory, err := getSrcDestString(cleanLine); err != nil {
				returnedError = errors.New(invalidUpload)
			} else {
				returnedError = processUpload(bs, cryptoKeys, srcFile, destinationDirectory)
			}
		case strings.HasPrefix(line, "list") || strings.HasPrefix(line, "ls"):
			if fileList, returnedError := getFileList(bs, cryptoKeys.EncryptionKey); returnedError == nil {
				enumeratePrint(fileList)
			}
		case strings.HasPrefix(line, "delete"):
			filepath := strings.TrimSpace(strings.TrimLeft(line, "delete"))
			returnedError = doDeleteObject(bs, cryptoKeys, filepath)
		case strings.HasPrefix(line, "download"):
			cleanLine := (strings.TrimSpace(strings.TrimLeft(line, "download")))
			if srcFile, destinationDirectory, err := getSrcDestString(cleanLine); err != nil {
				returnedError = errors.New(invalidDownload)
			} else {
				returnedError = doDownload(bs, cryptoKeys, srcFile, destinationDirectory)
			}
		case strings.HasPrefix(line, "exit"):
			return
		default:
			fmt.Println("invalid command, try: 'upload', 'list', 'delete', 'download', 'exit'")
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
		returnedError = doDownload(bs, cryptoKeys, flag.Lookup("download").Value.String(), flag.Lookup("dir").Value.String())
	case flag.Lookup("list").Value.String() == "true":
		if fileList, returnedError := getFileList(bs, cryptoKeys.EncryptionKey); returnedError == nil {
			enumeratePrint(fileList)
		}
	}

	if returnedError != nil {
		fmt.Println("Action returned error: " + returnedError.Error())
		os.Exit(1)
	}
}
