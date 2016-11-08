package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInteractiveMode(t *testing.T) {
	bs, keys := setupUp()
	cleanUp(bs)

	rl, err := setupReadline()
	assert.Nil(t, err)

	interactiveMode(rl, bs, &keys)
	parseInteractiveCommand(bs, &keys, "upload testdata/testdata1 abc/")

	dirs, err := getDirList(bs, &keys, "")
	assert.Nil(t, err)
	assert.Equal(t, []string{"abc/testdata"}, dirs)

	filesUploaded, err := getFileList(bs, &keys, "*")
	assert.Nil(t, err)
	assert.Equal(t, []string{"abc/testdata/testdata1"}, filesUploaded)

	parseInteractiveCommand(bs, &keys, "move abc/testdata/* /")
	filesUploaded, err = getFileList(bs, &keys, "*")
	assert.Nil(t, err)
	assert.Equal(t, []string{"testdata1"}, filesUploaded)

	err = parseInteractiveCommand(bs, &keys, "download testdata1 /tmp/")
	assert.Nil(t, err)

	parseInteractiveCommand(bs, &keys, "delete testdata1")
	filesUploaded, err = getFileList(bs, &keys, "*")
	assert.Nil(t, err)
	assert.Len(t, filesUploaded, 0)
}
