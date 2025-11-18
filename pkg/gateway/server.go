package gateway

import (
	"Avito/pkg/config"
	"Avito/pkg/usecase"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	HTTPServer http.Server
	Router     *gin.Engine
}

func NewServer(ctx context.Context, cfg *config.Config, cases *usecase.Cases) *Server {
	r := gin.Default()
	s := &Server{
		Router: r,
		HTTPServer: http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
			Handler: r,
		},
	}
	setupRouter(ctx, cfg, s.Router, cases)
	return s
}

func (s *Server) Run(ctx context.Context) error {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return s.HTTPServer.ListenAndServe()
	})

	<-ctx.Done()
	err := s.HTTPServer.Shutdown(ctx)
	return err
}
