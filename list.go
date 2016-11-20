package main

import (
	"path/filepath"
	"sort"

	"github.com/ryanuber/go-glob"
)

func (c *client) getDirList(matchGlob string) ([]string, error) {
	objects, err := c.bucket.List()
	if err != nil {
		return nil, err
	}

	dirs := []string{}
	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, c.keys)

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

func (c *client) getFileList(matchGlob string) ([]string, error) {
	objects, err := c.bucket.List()
	if err != nil {
		return nil, err
	}

	decToEncPaths := getDecryptedToEncryptedFileMapping(objects, c.keys)

	var keys []string
	for k := range decToEncPaths {
		if k == PASSWORD_CHECK_FILE {
			continue
		}
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
