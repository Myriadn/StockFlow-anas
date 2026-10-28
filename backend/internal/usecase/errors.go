package usecase

type AppError struct {
	Status int
	Code   string
	Extra  map[string]any
}

func (e *AppError) Error() string { return e.Code }

func NewAppError(status int, code string, extra map[string]any) *AppError {
	return &AppError{Status: status, Code: code, Extra: extra}
}
