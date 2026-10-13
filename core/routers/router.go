package routers

import (
	"net/http"

	"github.com/barrukurniawan/multi_user/core/controllers"
	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/barrukurniawan/multi_user/docs"
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
	c.SwaggerRoute()
	c.NotFoundRoute()
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

// buka http://localhost:8080/swagger di browser
func (c *Config) SwaggerRoute() {
	c.Server.GET("/swagger", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", docs.SwaggerHTML)
	})
	c.Server.GET("/swagger/openapi.yaml", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/yaml", docs.OpenAPI)
	})
}

// supaya 404 & 405 juga balas JSON, bukan teks bawaan gin
func (c *Config) NotFoundRoute() {
	c.Server.HandleMethodNotAllowed = true
	c.Server.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, dto.Error("NOT_FOUND", "endpoint tidak ditemukan"))
	})
	c.Server.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(http.StatusMethodNotAllowed, dto.Error("METHOD_NOT_ALLOWED", "method tidak diizinkan"))
	})
}
