package routes

import (
	"github.com/cplace/cssc-operator/internal/environment"
	"github.com/gin-gonic/gin"
)

// Init initializes the routes for the given environment.
func Init(env environment.Environment) {
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})
	router.POST("/instances", func(c *gin.Context) {
		c.String(200, "OK")
	})
}
