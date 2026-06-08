package gin_mw

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	userAgentKey = "user_agent"
	bodySizeKey  = "body_size"
)

type structLogger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
}

type structAttrProvider interface {
	PathKey() string
	StatusCodeKey() string
	MethodKey() string
	LatencyKey() string
	ProtocolKey() string
	ErrorKey() string
}

func StructLoggerHandler(structLogger structLogger, structAttrProvider structAttrProvider, skipPaths []string, skipper gin.Skipper, generators ...func(*gin.Context) (string, any)) gin.HandlerFunc {
	var skip map[string]struct{}
	if len(skipPaths) > 0 {
		skip = make(map[string]struct{})
		for _, path := range skipPaths {
			skip[path] = struct{}{}
		}
	}
	return func(gc *gin.Context) {
		start := time.Now()
		gc.Next()
		if _, ok := skip[gc.Request.URL.Path]; ok || (skipper != nil && skipper(gc)) {
			return
		}
		args := []any{
			structAttrProvider.StatusCodeKey(), gc.Writer.Status(),
			structAttrProvider.MethodKey(), gc.Request.Method,
			structAttrProvider.PathKey(), gc.Request.URL.String(),
			structAttrProvider.ProtocolKey(), gc.Request.Proto,
			userAgentKey, gc.Request.UserAgent(),
			structAttrProvider.LatencyKey(), time.Now().Sub(start),
			bodySizeKey, gc.Writer.Size(),
		}
		for _, generator := range generators {
			key, value := generator(gc)
			args = append(args, key, value)
		}
		if errMsg := joinErrors(gc.Errors.ByType(gin.ErrorTypePrivate)); errMsg != "" {
			args = append(args, structAttrProvider.ErrorKey(), errMsg)
		}
		structLogger.InfoContext(gc, http.StatusText(gc.Writer.Status()), args...)
	}
}

func joinErrors(errs []*gin.Error) string {
	errsLen := len(errs)
	if errsLen == 0 {
		return ""
	}
	var buffer strings.Builder
	for i, err := range errs {
		fmt.Fprintf(&buffer, "Err#%02d: %s", i+1, err.Err)
		if i < errsLen-1 {
			fmt.Fprint(&buffer, "; ")
		}
	}
	return buffer.String()
}
