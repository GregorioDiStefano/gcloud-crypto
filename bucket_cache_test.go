package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddFile(t *testing.T) {
	filesList := map[string]string{"a": "filea", "b": "fileb", "c": "filec", "d": "filed", "e": "filee"}
	finalFilesList := map[string]string{"a": "filea", "b": "fileb", "c": "filec", "d": "filed", "e": "filee", "f": "testfile"}

	sf := bucketCache{seenFiles: filesList}
	sf.addFile("f", "testfile")
	sf.addFile("a", "filea")
	sf.addFile("b", "fileb")
	assert.Equal(t, len(finalFilesList), len(sf.seenFiles))

	for k, _ := range finalFilesList {
		assert.Equal(t, sf.seenFiles[k], finalFilesList[k])
	}

}

func TestRemoveFile(t *testing.T) {
	filesList := map[string]string{"a": "filea", "b": "fileb", "c": "filec", "d": "filed", "e": "filee"}
	finalFilesList := map[string]string{"a": "filea", "b": "fileb"}

	sf := bucketCache{seenFiles: filesList}
	sf.removeFile("filec")
	sf.removeFile("filed")
	sf.removeFile("filee")
	sf.removeFile("testa")

	assert.Equal(t, len(finalFilesList), len(sf.seenFiles))

	for k, _ := range finalFilesList {
		assert.Equal(t, sf.seenFiles[k], finalFilesList[k])
	}
}
