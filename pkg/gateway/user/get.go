package user

import (
	"Avito/pkg/gateway/errors"
	"Avito/pkg/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserReviewsHandler(cases *usecase.Cases) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			errors.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "user_id query parameter is required")
			return
		}

		prs, err := cases.PullRequest.GetUserReviews(c.Request.Context(), userID)
		if err != nil {
			errors.HandleDomainError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":       userID,
			"pull_requests": prs,
		})
	}
}
