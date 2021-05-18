package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorReportEndpoint(c *gin.Context) {
	var e map[string]interface{}
	if c.BindJSON(&e) == nil {
		c.Set("detail", struct {
			ErrorReport map[string]interface{} `json:"errorReport"`
		}{
			ErrorReport: e,
		})
		c.JSON(http.StatusOK, nil)
	}
}
