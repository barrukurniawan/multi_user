package main

import (
	"github.com/barrukurniawan/multi_user/core/routers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	routerConfig := routers.Config{
		Server: r,
	}
	routerConfig.Setup()
	r.Run(":8080")
}
