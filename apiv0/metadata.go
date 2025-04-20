package apiv0

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"hash/adler32"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var apiMetadataKey = uuid.MustParse("0c2a5939-4c3c-4747-8f25-03751eccea76") // This is random UUID

type messageMetadata struct {
	TokenAdler32 uint32   `json:"token_adler32"`
	CallerAddr   string   `json:"caller_addr"`
	CallerAddrs  []string `json:"caller_addrs"`
}

type messageEncryptedMetadata struct {
	EncryptedMetadata string `json:"encrypted_caller_info"`
}

func storeEncryptedMetadataInCtx(c *fiber.Ctx, aesKey [32]byte) error {
	tokenString := extractBearerToken(c.Get(fiber.HeaderAuthorization, ""))
	hash := adler32.New()
	hash.Write([]byte(tokenString))
	metadata := messageMetadata{
		TokenAdler32: hash.Sum32(),
		CallerAddr:   c.IP(),
		CallerAddrs:  c.IPs(),
	}
	metadataBytes, err := json.Marshal(&metadata)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, metadataBytes, nil)
	encryptedMetadata := &messageEncryptedMetadata{
		EncryptedMetadata: base64.StdEncoding.EncodeToString(ciphertext),
	}
	c.Locals(apiMetadataKey, encryptedMetadata)
	return nil
}

func loadEncryptedMetadataFromCtx(c *fiber.Ctx) *messageEncryptedMetadata {
	return c.Locals(apiMetadataKey).(*messageEncryptedMetadata)
}
