package common

var (
	// Auth ErrorCode
	ErrUserOrPasswordIncorrect = ErrorCode{1001, "Incorrect username or password"}
	ErrUserExists              = ErrorCode{1002, "User already exists"}
	ErrInvalidParam            = ErrorCode{1003, "Invalid Param"}
	ErrTokenInvalid            = ErrorCode{1004, "Invalid token"}
)