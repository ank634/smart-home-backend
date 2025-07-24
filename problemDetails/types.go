package problemdetails

type problemDetailError string

const (
	NULL_NOT_ALLOWED_ERROR problemDetailError = "NULL_NOT_ALLOWED"
	NOT_UNIQUE_ERROR       problemDetailError = "NOT_UNIQUE"
	ILLEGAL_VALUE_ERROR    problemDetailError = "ILLEGAL_VALUE"
)

type ProblemDetail struct {
	ErrorType problemDetailError
	Title     string
	Status    int
	Detail    string
}
