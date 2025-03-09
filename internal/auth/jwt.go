package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/marchelhutagalung/go-service/internal/config"
	"github.com/marchelhutagalung/go-service/internal/database"
	"strconv"
	"time"
)

// Common errors
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTService struct {
	config      *config.JWTConfig
	redisClient *database.RedisClient
}

func NewJWTService(config *config.JWTConfig, redisClient *database.RedisClient) *JWTService {
	return &JWTService{
		config:      config,
		redisClient: redisClient,
	}
}

// GenerateToken creates a new JWT token for a user
func (s *JWTService) GenerateToken(userID int64) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.config.Expiration)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", err
	}

	tokenKey := fmt.Sprintf("token:valid:%d", userID)
	ctx := context.Background()
	err = s.redisClient.Set(ctx, tokenKey, tokenString, s.config.Expiration)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates the JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	ctx := context.Background()
	tokenKey := fmt.Sprintf("token:valid:%d", claims.UserID)
	storedToken, err := s.redisClient.Get(ctx, tokenKey)
	if err != nil || storedToken != tokenString {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *JWTService) InvalidateToken(userID int64) error {
	ctx := context.Background()
	tokenKey := fmt.Sprintf("token:valid:%d", userID)

	return s.redisClient.Delete(ctx, tokenKey)
}
