package routers

import (
	"net/http"

	"github.com/barrukurniawan/multi_user/core/controllers"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Server         *gin.Engine
	AuthController *controllers.AuthController
}

func (c *Config) Setup() {
	c.HealthRoute()
	c.SetupAuthRoute()
}
func (c *Config) HealthRoute() {
	c.Server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})
}

func (c *Config) SetupAuthRoute() {
	r := c.Server.Group("/api/v1/auth")
	r.POST("/register", c.AuthController.Register)
	r.POST("/login", c.AuthController.Login)
}
