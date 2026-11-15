package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"log"
)

func NewAEAD(hexKey string) cipher.AEAD {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		log.Fatal(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}
	return aead
}

func Encrypt(aead cipher.AEAD, plaintext []byte) []byte {
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(plaintext)+aead.Overhead())
	rand.Read(nonce)
	return aead.Seal(nonce, nonce, plaintext, nil)
}

func Decrypt(aead cipher.AEAD, datagram []byte) ([]byte, error) {
	nonceSize := aead.NonceSize()
	return aead.Open(nil, datagram[:nonceSize], datagram[nonceSize:], nil)
}
