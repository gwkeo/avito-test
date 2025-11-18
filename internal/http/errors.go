package http

import (
	"encoding/json"
)

type wrappedError struct {
	Code       string `json:"code,omitempty"`
	ErrMessage string `json:"error_message,omitempty"`
	Message    string `json:"message"`
}

type errMessage struct {
	Error wrappedError `json:"error"`
}

func Wrap(errCode, message string) string {
	v := errMessage{
		Error: wrappedError{
			Code:    errCode,
			Message: message,
		},
	}

	body, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return string(body)
}

func WrapWithoutCode(message, errorMessage string) string {

	v := errMessage{
		Error: wrappedError{
			Message:    message,
			ErrMessage: errorMessage,
		},
	}

	body, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return string(body)
}

func WrapMessageOnly(message string) string {
	v := errMessage{
		Error: wrappedError{
			Message: message,
		},
	}

	body, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return string(body)
}
