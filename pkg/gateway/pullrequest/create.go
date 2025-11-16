package pullrequest

import (
	"Avito/pkg/gateway/errors"
	"Avito/pkg/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

func CreatePullRequestHandler(cases *usecase.Cases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreatePullRequestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errors.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}

		pr, err := cases.PullRequest.CreatePullRequest(c.Request.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
		if err != nil {
			errors.HandleDomainError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"pr": pr})
	}
}
