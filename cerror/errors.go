package cerror

const (
	ErrInvalidUserID       = "invalid user ID"
	ErrInvalidPhoneNumber  = "invalid phone number"
	ErrUserNotFound        = "user not found"
	ErrUserAlreadyExists   = "user already exists"
	ErrDatabaseConnection  = "database connection error"
	ErrDatabaseOperation   = "database operation error"
	ErrInternalServerError = "internal server error"
	ErrUnauthorized        = "access denied: unauthorized"
	ErrForbidden           = "access denied: forbidden"
	ErrNotFound            = "resource not found"
	ErrBadRequest          = "bad request"
	ErrValidation          = "validation failed"
)

type AppError struct {
	Err        error
	StatusCode int
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}
