package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Server *gin.Engine
}

func (c *Config) Setup() {
	c.HealthRoute()
}
func (c *Config) HealthRoute() {
	c.Server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})
}
