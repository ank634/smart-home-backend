package problemdetails

import (
	"encoding/json"
	"net/http"
)

type problemDetailError string

const (
	NULL_NOT_ALLOWED_ERROR problemDetailError = "NULL_NOT_ALLOWED"
	NOT_UNIQUE_ERROR       problemDetailError = "NOT_UNIQUE"
	ILLEGAL_VALUE_ERROR    problemDetailError = "ILLEGAL_VALUE"
)

type problemDetail struct {
	ErrorType problemDetailError
	Title     string
	Status    int
	Detail    string
}

func ProblemDetail(w http.ResponseWriter, errorType problemDetailError, title string, statusCode int, detail string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/problem+json")
	json.NewEncoder(w).Encode(problemDetail{ErrorType: errorType, Title: title, Status: statusCode, Detail: detail})
}
