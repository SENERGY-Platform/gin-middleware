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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHeadersGeneratorDropsCredentialHeaders(t *testing.T) {
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	gc.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	blocked := []string{"Authorization", "Proxy-Authorization", "Cookie", "X-Api-Key", "X-Auth-Token", "Sec-WebSocket-Protocol"}
	for _, h := range blocked {
		gc.Request.Header.Set(h, "secret-"+h)
	}
	gc.Request.Header.Set("X-Custom-Header", "normal-value")

	key, value := HeadersGenerator(gc)
	if key != headersKey {
		t.Fatalf("key %q, want %q", key, headersKey)
	}
	headers := value.(map[string]string)
	for _, h := range blocked {
		if _, ok := headers[http.CanonicalHeaderKey(h)]; ok {
			t.Errorf("header %s is logged", h)
		}
	}
	if headers["X-Custom-Header"] != "normal-value" {
		t.Errorf("normal header missing: %v", headers)
	}
}
