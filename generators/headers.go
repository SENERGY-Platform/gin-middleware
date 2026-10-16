/*
 * Copyright 2025 InfAI (CC SES)
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

package generators

import (
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

const headersKey = "headers"

// Headers that may carry credentials; Sec-WebSocket-Protocol carries bearer tokens for browser websockets. Lower case!
var headersFilter = []string{"authorization", "proxy-authorization", "cookie", "x-api-key", "x-auth-token", "sec-websocket-protocol"}

func HeadersGenerator(gc *gin.Context) (string, any) {
	result := map[string]string{}
	for key := range gc.Request.Header {
		if slices.Contains(headersFilter, strings.ToLower(key)) {
			continue
		}
		result[key] = gc.Request.Header.Get(key)
	}
	return headersKey, result
}
