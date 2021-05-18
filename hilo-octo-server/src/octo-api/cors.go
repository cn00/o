package main

import (
	"strings"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
)

type CORS struct {
	AllowAllOrigins  bool
	AllowCredentials bool
	AllowHeaders     []string
	AllowMethods     []string
	AllowOrigins     []string
	ExposeHeaders    []string
	MaxAge           time.Duration
}

func (c *CORS) MiddleWare() gin.HandlerFunc {
	if !c.AllowAllOrigins && len(c.AllowOrigins) == 0 {
		panic("Please set value for AllowOrigins")
	}
	return func(context *gin.Context) {
		origin := context.Request.Header.Get("Origin")
		if !c.validateOrigin(origin) {
			return
		}
		preFlight := false
		valid := false
		// PreFlight Check
		if context.Request.Method == "OPTIONS" {
			requestHeaderMethod := context.Request.Header.Get("Access-Control-Request-Method")
			if len(requestHeaderMethod) != 0 {
				preFlight = true
				valid = c.handlePreFlight(context, requestHeaderMethod)
			}
		}

		if !preFlight {
			c.handleRequest(context)
			valid = true
		}

		if valid {
			if c.AllowCredentials {
				context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				context.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			if c.AllowAllOrigins {
				context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				context.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			if preFlight {
				context.AbortWithStatus(200)
			}
			return
		}

	}

}

func (c *CORS) handlePreFlight(context *gin.Context, requestMethod string) bool {
	if ok := c.validateMethods(requestMethod); ok == false {
		return false
	}

	headers := context.Request.Header.Get("Access-Control-Request-Headers")
	if ok := c.validateHeaders(headers); ok == true {
		context.Writer.Header().Set("Access-Control-Allow-Methods", requestMethod)
		context.Writer.Header().Add("Access-Control-Allow-Headers", headers)

		if c.MaxAge != 0 {
			context.Writer.Header().Set("Access-Control-Max-Age", strconv.FormatInt(int64(c.MaxAge/time.Second), 10))
		}
		return true
	}

	return false
}

func (c *CORS) handleRequest(context *gin.Context) {
	if len(c.ExposeHeaders) != 0 {
		for _, value := range c.ExposeHeaders {
			context.Writer.Header().Add("Access-Control-Expose-Headers", value)
		}
	}
}

func (c *CORS) validateOrigin(origin string) bool {
	if c.AllowAllOrigins {
		return true
	}
	if len(origin) == 0 {
		return false
	}
	for _, allowOrg := range c.AllowOrigins {
		if allowOrg == origin {
			return true
		}
	}
	return false
}

func (c *CORS) validateHeaders(requestHeaders string) bool {
	if len(requestHeaders) == 0 {
		return false
	}
	headerValues := strings.Split(requestHeaders, ",")
	for _, headerValue := range headerValues {
		header := strings.Trim(headerValue, " \t\r\n")
		match := false
		for _, allowHeader := range c.AllowHeaders {
			if strings.EqualFold(allowHeader, header) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	return true
}

func (c *CORS) validateMethods(requestMethod string) bool {
	if len(requestMethod) != 0 {
		for _, value := range c.AllowMethods {
			if requestMethod == value {
				return true
			}
		}
	}
	return false
}
