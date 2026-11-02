/*
 * Copyright 2026 InfAI (CC SES)
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

package otelx_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	gin_mw "github.com/SENERGY-Platform/gin-middleware"
	"github.com/SENERGY-Platform/gin-middleware/otelx"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// usernames the baggage format cannot carry unencoded, and two that look encoded already
var usernames = []string{"alice", "first last", "jürgen", "a,b", `x"y;z\w`, "50%", "a%20b"}

// unsignedToken builds a jwt; the middleware parses tokens without validation
func unsignedToken(sub string, username string) string {
	encode := func(v interface{}) string {
		b, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return "Bearer " + encode(map[string]string{"alg": "HS256", "typ": "JWT"}) + "." + encode(map[string]string{"sub": sub, "preferred_username": username}) + ".sig"
}

// recordSpans initializes otelx once with an unreachable collector, then replaces its provider by a recorder
func recordSpans(t *testing.T) *tracetest.SpanRecorder {
	_, err := otelx.HTTPOpenTelemetry(context.Background(), "localhost:1", "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	previous := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(previous) })
	recorder := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	return recorder
}

func lastSpanAttribute(t *testing.T, recorder *tracetest.SpanRecorder, key string) string {
	spans := recorder.Ended()
	if len(spans) == 0 {
		t.Fatal("no span recorded")
	}
	for _, a := range spans[len(spans)-1].Attributes() {
		if string(a.Key) == key {
			return a.Value.AsString()
		}
	}
	return ""
}

func TestAddBaggageToHTTPRequestKeepsTheValue(t *testing.T) {
	for _, value := range usernames {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		err := otelx.AddBaggageToHTTPRequest(request, "username", value)
		if err != nil {
			t.Errorf("%q: %v", value, err)
			continue
		}
		bag := baggage.FromContext(request.Context())
		if got := bag.Member("username").Value(); got != value {
			t.Errorf("%q: member value %q", value, got)
		}
		//what a downstream service reads from the header
		parsed, err := baggage.Parse(bag.String())
		if err != nil {
			t.Errorf("%q: header %q does not parse: %v", value, bag.String(), err)
			continue
		}
		if got := parsed.Member("username").Value(); got != value {
			t.Errorf("%q: header %q parses to %q", value, bag.String(), got)
		}
	}
}

func TestHTTPOpenTelemetryServesEveryUsername(t *testing.T) {
	recorder := recordSpans(t)
	var reached bool
	var seen string
	inner := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		reached = true
		seen = baggage.FromContext(request.Context()).Member("username").Value()
	})
	handler, err := otelx.HTTPOpenTelemetry(context.Background(), "localhost:1", "test", inner)
	if err != nil {
		t.Fatal(err)
	}
	for _, username := range usernames {
		reached, seen = false, ""
		request := httptest.NewRequest(http.MethodGet, "/devices/x", nil)
		request.Header.Set("Authorization", unsignedToken("user-1", username))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || !reached {
			body, _ := io.ReadAll(response.Body)
			t.Errorf("%q: status %v, handler reached %v, body %q", username, response.Code, reached, string(body))
			continue
		}
		if seen != username {
			t.Errorf("%q: baggage in the handler carries %q", username, seen)
		}
		if got := lastSpanAttribute(t, recorder, "username"); got != username {
			t.Errorf("%q: span attribute %q", username, got)
		}
		if got := lastSpanAttribute(t, recorder, "user_id"); got != "user-1" {
			t.Errorf("%q: span attribute user_id %q", username, got)
		}
	}
}

func TestGinOpenTelemetryWithErrorHandlerServesEveryUsername(t *testing.T) {
	recorder := recordSpans(t)
	otelHandler, err := otelx.GinOpenTelemetry(context.Background(), "test", "localhost:1")
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin_mw.ErrorHandler(func(error) int { return http.StatusInternalServerError }, "\n"), otelHandler)
	var seen string
	router.GET("/devices/:id", func(c *gin.Context) {
		seen = baggage.FromContext(c.Request.Context()).Member("username").Value()
		c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	})
	for _, username := range usernames {
		seen = ""
		request := httptest.NewRequest(http.MethodGet, "/devices/x", nil)
		request.Header.Set("Authorization", unsignedToken("user-1", username))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		body, _ := io.ReadAll(response.Body)
		if response.Code != http.StatusOK || string(body) != `{"id":"x"}` {
			t.Errorf("%q: status %v, body %q", username, response.Code, string(body))
			continue
		}
		if seen != username {
			t.Errorf("%q: baggage in the handler carries %q", username, seen)
		}
		if got := lastSpanAttribute(t, recorder, "username"); got != username {
			t.Errorf("%q: span attribute %q", username, got)
		}
	}
}
