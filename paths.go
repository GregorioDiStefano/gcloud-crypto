package main

import (
	"os"
	"path/filepath"
	"strings"
)

func relativePathFromGlob(glob, match string) string {
	splitGlob := strings.Split(glob, "/")
	splitMatch := strings.Split(match, "/")
	commonPath := ""
	for i, s := range splitGlob {
		if splitMatch[i] == s {
			commonPath += s + "/"
		} else {
			break
		}
	}

	return strings.TrimPrefix(match, commonPath)
}

func globMatchWithDirectories(path string) []string {
	globMatch, _ := filepath.Glob(path)
	matches := []string{}

	for _, matchedPath := range globMatch {
		filepath.Walk(matchedPath, func(matchedPath string, info os.FileInfo, err error) error {
			if !isDir(matchedPath) {
				matches = append(matches, matchedPath)
			}
			return nil
		})
	}
	return matches
}
