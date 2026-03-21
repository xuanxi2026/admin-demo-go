package ecode

const (
	Success = 200

	InvalidParams = 40001
	Unauthorized  = 40101
	TokenExpired  = 40102
	Forbidden     = 40301

	UserExists       = 40901
	LoginFailed      = 42201
	GoogleCodeFailed = 42202
	GoogleNotBound   = 42203

	InternalError = 50000
)
