package devicesCrud

type ErrorNotNullViolation struct{ message string }

func (e ErrorNotNullViolation) Error() string { return e.message }

type ErrorDuplicateData struct{ message string }

func (e ErrorDuplicateData) Error() string { return e.message }

type ErrorIllegalData struct{ message string }

func (e ErrorIllegalData) Error() string { return e.message }
