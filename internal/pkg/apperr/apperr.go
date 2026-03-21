package apperr

type AppError struct {
	Code int
	Msg  string
}

func (e *AppError) Error() string { return e.Msg }

func New(code int, msg string) error {
	return &AppError{Code: code, Msg: msg}
}

func Parse(err error, defaultCode int, defaultMsg string) (int, string) {
	if err == nil {
		return 200, "success"
	}
	if e, ok := err.(*AppError); ok {
		return e.Code, e.Msg
	}
	return defaultCode, defaultMsg
}
