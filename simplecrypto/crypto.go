package simplecrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/scrypt"

	"io"
	"os"
)

type Keys struct {
	EncryptionKey []byte
	HmacKey       []byte
}

func randomBytes(length int) []byte {
	rb := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, rb); err != nil {
		panic(err)
	}

	return rb
}

func GenerateRandomIV() []byte {
	return randomBytes(aes.BlockSize)
}

func GetKeyFromPassphrase(passphrase, salt []byte) *Keys {
	if passphrase == nil || salt == nil {
		panic("missing passphrase or salt required to generate crypto keys")
	}

	k, err := scrypt.Key([]byte(passphrase), salt, 16384, 16, 1, 64)

	if err != nil {
		panic(err)
	}

	dk := k[:32]
	mk := k[32:64]

	return &Keys{dk, mk}
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

func EncryptFile(filename string, key []byte) (string, []byte, error) {
	outputFilename := fmt.Sprintf("%s.%s", filename, "enc")
	readFile, err := os.Open(filename)

	if err != nil {
		return "", nil, err
	}

	writeFile, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", nil, err
	}

	block, err := aes.NewCipher(key)

	if err != nil {
		return "", nil, err
	}

	defer readFile.Close()
	defer writeFile.Close()

	iv := GenerateRandomIV()

	writer := &cipher.StreamWriter{S: cipher.NewCTR(block, iv), W: writeFile}

	writeFile.Write(iv)

	if _, err := io.Copy(writer, readFile); err != nil {
		fmt.Println("error during crypto")
		return "", nil, err
	}

	return outputFilename, iv, nil
}

func DecryptFile(filename string, key []byte) (string, error) {
	iv := make([]byte, aes.BlockSize)

	readFile, err := os.Open(filename)

	cwd, _ := os.Getwd()
	decryptedFilename, _ := ioutil.TempFile(cwd, "plaintext")

	writeFile, err := os.OpenFile(decryptedFilename.Name(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		panic(err)
	}

	defer readFile.Close()
	defer writeFile.Close()

	if err != nil {
		panic(err)
	}

	readFile.Read(iv)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(block, iv)
	reader := &cipher.StreamReader{S: stream, R: readFile}

	// Copy the input file to the output file, decrypting as we go.
	if _, err := io.Copy(writeFile, reader); err != nil {
		return "", err
	}

	return decryptedFilename.Name(), nil
}

func GetIVFromEncryptedFile(filename string) []byte {
	readFile, err := os.Open(filename)
	iv := make([]byte, aes.BlockSize)

	if err != nil {
		panic(err)
	}

	readFile.Read(iv)
	return iv
}

// compute HMAC-SHA256 as: hmac(key, IV + cipherText)
func CalculateHMAC(key, iv []byte, filepath string, skipLast32bytes bool) []byte {
	f, err := os.Open(filepath)
	defer f.Close()

	const idealBufferSize = 16 * 1024

	if err != nil {
		panic(err)
	}

	fileStat, _ := f.Stat()
	fileSize := fileStat.Size()

	hash := hmac.New(sha256.New, key)
	hash.Write(iv)

	bytesCounted := int64(0)

	for {
		var data []byte

		//check if we are near the end of the file
		if bytesCounted+idealBufferSize+32 >= fileSize {
			data = make([]byte, 1)
		} else {
			//if we aren't near the end, read idealBufferSize bytes
			data = make([]byte, idealBufferSize)
		}

		actuallyRead, _ := f.Read(data)

		if actuallyRead == 0 {
			break
		}

		bytesCounted += int64(actuallyRead)

		if skipLast32bytes && fileSize-bytesCounted+1 == 32 {
			break
		} else {
			hash.Write(data)
		}
	}

	return hash.Sum(nil)
}
