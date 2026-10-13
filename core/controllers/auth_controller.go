package controllers

import (
	"net/http"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/barrukurniawan/multi_user/core/services"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ctl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.WriteBindError(c, err)
		return
	}

	user, err := ctl.authService.Register(c.Request.Context(), req)
	if err != nil {
		helpers.WriteBindError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.Success(user))
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.WriteBindError(c, err)
		return
	}

	token, err := ctl.authService.Login(c.Request.Context(), req)
	if err != nil {
		helpers.WriteBindError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Success(gin.H{"token": token, "token_type": "Bearer"}))
}
