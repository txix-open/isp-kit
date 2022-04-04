package httperrors

import (
	"fmt"
	"net/http"

	"github.com/integration-system/isp-kit/json"
)

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func (e Error) WriteError(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	return json.NewEncoder(w).Encode(e)
}

func New(code int, err error) Error {
	return Error{
		Code:    code,
		Message: err.Error(),
	}
}
