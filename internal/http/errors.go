package http

import (
	"encoding/json"
)

func Wrap(errCode, message string) string {
	type wrappedError struct {
		code    string
		message string
	}

	type errMessage struct {
		error wrappedError
	}

	v := errMessage{
		error: wrappedError{
			code:    errCode,
			message: message,
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
		message string
	}

	type errMessage struct {
		error wrappedError
	}

	v := errMessage{
		error: wrappedError{
			message: message,
		},
	}

	body, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return string(body)
}
