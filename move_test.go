package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoMove(t *testing.T) {
	bs, keys := setupUp()

	defer tearDown(bs)
	tearDown(bs)

	//TODO: add error tests
	moveTests := []struct {
		src               string
		dst               string
		expectedStructure []string
	}{
		{"testdata/nested_1/*", "move/", []string{
			"move/nested_nested_1/nested_nested_nested_1/testdata1",
			"move/nested_nested_1/testdata1",
			"move/testdata1"}},
		{"testdata/testdata*", "dst/", []string{
			"dst/testdata1",
			"dst/testdata2",
			"dst/testdata3",
			"dst/testdata4",
			"dst/testdata5",
			"dst/testdata6"}},
		{"testdata/testdata1", "abc/", []string{
			"abc/testdata/testdata1"}},
		{"testdata/nested_1/*", "new_dir/", []string{
			"new_dir/nested_nested_1/nested_nested_nested_1/testdata1",
			"new_dir/nested_nested_1/testdata1",
			"new_dir/testdata1"}},
	}

	for _, e := range moveTests {
		err := processUpload(bs, &keys, e.src, "")
		assert.Nil(t, err)

		err = bs.doMoveObject(&keys, e.src, e.dst)
		assert.Nil(t, err)

		getFileList(bs, &keys, "")
		filesInBucket, _ := getFileList(bs, &keys, "")
		assert.Equal(t, e.expectedStructure, filesInBucket)
		tearDown(bs)
	}
}

func TestTransativeMove(t *testing.T) {
	bs, keys := setupUp()

	defer tearDown(bs)
	tearDown(bs)

	moveTests := []struct {
		src               string
		dst1              string
		dst2              string
		expectedStructure []string
	}{
		{"testdata/nested_1/*", "move/", "/", []string{
			"nested_nested_1/nested_nested_nested_1/testdata1",
			"nested_nested_1/testdata1",
			"testdata1"}},
	}

	for _, e := range moveTests {
		err := processUpload(bs, &keys, e.src, "")
		assert.Nil(t, err)

		err = bs.doMoveObject(&keys, e.src, e.dst1)
		assert.Nil(t, err)

		err = bs.doMoveObject(&keys, e.dst1+"*", e.dst2)
		assert.Nil(t, err)

		getFileList(bs, &keys, "")
		filesInBucket, _ := getFileList(bs, &keys, "")
		assert.Equal(t, e.expectedStructure, filesInBucket)
		tearDown(bs)
	}
}
