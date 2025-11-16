package team

import (
	"Avito/pkg/gateway/errors"
	"Avito/pkg/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTeamHandler(cases *usecase.Cases) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamName := c.Query("team_name")
		if teamName == "" {
			errors.RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", "team_name query parameter is required")
			return
		}

		team, err := cases.Team.GetTeam(c.Request.Context(), teamName)
		if err != nil {
			errors.HandleDomainError(c, err)
			return
		}

		c.JSON(http.StatusOK, team)
	}
}
