package errors

import (
	"Avito/pkg/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondError(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func HandleDomainError(c *gin.Context, err error) {
	var domainErr *domain.DomainError
	ok := errors.As(err, &domainErr)
	if !ok {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error: %v")
		return
	}

	statusCode := getHTTPStatusCode(domainErr.Code)
	RespondError(c, statusCode, string(domainErr.Code), domainErr.Message)
}

func getHTTPStatusCode(code domain.ErrorCode) int {
	switch code {
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrTeamExists, domain.ErrPRExists:
		return http.StatusBadRequest
	case domain.ErrPRMerged, domain.ErrNotAssigned, domain.ErrNoCandidate:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
