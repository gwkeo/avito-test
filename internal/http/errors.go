package http

import (
	"encoding/json"
)

func Wrap(errCode, message string) string {
	type wrappedError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	type errMessage struct {
		Error wrappedError `json:"error"`
	}

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

func WrapWithoutCode(message string) string {
	type wrappedError struct {
		Message string `json:"message"`
	}

	type errMessage struct {
		Error wrappedError `json:"error"`
	}

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
