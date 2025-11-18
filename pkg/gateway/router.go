package gateway

import (
	"Avito/pkg/config"
	"Avito/pkg/gateway/pullrequest"
	"Avito/pkg/gateway/team"
	"Avito/pkg/gateway/user"
	"Avito/pkg/usecase"
	"context"
	"log"

	"github.com/gin-gonic/gin"
)

func setupRouter(ctx context.Context, cfg *config.Config, r *gin.Engine, cases *usecase.Cases) {
	if err := RegisterSwagger(ctx, r, cfg, "./docs/openapi.yml.tpl"); err != nil {
		log.Fatalf("swagger init: %v", err)
	}
	teamGroup := r.Group("/team")
	{
		teamGroup.POST("/add", team.CreateTeamHandler(cases))
		teamGroup.GET("/get", team.GetTeamHandler(cases))
	}

	userGroup := r.Group("/users")
	{
		userGroup.POST("/setIsActive", user.SetIsActiveHandler(cases))
		userGroup.GET("/getReview", user.GetUserReviewsHandler(cases))
	}

	prGroup := r.Group("/pullRequest")
	{
		prGroup.POST("/create", pullrequest.CreatePullRequestHandler(cases))
		prGroup.POST("/merge", pullrequest.MergePullRequestHandler(cases))
		prGroup.POST("/reassign", pullrequest.ReassignPullRequestHandler(cases))
	}
}
