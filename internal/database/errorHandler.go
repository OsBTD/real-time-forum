package db

import (
	"net/http"
)

type ErrorPage struct {
	Code    int
	Message string
	Is405   bool
	Is404   bool
	Is500   bool
	Is403   bool
	Is400   bool
	Is401   bool
}

// helper for handeling errors
func HandleError(w http.ResponseWriter, code int, message string) {
	errorPage := ErrorPage{
		Code:    code,
		Message: message,
		Is405:   code == http.StatusMethodNotAllowed,
		Is404:   code == http.StatusNotFound,
		Is500:   code == http.StatusInternalServerError,
		Is403:   code == http.StatusForbidden,
		Is400:   code == http.StatusBadRequest,
		Is401:   code == http.StatusUnauthorized,
	}

	w.WriteHeader(code)
	RenderTemplate(w, "error", map[string]interface{}{
		"Code":    errorPage.Code,
		"Message": errorPage.Message,
		"Is405":   errorPage.Is405,
		"Is404":   errorPage.Is404,
		"Is500":   errorPage.Is500,
		"Is403":   errorPage.Is403,
		"Is400":   errorPage.Is400,
		"Is401":   errorPage.Is401,
	})
}
