package util

import (
	"errors"
	"fmt"
	"log"
	"syscall"

	"github.com/dgrijalva/jwt-go/v4"
	iam "github.com/netsoc/iam/client"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// ErrPasswordMismatch indicates that a user entered two different passwords
	ErrPasswordMismatch = errors.New("passwords didn't match")
)

// IsDebug determines if debugging is enabled
var IsDebug bool

// Debugf prints log messages only if debugging is enabled
func Debugf(format string, v ...interface{}) {
	if !IsDebug {
		return
	}

	log.Printf(format, v...)
}

// APIError re-formats an OpenAPI-generated API client error
func APIError(err error) error {
	var iamGeneric iam.GenericOpenAPIError
	if ok := errors.As(err, &iamGeneric); ok {
		if iamError, ok := iamGeneric.Model().(iam.Error); ok {
			return errors.New(iamError.Message)
		}
		return err
	}

	return err
}

// ReadPassword reads a password from stdin
func ReadPassword(confirm bool) (string, error) {
	fmt.Print("Enter password: ")
	p, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}
	fmt.Println()

	if confirm {
		fmt.Print("Again: ")
		p2, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return "", fmt.Errorf("read failed: %w", err)
		}
		fmt.Println()

		if string(p2) != string(p) {
			return "", ErrPasswordMismatch
		}
	}

	return string(p), nil
}

// UserClaims represents claims in an auth JWT
type UserClaims struct {
	jwt.StandardClaims
	IsAdmin bool `json:"is_admin"`
	Version uint `json:"version"`
}
