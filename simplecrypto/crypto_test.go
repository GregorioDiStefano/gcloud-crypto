package simplecrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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

func TestEncryptDecryptText_Large_File(t *testing.T) {
	t.Parallel()
	key := randomByte(16)

	plainText := strings.Repeat("this is a string 1234", 1024*1024*5)

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

	const tempFile = "/tmp/plaintext-file"
	encryptDecryptTests := []struct {
		key      []byte
		iv       []byte
		filepath string

		expectedError error
	}{
		{[]byte("foobar"), []byte("longtestiv123456"), "test_data/test-encrypt_decrypt_1", nil},
		{[]byte("foobar"), []byte("longtestiv123456"), "test_data/test-encrypt_decrypt_2", nil},
		{[]byte("foobar"), []byte("longtestiv123456"), "test_data/test-encrypt_decrypt_3", nil},
		{[]byte("foobar"), []byte("longtestiv123456"), "test_data/test-encrypt_decrypt_4", nil},
		{[]byte("foobar"), []byte("longtestiv123456"), "test_data/404", errors.New(unableToOpenFileReading)},
	}

	for _, e := range encryptDecryptTests {
		key, _ := GetKeyFromPassphrase(e.key, e.iv)

		dataToEncrypt, _ := ioutil.ReadFile(e.filepath)
		encryptedFilename, err := EncryptFile(e.filepath, *key)

		if e.expectedError != nil {
			assert.Equal(t, e.expectedError, err)
			return
		}

		encryptedDataBytes, err := ioutil.ReadFile(encryptedFilename)

		if err != nil {
			panic(err)
		}

		assert.True(t, len(encryptedDataBytes) == len(dataToEncrypt)+aes.BlockSize+sha256.Size, "Looks like the IV/HMAC is missing from the file?")

		assert.NotEqual(t, encryptedDataBytes, dataToEncrypt)

		plainTextFile, err := DecryptFile(encryptedFilename, *key)
		defer os.Remove(plainTextFile)

		if err != nil {
			panic(err)
		}

		decryptedDataBytes, err := ioutil.ReadFile(plainTextFile)

		if err != nil {
			panic(err)
		}

		assert.Equal(t, decryptedDataBytes, dataToEncrypt)
	}

}

func TestCalculateHMAC(t *testing.T) {
	t.Parallel()

	calculateHMACTests := []struct {
		key      []byte
		iv       []byte
		filepath string

		expectedHMACHex string
		expectedError   error
	}{
		{[]byte("foobar"), []byte("longtestiv123456"), "test_data/test-calculate_hmac-1", "de48a455255028af130a99726a3a30144aec11cb80713cf67210d851af26774f", nil},
		{[]byte("foobar"), []byte("longtestiv123456"), "test_data/test-calculate_hmac-2", "de48a455255028af130a99726a3a30144aec11cb80713cf67210d851af26774f", nil},
	}

	for _, h := range calculateHMACTests {
		fh, err := os.Open(h.filepath)

		HMAC, err := calculateHMAC(h.key, h.iv, *fh)

		if err == nil {
			assert.True(t, hex.EncodeToString(HMAC) == h.expectedHMACHex, "HMAC calculation failed")
		} else {
			assert.Equal(t, h.expectedError, err, "Unexpected error")
		}

	}
}

func TestGetKeyFromPassphrase(t *testing.T) {
	t.Parallel()
	key, _ := GetKeyFromPassphrase([]byte("password"), []byte("salt1234"))
	keyAsByte, _ := hex.DecodeString("881532abb8344b9f6720bf8cba43ec1c5ccd717bf5ca9b1461fb8afeb832aa6473c7d97ef9f6c203cd15763884b75b958347bffbe28bbee351c09818a23c4632")
	actualKey := &Keys{keyAsByte[:32], keyAsByte[32:64]}
	assert.Equal(t, key, actualKey)
}

func TestGetKeyFromPassphraseError_1(t *testing.T) {
	t.Parallel()

	_, err := GetKeyFromPassphrase([]byte(""), []byte("abc"))
	assert.Error(t, err, "No error returned")
}

func TestGetKeyFromPassphraseError_2(t *testing.T) {
	t.Parallel()

	_, err := GetKeyFromPassphrase(nil, []byte("abcd1234"))
	assert.Error(t, err, "No error returned")
}

/*
func TestGetHMACFromFile(t *testing.T) {
	t.Parallel()

	getHMACTests := []struct {
		file string
		hmac []byte
		err  error
	}{
		{"test_data/test-get_hmac-1", bytes.Repeat([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, 4), nil},
		{"test_data/test-get_hmac-2", bytes.Repeat([]byte{0xFF}, 32), nil},
		{"test_data/test-get_hmac-3", bytes.Repeat([]byte{0xFF}, 32), nil},
		{"test_data/test-get_hmac-4", nil, errors.New(errorReadingHMAC)},
		{"test_data/404", nil, errors.New(unableToOpenFileReading)},
	}

	for _, h := range getHMACTests {
		fh, err := os.Open(h.file)
		if err != nil {
			panic(err)
		}
		hmac, e := getHMACFromFile(*fh)

		if h.err != nil {
			assert.Equal(t, h.err, e, "Unexpected error returned")
		}

		if h.hmac != nil {
			assert.Equal(t, h.hmac, hmac, "Failed to read HMAC from file")
		}
	}
}
*/
func TestAddHMACToFile(t *testing.T) {
	t.Parallel()

	addHMACTests := []struct {
		file string
		hmac []byte
		err  error
	}{
		{"test_data/test-add_hmac-1", bytes.Repeat([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, 4), nil},
		{"test_data/test-add_hmac-2", bytes.Repeat([]byte{0x07}, 4), nil},
		{"test_data/404", bytes.Repeat([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, 4), errors.New(unableToOpenFileWriting)},
	}

	for _, h := range addHMACTests {

		if len(h.hmac) < 32 {
			assert.Panics(t, func() { addHMACToFile(h.file, h.hmac) }, "Failed to panic")
		} else {
			e := addHMACToFile(h.file, h.hmac)

			if h.err != nil {
				assert.Equal(t, h.err, e, "Unexpected error returned")
			}
		}
	}
}

func TestTruncateHMACSignature(t *testing.T) {
	t.Parallel()

	truncateHMACTests := []struct {
		filepath string
		err      error

		expectedHMAC string
	}{
		{"test_data/test-truncate_hmac-1", nil, "9312d9e6a4dd7571481298141165206a4de52cca5c7c01ba3b7de0122e147d08"},
		{"test_data/test-truncate_hmac-2", errors.New(errorReadingHMAC), ""},
	}

	for _, h := range truncateHMACTests {
		exec.Command("cp", h.filepath, "/tmp/truncate-file").Run()
		fh, err := os.Open("/tmp/truncate-file")

		if err != nil {
			panic(err)
		}

		hmac, e := truncateHMACSignature(fh)

		if h.expectedHMAC != "" {
			assert.Equal(t, h.expectedHMAC, hex.EncodeToString(hmac))
		}

		if h.err != nil {
			fmt.Println(h.err, e)
			assert.Equal(t, h.err, e, "Unexpected error returned")
		}
	}
}
