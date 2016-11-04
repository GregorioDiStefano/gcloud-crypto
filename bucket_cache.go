package main

type bucketCache struct {
	seenFiles map[string]string
}

func (bc *bucketCache) addFile(encrypted, decrypted string) {

	if bc.seenFiles == nil {
		bc.seenFiles = make(map[string]string, 100)
	}

	for _, decryptedFilePath := range bc.seenFiles {
		if decryptedFilePath == decrypted {
			return
		}
	}

	bc.seenFiles[encrypted] = decrypted
}

func (bc *bucketCache) removeFile(decrypted string) {
	for encryptedFilePath, decryptedFilePath := range bc.seenFiles {
		if decryptedFilePath == decrypted {
			delete(bc.seenFiles, encryptedFilePath)
		}
	}
}

func (bc *bucketCache) empty() {
	bc.seenFiles = make(map[string]string, 100)
}
