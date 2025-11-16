package team

import (
	"Avito/pkg/domain"
	"Avito/pkg/gateway/errors"
	"Avito/pkg/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTeamHandler(cases *usecase.Cases) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &domain.Team{}
		if err := c.ShouldBindJSON(&req); err != nil {
			errors.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}

		team, err := cases.Team.CreateTeam(c.Request.Context(), req)
		if err != nil {
			errors.HandleDomainError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"team": team})
	}
}
