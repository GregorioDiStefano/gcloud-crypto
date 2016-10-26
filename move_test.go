package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoMove(t *testing.T) {
	bs, keys := setupUp()
	defer tearDown(bs)

	moveTests := []struct {
		src               string
		dst               string
		expectedStructure []string
	}{
		{"testdata/testdata*", "dst/", []string{
			"dst/testdata/testdata1",
			"dst/testdata/testdata2",
			"dst/testdata/testdata3",
			"dst/testdata/testdata4",
			"dst/testdata/testdata5",
			"dst/testdata/testdata6"}},
	}

	for _, e := range moveTests {
		err := processUpload(bs, keys, e.src, "")
		assert.Nil(t, err)

		bs.doMoveObject(keys, "/"+e.src, e.dst)

		getFileList(bs, keys.EncryptionKey, "")
		filesInBucket, _ := getFileList(bs, keys.EncryptionKey, "")
		assert.Equal(t, e.expectedStructure, filesInBucket)
	}
}
