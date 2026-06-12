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

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectContextToRequest adds context to an outbound HTTP request and injects
// propagation headers (tracecontext and baggage) for downstream services.
func InjectContextToRequest(ctx context.Context, req *http.Request) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if gc, ok := ctx.(*gin.Context); ok {
		if gc.Request == nil {
			return fmt.Errorf("gin request is nil")
		}
		ctx = gc.Request.Context()
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req = req.WithContext(ctx)
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, propagation.HeaderCarrier(req.Header))
	return nil
}
