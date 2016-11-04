package main

import (
	"path/filepath"
	"sort"

	"github.com/GregorioDiStefano/gcloud-crypto/simplecrypto"
	"github.com/ryanuber/go-glob"
)

func getDirList(bs *bucketService, key *simplecrypto.Keys, matchGlob string) ([]string, error) {
	objects, err := bs.getObjects()
	if err != nil {
		return nil, err
	}

	dirs := []string{}
	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, key)

	for e := range decToEncPaths {
		e = filepath.Dir(e)

		if isStringInSlice(e, dirs) == false {
			if len(matchGlob) > 0 && glob.Glob(matchGlob, e) {
				dirs = append(dirs, e)
			}
			if len(matchGlob) == 0 {
				dirs = append(dirs, e)
			}
		}
	}

	sort.Strings(dirs)
	return dirs, nil
}

func getFileList(bs *bucketService, key *simplecrypto.Keys, matchGlob string) ([]string, error) {
	objects, err := bs.getObjects()
	if err != nil {
		return nil, err
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, key)

	var keys []string
	for k := range decToEncPaths {
		if len(matchGlob) > 0 && glob.Glob(matchGlob, k) {
			keys = append(keys, k)
		}
		if len(matchGlob) == 0 {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)
	return keys, nil
}
