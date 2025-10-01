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
	"log/slog"

	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	userInfoKey = "userinfo"
	userNameKey = "name"
	userIdKey   = "id"
)

type UserInfoGenerator struct {
	errorLogger        *slog.Logger
	structAttrProvider structAttrProvider
}

func NewUserInfoGenerator(errorLogger *slog.Logger, structAttrProvider structAttrProvider) *UserInfoGenerator {
	return &UserInfoGenerator{
		errorLogger:        errorLogger,
		structAttrProvider: structAttrProvider,
	}
}

func (c *UserInfoGenerator) Generate(gc *gin.Context) (string, any) {
	token, err := jwt.GetParsedToken(gc.Request)
	if err != nil {
		c.errorLogger.Debug("could not decode token", c.structAttrProvider.ErrorKey(), err)
		return userInfoKey, map[string]string{
			userIdKey: gc.Request.Header.Get("X-User-Id"),
		}
	}
	return userInfoKey, map[string]string{
		userIdKey:   token.Sub,
		userNameKey: token.Username,
	}
}
