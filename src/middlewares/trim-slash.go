package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func TrimSlashMiddleware(c *gin.Context) {
	path := c.Request.URL.Path
	if path != "/" && path[len(path)-1] == '/' {
		trimmed := path[:len(path)-1]
		c.Redirect(http.StatusPermanentRedirect, trimmed)
		return
	}
	c.Next()
}
