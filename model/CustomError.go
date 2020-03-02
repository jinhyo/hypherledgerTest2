package model

import "fmt"

// ConvertErrorType is ...
const (
	ConvertErrorType = "Convert"
)

// CustomError is ...
type CustomError struct {
	ErrorType  string
	TargetName string
	Message    string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("failed to %s %s, error: %s", e.ErrorType, e.TargetName, e.Message)
}
