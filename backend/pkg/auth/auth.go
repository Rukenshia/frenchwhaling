package auth

import (
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrCouldNotParse  = errors.New("could not parse jwt")
	ErrInvalidSubject = errors.New("user is not authorized for this account id")
)

func VerifyToken(token, accountId string) error {
	parser := jwt.Parser{}

	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(strings.Replace(token, "Bearer ", "", 1), &claims)
	if err != nil {
		return ErrCouldNotParse
	}

	// A valid JWT is supplied, but for another account
	if claims["sub"] != accountId {
		return ErrInvalidSubject
	}

	return nil
}
