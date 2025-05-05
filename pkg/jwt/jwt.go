package jwt

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

//jti (JWT ID) — уникальный идентификатор токена.
//iss (issuer) — издатель токена;
//sub (subject) — субъект, которому выдан токен;
//aud (audience) — получатели, которым предназначается данный токен;
//exp (expiration time) — время, когда токен станет невалидным;
//nbf (not before) — время, с которого токен должен считаться действительным;
//iat (issued at) — время, в которое был выдан токен;

type TokenClaims struct {
	Jti        *uuid.UUID `json:"jti"`
	Sub        *uuid.UUID `json:"sub"`
	DeviceCode string     `json:"device"`
	TokenType  string     `json:"type"`
	jwt.RegisteredClaims
}

func CreateToken(userId *uuid.UUID, deviceCode string, tokenType string, lifetime time.Duration, privateKey *rsa.PrivateKey) (string, *TokenClaims, error) {
	tokenId := uuid.New()
	now := time.Now()
	tokenClaims := TokenClaims{
		&tokenId,
		userId,
		deviceCode,
		tokenType,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	tokenJwt := jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaims)
	tokenString, err := tokenJwt.SignedString(privateKey)
	if err != nil {
		return "", nil, err
	}
	return tokenString, &tokenClaims, nil

}
