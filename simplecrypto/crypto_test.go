package simplecrypto

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"testing"
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

	os.Remove(dataFile.Name())
	os.Remove(plainTextFile)
}

func TestCalculateHMAC(t *testing.T) {
	t.Parallel()
	dataFile, _ := ioutil.TempFile(os.TempDir(), "test")
	dataFile.Write([]byte("payload1234"))
	dataFile.Close()

	HMACAsHex := hex.EncodeToString(CalculateHMAC([]byte("foobar"), []byte("longtestiv123456"), dataFile.Name(), false))

	assert.Equal(t, HMACAsHex, "9f705902feb423df9216dc6e97ca8e30ce7edad53ac10dd84a34377efcc24dcc")
}

func TestGetKeyFromPassphrase(t *testing.T) {
	t.Parallel()
	key, _ := GetKeyFromPassphrase([]byte("password"), []byte("salt"))

	keyAsByte, _ := hex.DecodeString("ecdaab8f7ea0ea6f4b9f4e930cef2a1bb277736f64c971c43ca5d73cfb4bb80ff4fdba62d7a91d9d6ede70f8f7be28d8e161aa1b6962a7e577ffb9aa74a934ed")

	actualKey := &Keys{keyAsByte[:32], keyAsByte[32:64]}
	assert.Equal(t, key, actualKey)
}
