package simplecrypto

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randomByte(length int) []byte {
	b := make([]byte, length)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}

	return b
}

func TestEncryptDecryptText_1(t *testing.T) {
	t.Parallel()
	key := randomByte(16)

	plainText := "this is a string 1234"

	et, err := EncryptText(plainText, key)

	if err != nil {
		t.Fatalf(err.Error())
	}

	dt, err := DecryptText(et, key)

	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, dt, plainText)
}

func TestEncryptDecryptText_2(t *testing.T) {
	t.Parallel()
	key := randomByte(16)

	plainText := "this is a string 1234"

	et, err := EncryptText(plainText, key)

	if err != nil {
		t.Fatalf(err.Error())
	}

	//change a byte in the middle of the encrypted text
	etBytes, err := base64.URLEncoding.DecodeString(et)
	middle := len(etBytes) / 2

	if etBytes[middle] < 0xFF {
		etBytes[middle]++
	} else {
		etBytes[middle]--
	}

	et = base64.URLEncoding.EncodeToString(etBytes)
	_, err = DecryptText(et, key)

	if err == nil {
		t.Fail()
	} else if err == errors.New("cipher: message authentication failed") {
		//pass
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	t.Parallel()

	key := randomByte(16)

	dataToEncrypt := randomByte(1024 * 512)
	dataFile, _ := ioutil.TempFile(os.TempDir(), "test")

	dataFile.Write(dataToEncrypt)
	dataFile.Close()

	encryptedFilename, _, _ := EncryptFile(dataFile.Name(), key)

	encryptedDataBytes, _ := ioutil.ReadFile(encryptedFilename)

	assert.True(t, len(encryptedDataBytes) == len(dataToEncrypt)+aes.BlockSize, "Looks like the IV is missing from the file?")

	assert.NotEqual(t, encryptedDataBytes, dataToEncrypt)

	plainTextFile, _ := DecryptFile(encryptedFilename, key)

	decryptedDataBytes, _ := ioutil.ReadFile(plainTextFile)

	assert.Equal(t, decryptedDataBytes, dataToEncrypt)

	dataFile.Close()

	os.Remove(dataFile.Name())
	os.Remove(plainTextFile)
}
