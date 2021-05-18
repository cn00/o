package main

import (
	"encoding/json"
	"os"
	"time"

	"octo/models"
	"octo/utils"

	"github.com/gin-gonic/gin"
)

func engine() *gin.Engine {
	engine := gin.New()
	if gin.IsDebugging() {
		engine.Use(gin.Logger())
	}
	c := &CORS{
		AllowAllOrigins: true,
		AllowHeaders:    []string{"X-Octo-Key", "Content-Type"},
		AllowMethods:    []string{"GET", "PUT", "POST", "DELETE"},
	}
	engine.Use(c.MiddleWare())
	engine.Use(
		utils.StackTraceLogger,
		jsonLogger,
		gin.Recovery(),
	)
	return engine
}

type jsonLog struct {
	Date       string      `json:"date"`
	StatusCode int         `json:"statusCode"`
	Latency    int64       `json:"latency"`
	ClientIP   string      `json:"clientIP"`
	Method     string      `json:"method"`
	Path       string      `json:"path"`
	AppID      int         `json:"appID,omitempty"`
	Expect     string      `json:"expect,omitempty"`
	UserAgent  string      `json:"userAgent,omitempty"`
	Key        string      `json:"key,omitempty"`
	CliVersion string      `json:"cliVersion,omitempty"`
	Detail     interface{} `json:"detail,omitempty"`
	Comment    string      `json:"comment,omitempty"`
}

func jsonLogger(c *gin.Context) {
	start := time.Now()

	// Process request
	c.Next()

	end := time.Now()
	latency := end.Sub(start)
	request := c.Request
	header := request.Header

	json.NewEncoder(os.Stdout).Encode(jsonLog{
		Date:       end.Format(time.RFC3339Nano),
		StatusCode: c.Writer.Status(),
		Latency:    latency.Nanoseconds(),
		ClientIP:   c.ClientIP(),
		Method:     request.Method,
		Path:       request.URL.Path,
		AppID:      contextAppID(c),
		Expect:     header.Get("Expect"),
		UserAgent:  header.Get("User-Agent"),
		Key:        header.Get("X-Octo-Key"),
		CliVersion: header.Get("X-Octo-Cli-Version"),
		Detail:     contextDetail(c),
		Comment:    c.Errors.String(),
	})
}

func contextAppID(c *gin.Context) int {
	val, exist := c.Get("app")
	if !exist {
		return 0
	}
	app, ok := val.(models.App)
	if !ok {
		return 0
	}
	return app.AppId
}

func contextDetail(c *gin.Context) interface{} {
	val, exist := c.Get("detail")
	if !exist {
		return nil
	}
	return val
}
