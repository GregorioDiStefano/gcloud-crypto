package main

import "strings"

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
