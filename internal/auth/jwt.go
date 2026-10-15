package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *TokenManager) Generate(userId uuid.UUID) (token string, expiredAt time.Time, err error) {
	now := time.Now()
	expiredAt = now.Add(m.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   userId.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiredAt),
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return token, expiredAt, err
}

func (m *TokenManager) Parse(tokenString string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		// Wajib cek algoritma, supaya token dengan alg lain (mis. "none") ditolak.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritma tidak didukung: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("subject token tidak valid")
	}
	return id, nil
}
