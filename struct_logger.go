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
	methodAttrKey       = "method"
	statusCodeAttrKey   = "status"
	latencyAttrKey      = "latency"
	pathAttrKey         = "path"
	protocolAttrKey     = "protocol"
	userAgentAttrKey    = "user_agent"
	bodySizeAttrKey     = "body_size"
	errorMessageAttrKey = "error_msg"
)

type structuredLogger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
}

func StructuredLoggerHandler(structLogger structuredLogger, skipPaths []string, skipper gin.Skipper, generators ...func(*gin.Context) (string, any)) gin.HandlerFunc {
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
			statusCodeAttrKey, gc.Writer.Status(),
			methodAttrKey, gc.Request.Method,
			pathAttrKey, path,
			protocolAttrKey, gc.Request.Proto,
			userAgentAttrKey, gc.Request.UserAgent(),
			latencyAttrKey, time.Now().Sub(start),
			bodySizeAttrKey, gc.Writer.Size(),
		}
		for _, generator := range generators {
			key, value := generator(gc)
			args = append(args, key, value)
		}
		if errMsg := joinErrors(gc.Errors.ByType(gin.ErrorTypePrivate)); errMsg != "" {
			args = append(args, errorMessageAttrKey, errMsg)
		}
		structLogger.DebugContext(gc.Request.Context(), http.StatusText(gc.Writer.Status()), args...)
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
