package pullrequest

import (
	"Avito/pkg/gateway/errors"
	"Avito/pkg/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReassignPullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldUserID     string `json:"old_reviewer_id" binding:"required"`
}

func ReassignPullRequestHandler(cases *usecase.Cases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReassignPullRequestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errors.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}

		pr, replacedBy, err := cases.PullRequest.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldUserID)
		if err != nil {
			errors.HandleDomainError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"pr":          pr,
			"replaced_by": replacedBy,
		})
	}
}
