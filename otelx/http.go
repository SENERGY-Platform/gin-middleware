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

package otelx

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"

	"go.opentelemetry.io/otel/trace"
)

// HTTPOpenTelemetry initializes OpenTelemetry with the given service name and options, and returns an http.Handler for tracing incoming requests.
// OpenTelemetry is initialized only once, subsequent calls will return a handler without reinitializing OpenTelemetry.
// endpoint might be empty, in that case a default will be used.
func HTTPOpenTelemetry(ctx context.Context, endpoint string, serviceName string, handler http.Handler, options ...otelhttp.Option) (http.Handler, error) {
	initOpenTelemetry(ctx, serviceName, endpoint)

	h := &otelHandler{
		handler:     handler,
		serviceName: serviceName,
		options:     options,
	}

	return h, nil
}

func AddBaggageToHTTPRequest(r *http.Request, key, value string) error {
	if r == nil {
		return fmt.Errorf("request is nil")
	}
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	if value == "" {
		return fmt.Errorf("value is empty")
	}
	if strings.Contains(key, " ") {
		return fmt.Errorf("key contains spaces")
	}
	if strings.Contains(value, " ") {
		return fmt.Errorf("value contains spaces")
	}
	bag := baggage.FromContext(r.Context())
	member, err := baggage.NewMember(key, value)
	if err != nil {
		return fmt.Errorf("failed to create baggage member: %w", err)
	}
	bag, err = bag.SetMember(member)
	if err != nil {
		return fmt.Errorf("failed to set baggage member: %w", err)
	}
	ctx := baggage.ContextWithBaggage(r.Context(), bag)
	*r = *r.WithContext(ctx)

	// Keep the currently active span in sync with baggage updates done during handling.
	span := trace.SpanFromContext(ctx)
	if span != nil && span.IsRecording() {
		span.SetAttributes(attribute.String(key, value))
	}

	return nil
}

// Internal

type otelHandler struct {
	handler     http.Handler
	serviceName string
	options     []otelhttp.Option
}

func (h *otelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := jwt.GetParsedToken(r)
	if err == nil {
		if err := AddBaggageToHTTPRequest(r, "user_id", token.Sub); err != nil {
			http.Error(w, "Failed to add user_id baggage", http.StatusInternalServerError)
			return
		}
		if err := AddBaggageToHTTPRequest(r, "username", token.Username); err != nil {
			http.Error(w, "Failed to add username baggage", http.StatusInternalServerError)
			return
		}
	}

	options := append([]otelhttp.Option{}, h.options...)
	spanAttributes := BaggageToSpanAttributes(r.Context())
	if len(spanAttributes) > 0 {
		options = append(options, otelhttp.WithSpanOptions(trace.WithAttributes(spanAttributes...)))
	}

	otelhttp.NewHandler(h.handler, h.serviceName, options...).ServeHTTP(w, r)
}
