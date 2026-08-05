package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

// deriveAESKey 将任意长度的密钥材料派生为 32 字节的 AES-256 密钥
func deriveAESKey(material []byte) []byte {
	hash := sha256.Sum256(material)
	return hash[:] // 返回 32 字节，完美适配 AES-256
}

// AesEncryptGCM 使用 AES-GCM 加密数据
// 返回格式: [nonce][ciphertext][tag]
// key 长度必须为 16/24/32 字节
func AesEncryptGCM(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(deriveAESKey(key))
	if err != nil {
		return nil, fmt.Errorf("aes new cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm init failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce failed: %w", err)
	}

	// Seal 将 nonce 作为 dst 前缀，输出 = nonce + ciphertext + tag
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// AesDecryptGCM 解密 AES-GCM 密文并验证完整性
// 输入格式: [nonce][ciphertext][tag]（与 Encrypt 输出对应）
// key 长度必须为 16/24/32 字节
func AesDecryptGCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes new cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm init failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	minLen := nonceSize + gcm.Overhead() // overhead = tag size (16 bytes)
	if len(ciphertext) < minLen {
		return nil, fmt.Errorf("ciphertext too short: got %d, need at least %d", len(ciphertext), minLen)
	}

	nonce, sealedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, sealedData, nil)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm decrypt failed (tampered or wrong key): %w", err)
	}

	return plaintext, nil
}
