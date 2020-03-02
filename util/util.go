package util

import (
	"hypherledgertest2/model"
	"strconv"
)

// ConverToPositive is ...
func ConverToPositive(value, targetName string) (int, error) {
	amountInt, err := strconv.Atoi(value)
	if err != nil {
		return 0, &model.CustomError{
			ErrorType:  model.ConvertErrorType,
			TargetName: targetName,
			Message:    "must be integer"}
	}

	if amountInt <= 0 {
		return 0, &model.CustomError{
			ErrorType:  model.ConvertErrorType,
			TargetName: targetName,
			Message:    "must be more than zero"}
	}

	return amountInt, nil
}
