package crypto

import (
	"crypto/cipher"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

func DeriveKey(password string) []byte {
	const pepper = "pepper"
	return argon2.IDKey([]byte(password), []byte(pepper), 4, 64*1024, 4, 32)
}

func CreateCipher(password string) (cipher.AEAD, error) {
	derivedKey := DeriveKey(password)
	return chacha20poly1305.NewX(derivedKey)
}

func DeriveNonce(data []byte, cipher cipher.AEAD) []byte {
	hash := sha256.Sum256(data)
	return hash[:cipher.NonceSize()]
}

func EncryptString(data string, cipher cipher.AEAD) (string, error) {
	nonce := DeriveNonce([]byte(data), cipher)
	encrypted := cipher.Seal(nil, nonce, []byte(data), nil)
	return fmt.Sprintf("%x", encrypted), nil
}
