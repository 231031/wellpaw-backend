package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/repository"
)

const (
	defaultOTPLength = 6
	defaultOTPTTL    = 5 * time.Minute
)

var (
	ErrOTPExpired = errors.New("otp expired")
	ErrInvalidOTP = errors.New("invalid otp")
)

type OTPService interface {
	SendOTP(ctx context.Context, email string) error
	ValidateOTP(ctx context.Context, email string, otp string) error
}

type otpService struct {
	otpRepo      repository.OTPRepository
	emailService EmailService
	otpTTL       time.Duration
}

func NewOTPService(otpRepo repository.OTPRepository, emailService EmailService) OTPService {
	return &otpService{
		otpRepo:      otpRepo,
		emailService: emailService,
		otpTTL:       defaultOTPTTL,
	}
}

func (s *otpService) SendOTP(ctx context.Context, email string) error {
	normalizedEmail := normalizeEmail(email)

	otp, err := generateOTP(defaultOTPLength)
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to generate otp: %s", err.Error()), serviceLog)
		return fmt.Errorf("failed to generate otp: %w", err)
	}

	if err := s.otpRepo.SetOTP(ctx, normalizedEmail, otp, s.otpTTL); err != nil {
		return fmt.Errorf("failed to store otp: %w", err)
	}

	if err := s.emailService.SendOTPEmail(ctx, normalizedEmail, otp); err != nil {
		if deleteErr := s.otpRepo.DeleteOTP(ctx, normalizedEmail); deleteErr != nil && !errors.Is(deleteErr, repository.ErrOTPNotFound) {
			applogger.LogError(fmt.Sprintf("failed to cleanup otp after email error: %s", deleteErr.Error()), serviceLog)
		}

		return fmt.Errorf("failed to send otp: %w", err)
	}

	return nil
}

func (s *otpService) ValidateOTP(ctx context.Context, email string, otp string) error {
	normalizedEmail := normalizeEmail(email)

	normalizedOTP := strings.TrimSpace(otp)
	if !isValidOTP(normalizedOTP) {
		return ErrInvalidOTP
	}

	storedOTP, err := s.otpRepo.GetOTP(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repository.ErrOTPNotFound) {
			return ErrOTPExpired
		}

		return fmt.Errorf("failed to validate otp: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(storedOTP), []byte(normalizedOTP)) != 1 {
		return ErrInvalidOTP
	}

	if err := s.otpRepo.DeleteOTP(ctx, normalizedEmail); err != nil {
		return fmt.Errorf("failed to complete otp validation: %w", err)
	}

	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidOTP(otp string) bool {
	if len(otp) != defaultOTPLength {
		return false
	}

	for _, char := range otp {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func generateOTP(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid otp length")
	}

	maxNumber := big.NewInt(1)
	for i := 0; i < length; i++ {
		maxNumber = maxNumber.Mul(maxNumber, big.NewInt(10))
	}

	number, err := rand.Int(rand.Reader, maxNumber)
	if err != nil {
		return "", fmt.Errorf("failed to generate random otp: %w", err)
	}

	return fmt.Sprintf("%0*d", length, number.Int64()), nil
}
