package utils

import "github.com/gin-gonic/gin"

type RouterGroup struct {
	gin.RouterGroup
}

func (r *RouterGroup) GETorHEAD(relativePath string, handlers ...gin.HandlerFunc) {
	r.GET(relativePath, handlers...)
	r.HEAD(relativePath, handlers...)
}
