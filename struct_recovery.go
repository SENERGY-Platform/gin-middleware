package gin_mw

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const stackTraceKey = "stack_trace"

type recoveryStructLogger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}

func StructuredRecoveryHandler(structLogger recoveryStructLogger, handle gin.RecoveryFunc) gin.HandlerFunc {
	return func(gc *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				if brokenPipe {
					gc.Error(err.(error))
					gc.Abort()
				} else {
					var stack []byte
					buf := make([]byte, 1024)
					for {
						n := runtime.Stack(buf, false)
						if n < len(buf) {
							stack = buf[:n]
							break
						}
						buf = make([]byte, 2*len(buf))
					}
					structLogger.ErrorContext(gc.Request.Context(), fmt.Sprintf("%s", err), stackTraceKey, string(stack))
					handle(gc, err)
				}
			}
		}()
		gc.Next()
	}
}

func DefaultRecoveryFunc(gc *gin.Context, _ any) {
	gc.AbortWithStatus(http.StatusInternalServerError)
}
