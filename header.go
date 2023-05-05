package gin_mw

import "github.com/gin-gonic/gin"

func NewStaticHeaderHandler(items map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for key, val := range items {
			c.Header(key, val)
		}
		c.Next()
	}
}
