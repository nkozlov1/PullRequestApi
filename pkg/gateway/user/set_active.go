package user

import (
	"Avito/pkg/gateway/errors"
	"Avito/pkg/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

func SetIsActiveHandler(cases *usecase.Cases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SetIsActiveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errors.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}
		user, err := cases.User.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
		if err != nil {
			errors.HandleDomainError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}
