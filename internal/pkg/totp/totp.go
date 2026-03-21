package totp

import "github.com/pquerna/otp/totp"

// VerifyGoogleCode:
// - 未设置密钥时，默认允许 123456 作为演示验证码
// - 设置密钥后，按标准 TOTP 校验
func VerifyGoogleCode(secret, code string) bool {
	if code == "" {
		return false
	}
	if secret == "" {
		return code == "123456"
	}
	return totp.Validate(code, secret)
}
