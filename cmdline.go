package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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
	readline.PcItem("ls"),
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
	} else if len(parsedLine) == 0 {
		return "", nil
	}
	return "", errors.New(invalidFormat)
}

func readSrcAndDstString(line string) (string, string, error) {
	if parsedLine, err := shellwords.Parse(line); err == nil && len(parsedLine) == 2 {
		return parsedLine[0], parsedLine[1], nil
	} else if len(parsedLine) == 1 {
		return parsedLine[0], "", nil
	} else {
		return "", "", errors.New(invalidFormat)
	}
}

func parseInteractiveCommand(c *client, line string) error {
	var returnedError error

	switch {
	case strings.HasPrefix(line, "upload"):
		cleanLine := strings.TrimSpace(strings.TrimLeft(line, "upload"))
		if src, dst, err := readSrcAndDstString(cleanLine); err != nil {
			returnedError = errors.New(invalidUpload)
		} else {
			returnedError = c.processUpload(src, dst)
		}
	case strings.HasPrefix(line, "ls") || strings.HasPrefix(line, "list"):
		var (
			fileList  []string
			matchGlob string
		)
		if strings.HasPrefix(line, "ls") {
			matchGlob = strings.TrimSpace(strings.TrimLeft(line, "ls"))
		} else if strings.HasPrefix(line, "list") {
			matchGlob = strings.TrimSpace(strings.TrimLeft(line, "list"))
		}
		if matchGlob, returnedError = readString(matchGlob); returnedError != nil {
			return returnedError
		} else if fileList, returnedError = c.getFileList(matchGlob); returnedError == nil {
			enumeratePrint(fileList)
		}
	case strings.HasPrefix(line, "dirs"):
		var dirList []string
		matchGlob := strings.TrimSpace(strings.TrimLeft(line, "dirs"))
		if matchGlob, returnedError = readString(matchGlob); returnedError != nil {
			fmt.Println(returnedError, matchGlob)
			return returnedError
		} else if dirList, returnedError = c.getDirList(matchGlob); returnedError == nil {
			enumeratePrint(dirList)
		}
	case strings.HasPrefix(line, "delete"):
		filepath := strings.TrimSpace(strings.TrimLeft(line, "delete"))
		if deletePath, err := readString(filepath); err != nil {
			returnedError = errors.New(invalidDelete)
		} else {
			returnedError = c.doDeleteObject(deletePath, false)
		}
	case strings.HasPrefix(line, "download"):
		cleanLine := strings.TrimSpace(strings.TrimLeft(line, "download"))
		if src, dst, err := readSrcAndDstString(cleanLine); err != nil {
			returnedError = errors.New(invalidDownload)
		} else {
			returnedError = c.doDownload(src, dst)
		}
	case strings.HasPrefix(line, "move"):
		cleanLine := strings.TrimSpace(strings.TrimLeft(line, "move"))
		if src, dst, err := readSrcAndDstString(cleanLine); err != nil {
			returnedError = errors.New(invalidMove)
		} else {
			returnedError = c.doMoveObject(src, dst)
		}
	case strings.HasPrefix(line, "exit"):
		os.Exit(0)
	default:
		fmt.Println("invalid command, try: 'upload', 'list', 'delete', 'download', 'move', 'exit'")
	}
	return returnedError
}

func interactiveMode(c *client, rl *readline.Instance) {

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
		err = parseInteractiveCommand(c, line)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
