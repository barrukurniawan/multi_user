package routers

import (
	"net/http"

	"github.com/barrukurniawan/multi_user/core/controllers"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Server         *gin.Engine
	JWTManager     *helpers.JWTManager
	AuthController *controllers.AuthController
	TaskController *controllers.TaskController
}

func (c *Config) Setup() {
	c.HealthRoute()
	c.AuthRoute()
	c.TaskRoute()
}
func (c *Config) HealthRoute() {
	c.Server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})
}

func (c *Config) AuthRoute() {
	r := c.Server.Group("/api/v1/auth")
	r.POST("/register", c.AuthController.Register)
	r.POST("/login", c.AuthController.Login)
}

func (c *Config) TaskRoute() {
	r := c.Server.Group("/api/v1/tasks", helpers.AuthMiddleware(c.JWTManager))
	r.POST("", c.TaskController.Create)
	r.GET("", c.TaskController.List)
	r.GET("/:id", c.TaskController.GetByID)
	r.PUT("/:id", c.TaskController.Update)
	r.DELETE("/:id", c.TaskController.Delete)
	r.POST("/:id/assign", c.TaskController.Assign)
}
