package service

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/argon2"
)

type TokenService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password string, hashedPassword string) (bool, error)

	GenerateNewPairToken(ctx context.Context, userAuth *model.UserAuth, prevToken string) (*model.TokenPair, error)
	generateAccessToken(user *model.UserAuth, key *rsa.PrivateKey, exp int64) (string, error)
	generateRefreshToken(id uint, key string, exp int64) (*model.RefreshTokenData, error)
	ValidateRefreshToken(refreshTokenStr string) (*model.RefreshTokenClaims, error)
	HandleRefreshToken(ctx context.Context, refreshToken string) (*model.TokenPair, error)
	ValidateToken(tokenStr string) (*model.TokenClaims, error)
}

type tokenService struct {
	tokenRepo                 repository.TokenRepository
	userRepo                  repository.UserRepository
	PrivateKey                *rsa.PrivateKey
	PublicKey                 *rsa.PublicKey
	RefreshSecret             string
	AccessTokenExpirationSecs int64
	RefreshExpirationSecs     int64
}

func NewTokenService(tokenRepo repository.TokenRepository, userRepo repository.UserRepository, cfg *model.TokenConfig) TokenService {
	return &tokenService{
		tokenRepo:                 tokenRepo,
		userRepo:                  userRepo,
		PrivateKey:                cfg.PrivateKey,
		PublicKey:                 cfg.PublicKey,
		RefreshSecret:             cfg.RefreshSecret,
		AccessTokenExpirationSecs: cfg.AccessTokenExpirationSecs,
		RefreshExpirationSecs:     cfg.RefreshExpirationSecs,
	}
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

func (s *tokenService) GenerateNewPairToken(ctx context.Context, userAuth *model.UserAuth, prevToken string) (*model.TokenPair, error) {
	if prevToken != "" {
		key := fmt.Sprintf("refresh_token:%s", prevToken)
		value, err := s.tokenRepo.GetandDelRefreshToken(ctx, key)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil, utils.ErrUnauth
			}
			return nil, err
		}

		identical := strings.Split(value, ":")
		if len(identical) != 2 {
			return nil, utils.ErrUnauth
		}

		uint64Val, err := strconv.ParseUint(identical[1], 10, 64)
		if err != nil {
			applogger.LogError(fmt.Sprintln("failed to convert string to uint", err), serviceLog)
			return nil, err
		}

		user, err := s.userRepo.GetUserByID(ctx, uint(uint64Val))
		if err != nil {
			return nil, utils.ErrUnauth
		}
		userAuth.ID = user.ID
		userAuth.Tier = user.PaymentPlan
	}

	// generate new token - login, refresh token
	newToken, err := s.generateAccessToken(userAuth, s.PrivateKey, s.AccessTokenExpirationSecs)
	if err != nil {
		return nil, err
	}

	newRefresh, err := s.generateRefreshToken(userAuth.ID, s.RefreshSecret, s.RefreshExpirationSecs)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("refresh_token:%s", newRefresh.SS)
	val := fmt.Sprintf("%s:%s", strconv.Itoa(int(userAuth.Tier)), strconv.Itoa(int(userAuth.ID)))
	err = s.tokenRepo.SetRefreshToken(ctx, key, val, newRefresh.ExpiresIn)
	if err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  newToken,
		RefreshToken: newRefresh.SS,
	}, nil
}

func (s *tokenService) generateAccessToken(user *model.UserAuth, key *rsa.PrivateKey, exp int64) (string, error) {
	curTime := time.Now()
	tokenExp := curTime.Unix() + exp

	claims := model.TokenClaims{
		User: *user,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  curTime.Unix(),
			ExpiresAt: tokenExp,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to signed string of token", err), serviceLog)
		return "", err
	}

	return ss, nil
}

func (s *tokenService) generateRefreshToken(id uint, key string, exp int64) (*model.RefreshTokenData, error) {
	curTime := time.Now()
	tokenExp := curTime.Add(time.Duration(exp) * time.Second)
	tokenID, err := uuid.NewRandom()
	if err != nil {
		applogger.LogError(fmt.Sprintln("falied to generate uuid", err), serviceLog)
		return nil, err
	}

	claims := model.RefreshTokenClaims{
		ID: id,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  curTime.Unix(),
			ExpiresAt: tokenExp.Unix(),
			Id:        tokenID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(key))
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to signed string of refresh token", err), serviceLog)
		return nil, err
	}

	return &model.RefreshTokenData{
		SS:        ss,
		ID:        tokenID.String(),
		ExpiresIn: tokenExp.Sub(curTime),
	}, nil
}

func (s *tokenService) ValidateRefreshToken(refreshTokenStr string) (*model.RefreshTokenClaims, error) {
	claims := &model.RefreshTokenClaims{}
	refreshToken, err := jwt.ParseWithClaims(refreshTokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.RefreshSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if !refreshToken.Valid {
		return nil, utils.ErrUnauth
	}

	claims, ok := refreshToken.Claims.(*model.RefreshTokenClaims)
	if !ok {
		return nil, utils.ErrUnauth
	}

	return claims, nil
}

func (s *tokenService) HandleRefreshToken(ctx context.Context, refreshToken string) (*model.TokenPair, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	_, err = uuid.Parse(claims.Id)
	if err != nil {
		return nil, err
	}

	userAuth := &model.UserAuth{}
	tokenPair, err := s.GenerateNewPairToken(ctx, userAuth, refreshToken)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *tokenService) ValidateToken(tokenStr string) (*model.TokenClaims, error) {
	claims := &model.TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return s.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, utils.ErrUnauthHeader
	}

	claims, ok := token.Claims.(*model.TokenClaims)
	if !ok {
		return nil, utils.ErrUnauthHeader
	}

	return claims, nil
}
