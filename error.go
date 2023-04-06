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
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func ErrorHandler(f func(error) int, sep string) gin.HandlerFunc {
	return func(gc *gin.Context) {
		gc.Next()
		if !gc.IsAborted() && len(gc.Errors) > 0 {
			var errs []string
			for _, e := range gc.Errors {
				if sc := f(e); sc != 0 {
					gc.Status(sc)
				}
				errs = append(errs, e.Error())
			}
			if gc.Writer.Status() < 400 {
				gc.Status(http.StatusInternalServerError)
			}
			gc.String(-1, strings.Join(errs, sep))
		}
	}
}
