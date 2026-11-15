package crypto

import (
	"bytes"
	"testing"
)

const (
	KeyA = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	KeyB = "1f1e1d1c1b1a191817161514131211100f0e0d0c0b0a09080706050403020100"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"one byte", []byte{0x45}},
		{"text", []byte("Hello, Server!")},
		{"zeros", make([]byte, 64)},
		{"mtu sized", bytes.Repeat([]byte{0xab}, 1420)},
	}

	aead := NewAEAD(KeyA)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datagram := Encrypt(aead, tt.plaintext)
			got, err := Decrypt(aead, datagram)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.plaintext) {
				t.Fatalf("got %x, want %x", got, tt.plaintext)
			}
		})
	}
}

func TestEncryptUniqueNonce(t *testing.T) {
	aead := NewAEAD(KeyA)
	plaintext := []byte("same packet")

	first := Encrypt(aead, plaintext)
	second := Encrypt(aead, plaintext)

	if bytes.Equal(first, second) {
		t.Fatal("two encryptions of the same packet are identical")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"text", []byte("Hello, Server!")},
		{"mtu sized", bytes.Repeat([]byte{0xab}, 1420)},
	}

	sender := NewAEAD(KeyA)
	receiver := NewAEAD(KeyB)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datagram := Encrypt(sender, tt.plaintext)
			if _, err := Decrypt(receiver, datagram); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDecryptTampered(t *testing.T) {
	aead := NewAEAD(KeyA)
	nonceSize := aead.NonceSize()

	tests := []struct {
		name   string
		tamper func(datagram []byte) []byte
	}{
		{"flip nonce byte", func(d []byte) []byte {
			d[0] ^= 0x01
			return d
		}},
		{"flip ciphertext byte", func(d []byte) []byte {
			d[nonceSize] ^= 0x01
			return d
		}},
		{"flip tag byte", func(d []byte) []byte {
			d[len(d)-1] ^= 0x01
			return d
		}},
		{"truncate tag", func(d []byte) []byte {
			return d[:len(d)-1]
		}},
		{"append byte", func(d []byte) []byte {
			return append(d, 0x00)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datagram := Encrypt(aead, []byte("Hello, Server!"))
			datagram = tt.tamper(datagram)
			if _, err := Decrypt(aead, datagram); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
