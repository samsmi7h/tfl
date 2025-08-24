package tfl

import "net/http"

type Error struct {
	Message  string
	HTTPCode int
}

func (e Error) Error() string {
	return e.Message
}

func HttpError(w http.ResponseWriter, err error) {
	if tflErr, ok := err.(Error); ok {
		http.Error(w, tflErr.Message, tflErr.HTTPCode)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewInternalError(message string) Error {
	return Error{
		Message:  message,
		HTTPCode: http.StatusInternalServerError,
	}
}

func NewBadRequestError(message string) Error {
	return Error{
		Message:  message,
		HTTPCode: http.StatusBadRequest,
	}
}

func NewNotFoundError(message string) Error {
	return Error{
		Message:  message,
		HTTPCode: http.StatusNotFound,
	}
}
