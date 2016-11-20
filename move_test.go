package main

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoMove(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)

	//TODO: add error tests
	moveTests := []struct {
		uploadSrc         string
		uploadDst         string
		moveSrc           string
		moveDst           string
		expectedStructure []string
	}{
		{"testdata/nested_1/*", "", "*", "move/", []string{
			"move/nested_nested_1/nested_nested_nested_1/testdata1",
			"move/nested_nested_1/testdata1",
			"move/testdata1"}},
		{"testdata/nested_1/", "", "testdata/", "move/", []string{
			"move/testdata/nested_1/nested_nested_1/nested_nested_nested_1/testdata1",
			"move/testdata/nested_1/nested_nested_1/testdata1",
			"move/testdata/nested_1/testdata1"}},
		{"testdata/testdata*", "a_dir", "a_dir/testdata*", "dst/", []string{
			"dst/testdata1",
			"dst/testdata2",
			"dst/testdata3",
			"dst/testdata4",
			"dst/testdata5",
			"dst/testdata6"}},
		{"testdata/nested_3/", "", "*", "dst/", []string{
			"dst/testdata/nested_3/testdata1",
			"dst/testdata/nested_3/testdata2",
			"dst/testdata/nested_3/testdata3",
			"dst/testdata/nested_3/testdata4"}},
		{"testdata/testdata1", "", "testdata/*", "abc/", []string{
			"abc/testdata1"}},
	}

	for n, e := range moveTests {
		fmt.Println("Test #", n)
		err := c.processUpload(e.uploadSrc, e.uploadDst)
		assert.Nil(t, err)

		err = c.doMoveObject(e.moveSrc, e.moveDst)
		assert.Nil(t, err)

		filesInBucket, _ := c.getFileList("")
		assert.Equal(t, e.expectedStructure, filesInBucket)
		cleanUp(c)
	}
}

func TestMovePartialDirectoryWithoutGlob(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)

	err := c.processUpload("testdata/", "")
	assert.Nil(t, err)

	err = c.doMoveObject("testdata/nested_3/", "testdata_moved/")
	assert.Nil(t, err)

	expectedObjects := []string{
		"testdata/nested_1/nested_nested_1/nested_nested_nested_1/testdata1",
		"testdata/nested_1/nested_nested_1/testdata1",
		"testdata/nested_1/testdata1",
		"testdata/nested_2/testdata2",
		"testdata_moved/testdata/nested_3/testdata1",
		"testdata_moved/testdata/nested_3/testdata2",
		"testdata_moved/testdata/nested_3/testdata3",
		"testdata_moved/testdata/nested_3/testdata4",
		"testdata/test_a/a",
		"testdata/test_b/b",
		"testdata/testdata1",
		"testdata/testdata2",
		"testdata/testdata3",
		"testdata/testdata4",
		"testdata/testdata5",
		"testdata/testdata6"}

	filesInBucket, _ := c.getFileList("")
	sort.Strings(filesInBucket)
	sort.Strings(expectedObjects)
	assert.EqualValues(t, expectedObjects, filesInBucket)
}

func TestMovePartialDirectoryWithGlob(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)

	err := c.processUpload("testdata/", "")
	assert.Nil(t, err)

	err = c.doMoveObject("testdata/nested_1/*", "/")
	assert.Nil(t, err)

	expectedObjects := []string{
		"nested_nested_1/nested_nested_nested_1/testdata1",
		"nested_nested_1/testdata1",
		"testdata1",
		"testdata/nested_2/testdata2",
		"testdata/nested_3/testdata1",
		"testdata/nested_3/testdata2",
		"testdata/nested_3/testdata3",
		"testdata/nested_3/testdata4",
		"testdata/test_a/a",
		"testdata/test_b/b",
		"testdata/testdata1",
		"testdata/testdata2",
		"testdata/testdata3",
		"testdata/testdata4",
		"testdata/testdata5",
		"testdata/testdata6"}

	filesInBucket, _ := c.getFileList("")
	sort.Strings(filesInBucket)
	sort.Strings(expectedObjects)
	assert.EqualValues(t, expectedObjects, filesInBucket)
}

func TestMoveInEmptyBucket(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)

	err := c.doMoveObject("12345/*", "test/")
	assert.Error(t, err)
}

/*
func TestMoveOfNonExistingFile(t *testing.T) {
	bs, keys := setupUp()
	cleanUp(bs)

	err := processUpload(bs, &keys, "testdata/", "")
	assert.Nil(t, err)
	err = bs.doMoveObject(&keys, "12345/*", "test/")
	assert.Error(t, err)
}
*/

func TestMoveFailGettingObjects(t *testing.T) {
	bs, keys := brokenSetupUp()
	c := &client{&keys, bs, bucketCache{}}
	err := c.doMoveObject("12345/*", "test/")
	assert.Error(t, err)
}

func TestTransativeMove(t *testing.T) {
	bs, keys := setupUp()
	c := &client{&keys, bs, bucketCache{}}
	cleanUp(c)
	c.doMoveObject("12345/*", "test/")

	moveTests := []struct {
		uploadSrc                  string
		uploadDst                  string
		src1                       string
		dst1                       string
		src2                       string
		dst2                       string
		expectedStructureAfterDst1 []string
		expectedStructureAfterDst2 []string
	}{
		{"testdata/nested_1/*", "", "*", "move/", "move/*", "/",
			[]string{"move/nested_nested_1/testdata1",
				"move/nested_nested_1/nested_nested_nested_1/testdata1",
				"move/testdata1"},
			[]string{
				"nested_nested_1/nested_nested_nested_1/testdata1",
				"nested_nested_1/testdata1",
				"testdata1"}},

		{"testdata/nested_1/*", "foo", "foo/*", "move1/move2/", "*", "new_directory/",
			[]string{
				"move1/move2/nested_nested_1/testdata1",
				"move1/move2/nested_nested_1/nested_nested_nested_1/testdata1",
				"move1/move2/testdata1"},
			[]string{
				"new_directory/move1/move2/nested_nested_1/nested_nested_nested_1/testdata1",
				"new_directory/move1/move2/nested_nested_1/testdata1",
				"new_directory/move1/move2/testdata1"}},
	}

	for _, e := range moveTests {
		err := c.processUpload(e.uploadSrc, e.uploadDst)
		assert.Nil(t, err)

		filesInBucket, _ := c.getFileList("")

		err = c.doMoveObject(e.src1, e.dst1)
		c.getFileList("")
		filesInBucket, _ = c.getFileList("")
		sort.Strings(filesInBucket)
		sort.Strings(e.expectedStructureAfterDst1)
		assert.EqualValues(t, e.expectedStructureAfterDst1, filesInBucket)
		assert.Nil(t, err)

		err = c.doMoveObject(e.src2, e.dst2)
		assert.Nil(t, err)

		c.getFileList("")
		filesInBucket, _ = c.getFileList("")
		assert.EqualValues(t, e.expectedStructureAfterDst2, filesInBucket)
		cleanUp(c)
	}
}
