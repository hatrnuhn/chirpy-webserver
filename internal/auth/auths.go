package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// parses any AJWT or RJWT from Authorization: {scheme} {JWT} request header
func ParseReq(r *http.Request, jwtSecret string, scheme string) (*jwt.Token, error) {
	tokenString, err := GetAuthHeadToken(r, scheme)
	if err != nil {
		return nil, errors.New("couldn't read request header")
	}
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func GetAuthHeadToken(r *http.Request, scheme string) (string, error) {
	scheme = scheme + " "
	authHead := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHead, scheme) {
		return "", errors.New("invalid auth header")
	}

	tokenString := strings.TrimPrefix(authHead, scheme)

	return tokenString, nil
}
