package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GregorioDiStefano/gcloud-fuse/simplecrypto"
	"github.com/chzyer/readline"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("cd"),
	readline.PcItem("upload"),
	readline.PcItem("download"),
	readline.PcItem("delete"),
	readline.PcItem("list"),
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
			workingDirectory := ""
			file := ""

			if len(strings.Fields(line)) == 3 {
				workingDirectory = strings.Fields(line)[2]
				file = strings.Fields(line)[1]
			} else if len(strings.Fields(line)) > 3 {
				fmt.Println("invalid upload request; try using 'upload <file>' or 'upload <file> <destination folder>' ")
			}

			returnedError = processUpload(bs, cryptoKeys, file, workingDirectory)
		case strings.HasPrefix(line, "list") || strings.HasPrefix(line, "ls"):
			returnedError = printList(bs, cryptoKeys.EncryptionKey)
		case strings.HasPrefix(line, "delete"):
			filepath := strings.TrimSpace(strings.TrimLeft(line, "delete"))
			returnedError = doDeleteObject(bs, cryptoKeys, filepath)
		case strings.HasPrefix(line, "download"):
			filepath := strings.TrimSpace(strings.TrimLeft(line, "download"))
			returnedError = doDownload(bs, cryptoKeys, filepath)
		default:
			fmt.Println("invalid command, try: 'upload', 'list', 'delete', 'download'")
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
