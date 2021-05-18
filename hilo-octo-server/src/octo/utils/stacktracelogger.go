package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func StackTraceLogger(c *gin.Context) {
	c.Next()
	errors := c.Errors.ByType(gin.ErrorTypePrivate)
	for i, msg := range errors {
		fmt.Printf("StackTrace #%02d: %+v\n", (i + 1), msg.Err)
	}
}
