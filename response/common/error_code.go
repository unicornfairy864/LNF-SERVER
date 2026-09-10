package common

type ErrorCode struct {
	ErrorCode    int
	ErrorMessage string
}

var (
	ErrInternalServer = ErrorCode{500, "Internal Server Error"}
)