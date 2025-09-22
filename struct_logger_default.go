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

package gin_mw

import (
	"log/slog"

	gen "github.com/SENERGY-Platform/gin-middleware/generators"
	"github.com/gin-gonic/gin"
)

func StructLoggerHandlerWithDefaultGenerators(structLogger *slog.Logger, structAttrProvider structAttrProvider, skipPaths []string, skipper gin.Skipper, generators ...func(*gin.Context) (string, any)) gin.HandlerFunc {
	if generators == nil {
		generators = []func(*gin.Context) (string, any){}
	}
	generators = append(generators, gen.DefaultGenerators(structLogger, structAttrProvider)...)
	return StructLoggerHandler(structLogger, structAttrProvider, skipPaths, skipper, generators...)
}
