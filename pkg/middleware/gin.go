package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinMiddleware converts http.Handler middleware to gin.HandlerFunc
func GinMiddleware(middleware func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}
