package gin_mw

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
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

func StructuredLoggerHandler(structLogger structLogger, structAttrProvider structAttrProvider, skipPaths []string, skipper gin.Skipper, generators ...func(*gin.Context) (string, any)) gin.HandlerFunc {
	var skip map[string]struct{}
	if len(skipPaths) > 0 {
		skip = make(map[string]struct{})
		for _, path := range skipPaths {
			skip[path] = struct{}{}
		}
	}
	return func(gc *gin.Context) {
		start := time.Now()
		path := gc.Request.URL.Path
		rawQuery := gc.Request.URL.RawQuery
		gc.Next()
		if _, ok := skip[path]; ok || (skipper != nil && skipper(gc)) {
			return
		}
		if rawQuery != "" {
			path = path + "?" + rawQuery
		}
		args := []any{
			structAttrProvider.StatusCodeKey(), gc.Writer.Status(),
			structAttrProvider.MethodKey(), gc.Request.Method,
			structAttrProvider.PathKey(), path,
			structAttrProvider.ProtocolKey(), gc.Request.Proto,
			userAgentKey, gc.Request.UserAgent(),
			structAttrProvider.LatencyKey(), time.Now().Sub(start).String(),
			bodySizeKey, gc.Writer.Size(),
		}
		for _, generator := range generators {
			key, value := generator(gc)
			args = append(args, key, value)
		}
		if errMsg := joinErrors(gc.Errors.ByType(gin.ErrorTypePrivate)); errMsg != "" {
			args = append(args, structAttrProvider.ErrorKey(), errMsg)
		}
		structLogger.InfoContext(gc.Request.Context(), http.StatusText(gc.Writer.Status()), args...)
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
