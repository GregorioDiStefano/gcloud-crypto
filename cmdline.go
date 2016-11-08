package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	"github.com/chzyer/readline"
	"github.com/mattn/go-shellwords"
)

const (
	invalidFormat   = "invalid command line"
	invalidUpload   = "invalid upload request; try using 'upload <file>' or 'upload <file> <destination directroy>'"
	invalidDelete   = "invalid delete request; try using 'delete' <path>"
	invalidDownload = "invalid download request; try using 'download <file>' or 'download <file> <destination folder>'"
	invalidMove     = "invalid move request; try using 'move <file>' or 'move <file> <destination folder>'"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("upload"),
	readline.PcItem("dirs"),
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

func readString(line string) (string, error) {
	if parsedLine, err := shellwords.Parse(line); err == nil && len(parsedLine) == 1 {
		return parsedLine[0], nil
	} else {
		return "", errors.New(invalidFormat)
	}
}

func readSrcAndDstString(line string) (string, string, error) {
	if parsedLine, err := shellwords.Parse(line); err == nil && len(parsedLine) == 2 {
		return parsedLine[0], parsedLine[1], nil
	} else {
		return "", "", errors.New(invalidFormat)
	}
}

func parseInteractiveCommand(bs *bucketService, keys *simplecrypto.Keys, line string) error {
	var returnedError error

	switch {
	case strings.HasPrefix(line, "upload"):
		cleanLine := strings.TrimSpace(strings.TrimLeft(line, "upload"))
		if src, dst, err := readSrcAndDstString(cleanLine); err != nil {
			returnedError = errors.New(invalidUpload)
		} else {
			returnedError = processUpload(bs, keys, src, dst)
		}
	case strings.HasPrefix(line, "ls"):
		var fileList []string
		matchGlob := strings.TrimSpace(strings.TrimLeft(line, "ls"))
		if fileList, returnedError = getFileList(bs, keys, matchGlob); returnedError == nil {
			enumeratePrint(fileList)
		}
	case strings.HasPrefix(line, "dirs"):
		var dirList []string
		matchGlob := strings.TrimSpace(strings.TrimLeft(line, "dirs"))
		if dirList, returnedError = getDirList(bs, keys, matchGlob); returnedError == nil {
			enumeratePrint(dirList)
		}
	case strings.HasPrefix(line, "delete"):
		filepath := strings.TrimSpace(strings.TrimLeft(line, "delete"))
		if deletePath, err := readString(filepath); err != nil {
			returnedError = errors.New(invalidDelete)
		} else {
			returnedError = bs.doDeleteObject(keys, deletePath, false)
		}
	case strings.HasPrefix(line, "download"):
		cleanLine := strings.TrimSpace(strings.TrimLeft(line, "download"))
		if src, dst, err := readSrcAndDstString(cleanLine); err != nil {
			returnedError = errors.New(invalidDownload)
		} else {
			returnedError = doDownload(bs, keys, src, dst)
		}
	case strings.HasPrefix(line, "move"):
		cleanLine := strings.TrimSpace(strings.TrimLeft(line, "move"))
		if src, dst, err := readSrcAndDstString(cleanLine); err != nil {
			returnedError = errors.New(invalidMove)
		} else {
			returnedError = bs.doMoveObject(keys, src, dst)
		}
	case strings.HasPrefix(line, "exit"):
		os.Exit(0)
	default:
		fmt.Println("invalid command, try: 'upload', 'list', 'delete', 'download', 'move', 'exit'")
	}
	return returnedError
}

func interactiveMode(rl *readline.Instance, bs *bucketService, cryptoKeys *simplecrypto.Keys) {

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
		err = parseInteractiveCommand(bs, cryptoKeys, line)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func parseCmdLine(bs *bucketService, cryptoKeys *simplecrypto.Keys) {
	var returnedError error

	switch {
	case flag.Lookup("delete").Value.String() != "":
		returnedError = bs.doDeleteObject(cryptoKeys, flag.Lookup("delete").Value.String(), false)
	case flag.Lookup("upload").Value.String() != "":
		path := flag.Lookup("upload").Value.String()
		returnedError = processUpload(bs, cryptoKeys, path, flag.Lookup("dir").Value.String())
	case flag.Lookup("download").Value.String() != "":
		returnedError = doDownload(bs, cryptoKeys, flag.Lookup("download").Value.String(), flag.Lookup("dir").Value.String())
	case flag.Lookup("list").Value.String() == "true":
		var fileList []string
		if fileList, returnedError = getFileList(bs, cryptoKeys, ""); returnedError == nil {
			enumeratePrint(fileList)
		}
	}

	if returnedError != nil {
		fmt.Println("Action returned error: " + returnedError.Error())
		os.Exit(1)
	}
}
