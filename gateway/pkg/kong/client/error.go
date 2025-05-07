package client

import (
	"slices"
)

type ApiResponse interface {
	StatusCode() int
}

type ApiError interface {
	error
	Retriable() bool
}

type apiError struct {
	StatusCode   int
	Message      string
	RetryAllowed bool
}

func (e *apiError) Error() string {
	return e.Message
}

func (e *apiError) Retriable() bool {
	return e.RetryAllowed
}

func CheckStatusCode(res ApiResponse, okStatusCodes ...int) ApiError {
	if slices.Contains(okStatusCodes, res.StatusCode()) {
		return nil
	}

	if res.StatusCode() >= 500 {
		return &apiError{
			StatusCode:   res.StatusCode(),
			Message:      "Kong server error",
			RetryAllowed: true,
		}
	}

	return &apiError{
		StatusCode:   res.StatusCode(),
		Message:      "Kong client error",
		RetryAllowed: false,
	}
}
