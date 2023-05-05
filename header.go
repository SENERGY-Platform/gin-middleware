package gin_mw

import "github.com/gin-gonic/gin"

func StaticHeaderHandler(items map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for key, val := range items {
			c.Header(key, val)
		}
		c.Next()
	}
}
