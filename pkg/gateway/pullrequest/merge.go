package pullrequest

import (
	"Avito/pkg/gateway/errors"
	"Avito/pkg/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

func MergePullRequestHandler(cases *usecase.Cases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MergePullRequestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errors.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}

		pr, err := cases.PullRequest.MergePullRequest(c.Request.Context(), req.PullRequestID)
		if err != nil {
			errors.HandleDomainError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"pr": pr})
	}
}
