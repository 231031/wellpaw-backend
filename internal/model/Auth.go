package model

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt"
)

type Argon2Configuration struct {
	HashRaw    []byte
	Salt       []byte
	TimeCost   uint32
	MemoryCost uint32
	Threads    uint8
	KeyLength  uint32
}

type TokenConfig struct {
	PrivateKey                *rsa.PrivateKey
	PublicKey                 *rsa.PublicKey
	RefreshSecret             string
	AccessTokenExpirationSecs int64
	RefreshExpirationSecs     int64
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserAuth struct {
	ID   uint     `json:"id"`
	Tier TierType `json:"tier"`
}

type TokenClaims struct {
	User UserAuth `json:"user"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	ID uint `json:"id"`
	jwt.StandardClaims
}

type RefreshTokenData struct {
	SS        string
	ID        string
	ExpiresIn time.Duration
}
