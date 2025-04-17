package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ParseToken(tokenString string, secretKey []byte) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}

func GenerateToken(login string, secretKey []byte) (string, error) {
	claims := jwt.MapClaims{
		"iss": "skillsRockAuth",
		"sub": login,
		"aud": "skillsRockTodo",
		"exp": time.Now().Add(time.Minute * 5).Unix(),
		"nbf": time.Now().Unix(),
		"jat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
