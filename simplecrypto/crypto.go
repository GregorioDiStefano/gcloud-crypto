package simplecrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

var log = logrus.New()

type Keys struct {
	EncryptionKey []byte
	HMACKey       []byte
}

const (
	unableToOpenFileReading = "Unable to open file for reading"
	unableToOpenFileWriting = "Unable to open file for writing"
	noSaltOrPassword        = "No salt or password provided"
	saltTooSmall            = "Salt is too small, 64 bits needed"
	notEncrypted            = "Ciphertext too small to be encrypted"

	hmacValidationFailed = "HMAC validation failed"

	errorReadingIV   = "Unable to read IV from file"
	errorReadingHMAC = "Unable to extract HMAC from file"
)

func init() {
	log.Level = logrus.DebugLevel
}

func randomBytes(length int) []byte {
	rb := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, rb); err != nil {
		panic(err)
	}

	return rb
}

func generateRandomIV() []byte {
	return randomBytes(aes.BlockSize)
}

func GetKeyFromPassphrase(passphrase, salt []byte, N, r, p int) (*Keys, error) {
	if passphrase == nil || salt == nil {
		panic(noSaltOrPassword)
	}

	if len(salt) < 8 {
		panic(saltTooSmall)
	}

	k, err := scrypt.Key([]byte(passphrase), salt, N, r, p, 64)

	if err != nil {
		return nil, err
	}

	dk := k[:32]
	mk := k[32:64]

	return &Keys{dk, mk}, nil
}

func EncryptText(text string, key []byte) (string, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	nonce := randomBytes(12)

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	ciphertext = append(nonce, ciphertext...)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptText(cryptoText string, key []byte) (string, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	if len(ciphertext) < 12 {
		return "", errors.New(notEncrypted)
	}

	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", plaintext), nil
}

/*
// Nice for debugging, so leaving here
func (pt passThrough) Write(p []byte) (n int, err error) {
	fmt.Println("bytes: ", hex.EncodeToString(p), "\n")
	return len(p), nil
}

type passThrough struct {
	io.Writer
}
*/

func EncryptFile(filename string, keys *Keys) (string, []byte, error) {
	outputFilename := fmt.Sprintf("%s.%s", filename, "enc")
	readFile, err := os.Open(filename)
	md5Hash := md5.New()

	defer readFile.Close()

	if err != nil {
		log.Fatalf("error opening: %s, err: %s", filename, err.Error())
		return "", nil, errors.New(unableToOpenFileReading)
	}

	writeFile, err := os.OpenFile(outputFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return "", nil, err
	}
	defer writeFile.Close()

	block, err := aes.NewCipher(keys.EncryptionKey)

	if err != nil {
		return "", nil, err
	}

	iv := generateRandomIV()
	md5Hash.Write(iv)

	writer := &cipher.StreamWriter{S: cipher.NewCTR(block, iv), W: io.MultiWriter(writeFile, md5Hash)}

	writeFile.Write(iv)

	if _, err := io.Copy(writer, readFile); err != nil {
		log.Fatal("error during crypto: " + err.Error())
		return "", nil, err
	}

	writeFile.Sync()

	if hmac, err := calculateHMAC(keys.HMACKey, iv, writeFile); err == nil {
		addHMACToFile(writeFile, hmac)
		md5Hash.Write(hmac)
	} else {
		return "", nil, err
	}

	return outputFilename, md5Hash.Sum(nil), nil
}

func DecryptFile(filename string, keys *Keys) (string, error) {
	iv := make([]byte, aes.BlockSize)
	readFile, err := os.Open(filename)

	cwd, _ := os.Getwd()
	decryptedFilename, _ := ioutil.TempFile(cwd, "plaintext")

	writeFile, err := os.OpenFile(decryptedFilename.Name(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return "", err
	}

	defer readFile.Close()
	defer writeFile.Close()

	readFile.Read(iv)

	block, err := aes.NewCipher(keys.EncryptionKey)

	if err != nil {
		return "", err
	}

	stream := cipher.NewCTR(block, iv)
	reader := &cipher.StreamReader{S: stream, R: readFile}

	// Copy the input file to the output file, decrypting as we go.
	if _, err := io.Copy(writeFile, reader); err != nil {
		return "", err
	}

	writeFile.Sync()
	expectedHMAC, err := truncateHMACSignature(readFile)

	if err != nil {
		return "", err
	}

	if actualHMAC, err := calculateHMAC(keys.HMACKey, iv, readFile); err != nil || !bytes.Equal(actualHMAC, expectedHMAC) {
		log.Fatal("Failed to validate HMAC")
		return "", errors.New(hmacValidationFailed)
	}

	if _, err = truncateHMACSignature(writeFile); err != nil {
		return "", err
	}

	return decryptedFilename.Name(), nil

}

// compute HMAC-SHA256 as: hmac(key, IV + cipherText)
func calculateHMAC(key, iv []byte, fh *os.File) ([]byte, error) {
	const idealBufferSize = 16 * 1024
	fh.Seek(0, 0)

	hash := hmac.New(sha256.New, key)
	hash.Write(iv)

	for {
		data := make([]byte, idealBufferSize)
		actuallyRead, err := fh.Read(data)

		data = data[0:actuallyRead]

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		fh.Sync()
		hash.Write(data)
	}

	return hash.Sum(nil), nil
}

func truncateHMACSignature(file *os.File) ([]byte, error) {
	extractedHMAC := make([]byte, sha256.Size)

	if stat, err := file.Stat(); err != nil || stat.Size() < sha256.Size {
		return nil, errors.New(errorReadingHMAC)
	} else {
		if _, err := file.ReadAt(extractedHMAC, stat.Size()-sha256.Size); err != nil {
			return nil, err
		}
		if err := os.Truncate(file.Name(), stat.Size()-sha256.Size); err != nil {
			return nil, err
		}
	}
	return extractedHMAC, nil
}

func addHMACToFile(file *os.File, hmac []byte) error {
	if len(hmac) != sha256.Size {
		panic("Adding HMAC of incorrect size, this should never happen")
	}

	file.Write(hmac)
	return nil
}
