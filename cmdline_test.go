package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInteractiveMode(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)

	rl, err := setupReadline()
	assert.Nil(t, err)

	interactiveMode(c, rl)
	parseInteractiveCommand(c, "upload testdata/testdata1 abc/")

	dirs, err := c.getDirList("")
	assert.Nil(t, err)
	assert.Equal(t, []string{"abc/testdata"}, dirs)

	filesUploaded, err := c.getFileList("*")
	assert.Nil(t, err)
	assert.Equal(t, []string{"abc/testdata/testdata1"}, filesUploaded)

	parseInteractiveCommand(c, "move abc/testdata/* /")
	filesUploaded, err = c.getFileList("*")
	assert.Nil(t, err)
	assert.Equal(t, []string{"testdata1"}, filesUploaded)

	err = parseInteractiveCommand(c, "download testdata1 /tmp/")
	assert.Nil(t, err)

	parseInteractiveCommand(c, "delete testdata1")
	filesUploaded, err = c.getFileList("*")
	assert.Nil(t, err)
	assert.Len(t, filesUploaded, 0)
}
