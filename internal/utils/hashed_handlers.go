package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/231031/pethealth-backend/internal/applogger"
	"github.com/231031/pethealth-backend/internal/model"
)

var (
	utilsLog = "[UTILS LOGGER]"
)

func ParseArgon2Hash(encodedHash string) (*model.Argon2Configuration, error) {
	components := strings.Split(encodedHash, "$")
	if len(components) != 6 {
		return nil, errors.New("invalid hash format structure")
	}

	if !strings.HasPrefix(components[1], "argon2id") {
		return nil, errors.New("unsupported algorithm variant")
	}

	var version int
	fmt.Sscanf(components[2], "v=%d", &version)

	config := &model.Argon2Configuration{}
	fmt.Sscanf(components[3], "m=%d,t=%d,p=%d",
		&config.MemoryCost, &config.TimeCost, &config.Threads)

	// Decode salt component
	salt, err := base64.RawStdEncoding.DecodeString(components[4])
	if err != nil {
		applogger.LogError(fmt.Sprintln("error decoding salt:", err), utilsLog)
		return nil, err
	}
	config.Salt = salt

	// Decode hash component
	hash, err := base64.RawStdEncoding.DecodeString(components[5])
	if err != nil {
		applogger.LogError(fmt.Sprintln("error decoding hash:", err), utilsLog)
		return nil, err
	}
	config.HashRaw = hash
	config.KeyLength = uint32(len(hash))

	return config, nil
}

// generateSalt creates a cryptographically secure random salt
func GenerateSalt(saltLength uint32) ([]byte, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
