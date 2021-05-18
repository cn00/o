package main

import (
	"octo/utils"

	"github.com/gin-gonic/gin"
)

func engine() *gin.Engine {
	engine := gin.New()
	engine.Use(
		utils.StackTraceLogger,
		gin.Logger(),
		gin.Recovery(),
		gin.ErrorLogger(),
	)
	return engine
}
