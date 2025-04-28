package jwttoken

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims struct {
	DeviceCode string       `json:"device"`
	Roles      []*uuid.UUID `json:"roles"`
	Jti        *uuid.UUID   `json:"jti"`
	Sub        *uuid.UUID   `json:"sub"`
	jwt.RegisteredClaims
}

func CreateToken(sub *uuid.UUID, deviceCode string, roles []*uuid.UUID, lifetime time.Duration, secretKey []byte) (string, *TokenClaims, error) {
	jti := uuid.New()
	now := time.Now()
	tokenClaims := TokenClaims{
		deviceCode, //device code - код устройства
		roles,      // roles - слайс идентификаторов ролей
		&jti,       // JWT ID — уникальный идентификатор токена
		sub,        // subject — субъект, которому выдан токен
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)), // expiration time — время, когда токен станет невалидным
			NotBefore: jwt.NewNumericDate(now),               // not before — время, с которого токен должен считаться действительным
			IssuedAt:  jwt.NewNumericDate(now),               // issued at — время, в которое был выдан токен
		},
	}
	tokenJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := tokenJwt.SignedString(secretKey)
	if err != nil {
		return "", nil, err
	}
	return tokenString, &tokenClaims, nil

}
func GetToken(tokenString string, secretKey []byte) (*jwt.Token, error) {
	tokenJwt, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	switch {
	case tokenJwt.Valid:
		return tokenJwt, nil
	case errors.Is(err, jwt.ErrTokenMalformed):
		return nil, err
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return nil, err
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		return nil, err
	default:
		return nil, err
	}
}
func GetClaims(tokenString string, secretKey []byte) (*TokenClaims, error) {
	tokenJwt, err := GetToken(tokenString, secretKey)
	if err != nil {
		return nil, err
	}
	tokenClaims, ok := tokenJwt.Claims.(*TokenClaims)
	if !ok {
		return nil, err
	}
	return tokenClaims, nil
}
