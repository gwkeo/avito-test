package http

import (
	"encoding/json"
)

func WrapToJSON(errCode, message string) string {
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
