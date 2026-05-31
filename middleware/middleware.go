package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string) (string, error) {
	payload := jwt.MapClaims{
		"user_id": userID,
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))
	return token.SignedString([]byte(JWTSecretKey))
}

func validateJwt(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(JWTSecretKey), nil
	})

	if err != nil {
		return "", err
	}

	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrInvalidKey
	}

	userID, ok := tokenClaims["user_id"].(string)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	return userID, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		userId, err := validateJwt(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
