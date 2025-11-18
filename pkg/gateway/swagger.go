package gateway

import (
	"Avito/pkg/config"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterSwagger(ctx context.Context, r *gin.Engine, cfg *config.Config, file string) error {
	tplBytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read openapi template: %w", err)
	}
	tpl, err := template.New("openapi").Parse(string(tplBytes))
	if err != nil {
		return fmt.Errorf("parse openapi template: %w", err)
	}
	r.GET("/openapi.yaml", func(c *gin.Context) {
		var buf bytes.Buffer
		data := map[string]interface{}{
			"Host": "localhost",
			"Port": cfg.Server.Port,
		}
		if err := tpl.Execute(&buf, data); err != nil {
			c.String(http.StatusInternalServerError, "failed to render openapi: %v", err)
			return
		}
		c.Header("Content-Type", "application/yaml")
		_, _ = io.Copy(c.Writer, &buf)
	})
	urlOpts := ginSwagger.URL("/openapi.yaml")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, urlOpts))

	return nil
}
