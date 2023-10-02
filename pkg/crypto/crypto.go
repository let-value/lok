package crypto

import (
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

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

func getPredeterminedNonce(cipher cipher.AEAD) string {
	nonceSize := cipher.NonceSize()
	nonceBase := "fixed_nonce_"
	repeatedNonce := strings.Repeat(nonceBase, (nonceSize/len(nonceBase))+1)
	return repeatedNonce[:nonceSize]
}

var dummyCipher, err = CreateCipher("dummy password")
var predeterminedNonce = getPredeterminedNonce(dummyCipher)

func EncryptBytesWeak(data []byte, cipher cipher.AEAD) ([]byte, error) {
	if len(predeterminedNonce) != cipher.NonceSize() {
		return nil, fmt.Errorf("predetermined nonce size is incorrect")
	}
	return cipher.Seal(nil, []byte(predeterminedNonce), data, nil), nil
}

func DecryptBytesWeak(data []byte, cipher cipher.AEAD) ([]byte, error) {
	if len(predeterminedNonce) != cipher.NonceSize() {
		return nil, fmt.Errorf("predetermined nonce size is incorrect")
	}
	return cipher.Open(nil, []byte(predeterminedNonce), data, nil)
}

func EncryptString(inputString string, cipher cipher.AEAD) (string, error) {
	encrypted, err := EncryptBytesWeak([]byte(inputString), cipher)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(encrypted), nil
}

func DecryptString(encryptedString string, cipher cipher.AEAD) (string, error) {
	data, err := base64.URLEncoding.DecodeString(encryptedString)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex string: %v", err)
	}

	decrypted, err := DecryptBytesWeak(data, cipher)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}

	return string(decrypted), nil
}

func EncryptBytes(data []byte, nonceData string, cipher cipher.AEAD) ([]byte, error) {
	nonce := DeriveNonce([]byte(nonceData), cipher)
	return cipher.Seal(nil, nonce, data, nil), nil
}

func DecryptBytes(data []byte, nonceData string, cipher cipher.AEAD) ([]byte, error) {
	nonce := DeriveNonce([]byte(nonceData), cipher)
	return cipher.Open(nil, nonce, data, nil)
}
