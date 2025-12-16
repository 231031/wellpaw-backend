package service

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/231031/pethealth-backend/internal/applogger"
	"github.com/231031/pethealth-backend/internal/model"
	"github.com/231031/pethealth-backend/internal/utils"
	"golang.org/x/crypto/argon2"
)

type TokenService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password string, hashedPassword string) (bool, error)
	GenerateToken(user *model.User) (string, error)
	ValidateToken(token string) (*model.User, error)
}

type tokenService struct {
	// Add any dependencies like a token repository or secret keys if needed
}

func NewTokenService() TokenService {
	return &tokenService{}
}

func (s *tokenService) HashPassword(password string) (string, error) {
	cfg := &model.Argon2Configuration{
		TimeCost:   2,
		MemoryCost: 64 * 1024,
		Threads:    4,
		KeyLength:  32,
	}

	salt, err := utils.GenerateSalt(cfg.KeyLength)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to generate salt:", err), serviceLog)
		return "", err
	}
	cfg.Salt = salt

	hash := argon2.IDKey([]byte(password), cfg.Salt, cfg.TimeCost, cfg.MemoryCost, cfg.Threads, cfg.KeyLength)
	cfg.HashRaw = hash

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		cfg.MemoryCost,
		cfg.TimeCost,
		cfg.Threads,
		base64.RawStdEncoding.EncodeToString(cfg.Salt),
		base64.RawStdEncoding.EncodeToString(cfg.HashRaw),
	)

	return encodedHash, nil
}

func (s *tokenService) VerifyPassword(password, hashedPassword string) (bool, error) {
	// Parse stored hash parameters
	config, err := utils.ParseArgon2Hash(hashedPassword)
	if err != nil {
		applogger.LogError(fmt.Sprintln("error parsing stored hash:", err), serviceLog)
		return false, err
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		config.Salt,
		config.TimeCost,
		config.MemoryCost,
		config.Threads,
		config.KeyLength,
	)

	match := subtle.ConstantTimeCompare(config.HashRaw, computedHash) == 1
	return match, nil
}

func (s *tokenService) GenerateToken(user *model.User) (string, error) {
	// Implement token generation logic here
	return "", nil
}

func (s *tokenService) ValidateToken(token string) (*model.User, error) {
	// Implement token validation logic here
	return nil, nil
}
