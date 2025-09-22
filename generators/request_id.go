package generators

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

const requestIdKey = "request_id"

func RequestIdGenerator(gc *gin.Context) (string, any) {
	return requestIdKey, requestid.Get(gc)
}
