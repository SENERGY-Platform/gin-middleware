/*
 * Copyright 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gin_mw

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

type Logger interface {
	Error(v ...any)
	Debug(v ...any)
}

func LoggerHandler(logger Logger) gin.HandlerFunc {
	return func(gc *gin.Context) {
		start := time.Now().UTC()
		path := gc.Request.URL.Path
		raw := gc.Request.URL.RawQuery
		gc.Next()
		end := time.Now().UTC()
		latency := end.Sub(start)
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		}
		if raw != "" {
			path = path + "?" + raw
		}
		errs := gc.Errors.ByType(gin.ErrorTypePrivate)
		if len(errs) > 0 {
			for _, e := range gc.Errors {
				logger.Error(e)
			}
		}
		logger.Debug(fmt.Sprintf("%3d | %v | %s %#v", gc.Writer.Status(), latency, gc.Request.Method, path))
	}
}
