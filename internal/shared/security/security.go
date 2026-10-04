package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Algorithm is the identifier stored alongside the hash to allow future
	// migration between hashing algorithms without invalidating accounts.
	Algorithm = "argon2id"

	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	saltLength    = 16
)

var ErrInvalidHash = errors.New("invalid hash format")
var ErrMismatch = errors.New("password mismatch")

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash), nil
}

func VerifyPassword(password, encodedHash string) error {
	if strings.HasPrefix(encodedHash, "$argon2id$") {
		return verifyArgon2(password, encodedHash)
	}
	return verifyBcrypt(password, encodedHash)
}

func verifyArgon2(password, encodedHash string) error {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return ErrInvalidHash
	}

	var memory, time uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return ErrInvalidHash
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(hash)))

	if subtle.ConstantTimeCompare(hash, computedHash) == 1 {
		return nil
	}
	return ErrMismatch
}

func verifyBcrypt(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func GenerateAPIKey() (string, error) {
	return GenerateToken(32)
}
