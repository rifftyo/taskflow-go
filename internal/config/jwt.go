package config

import (
	"errors"
	"os"
	"time"
)

type JWTConfig struct {
	Secret    []byte
	ExpiresIn time.Duration
}

func LoadJWTConfig() (*JWTConfig, error) {
	secret := os.Getenv("JWT_SECRET")

	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 bytes")
	}

	expiresIn, err := time.ParseDuration(
		os.Getenv("JWT_EXPIRES_IN"),
	)

	if err != nil || expiresIn <= 0 {
		return nil, errors.New("invalid JWT expiration")
	}

	return &JWTConfig{
		Secret: []byte(secret),
		ExpiresIn: expiresIn,
	}, nil
}