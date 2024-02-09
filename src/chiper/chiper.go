package chiper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
)

type Chiper struct {
	Key   string
	block cipher.Block
}

func NewChiper(key string) (*Chiper, error) {
	// You may want to handle key validation and generation more securely in a real-world scenario
	if len(key) != 32 {
		log.Println("Key not valid need a 32 char...")
		return nil, fmt.Errorf("Not valid key. Need 32 chars...")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	return &Chiper{block: block}, nil
}

func (c *Chiper) Encrypt(data []byte) ([]byte, error) {

	ciphertext := make([]byte, aes.BlockSize+len(data))

	// Initialization Vector (IV) - should be unique, but not necessarily secret
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(c.block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

func (c *Chiper) Decrypt(encryptedData []byte) ([]byte, error) {

	if len(encryptedData) < aes.BlockSize {
		return nil, errors.New("ciphertext is too short")
	}

	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(c.block, iv)
	stream.XORKeyStream(encryptedData, encryptedData)

	return encryptedData, nil
}
