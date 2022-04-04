package httperrors

import (
	"fmt"
	"net/http"

	"github.com/integration-system/isp-kit/json"
)

type HttpError struct {
	Code    int
	Message string
}

func (e HttpError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func (e HttpError) WriteError(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	return json.NewEncoder(w).Encode(e)
}

func NewHttpError(code int, err error) HttpError {
	return HttpError{
		Code:    code,
		Message: err.Error(),
	}
}
