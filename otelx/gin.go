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

	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/trace"
)

// GinOpenTelemetry initializes OpenTelemetry with the given service name and options, and returns a gin.HandlerFunc for tracing incoming requests.
// OpenTelemetry is initialized only once, subsequent calls will return a handler without reinitializing OpenTelemetry.
// endpoint might be empty, in that case a default will be used.
func GinOpenTelemetry(ctx context.Context, serviceName string, endpoint string, options ...otelgin.Option) (gin.HandlerFunc, error) {
	err := initOpenTelemetry(ctx, serviceName, endpoint)
	if err != nil {
		return nil, err
	}

	propagator := otel.GetTextMapPropagator()

	handler := func(c *gin.Context) {
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		*c.Request = *c.Request.WithContext(ctx)

		token, err := jwt.GetParsedToken(c.Request)
		if err == nil {
			if err := AddBaggageToGinContext(c, "user_id", token.Sub); err != nil {
				_ = c.Error(fmt.Errorf("failed to add user_id baggage: %w", err))
			}
			if err := AddBaggageToGinContext(c, "username", token.Username); err != nil {
				_ = c.Error(fmt.Errorf("failed to add username baggage: %w", err))
			}
		}

		requestOptions := append([]otelgin.Option{}, options...)
		spanAttributes := BaggageToSpanAttributes(c.Request.Context())
		if len(spanAttributes) > 0 {
			requestOptions = append(requestOptions, otelgin.WithSpanStartOptions(trace.WithAttributes(spanAttributes...)))
		}

		otelgin.Middleware(serviceName, requestOptions...)(c)
	}

	return handler, nil
}

func AddBaggageToGinContext(c *gin.Context, key, value string) error {
	return AddBaggageToHTTPRequest(c.Request, key, value)
}
