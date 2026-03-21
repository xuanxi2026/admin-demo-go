package service

import "github.com/pquerna/otp/totp"

func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}
