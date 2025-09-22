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
	"net"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
)

const callerKey = "caller"

type CallerGenerator struct {
	mux                *sync.RWMutex
	cache              map[string]string
	errorLogger        *slog.Logger
	structAttrProvider structAttrProvider
}

func NewCallerGenerator(errorLogger *slog.Logger, structAttrProvider structAttrProvider) *CallerGenerator {
	return &CallerGenerator{
		mux:                &sync.RWMutex{},
		cache:              map[string]string{},
		errorLogger:        errorLogger,
		structAttrProvider: structAttrProvider,
	}
}

func (c *CallerGenerator) Generate(gc *gin.Context) (string, any) {
	remote := gc.RemoteIP()
	c.mux.RLock()
	result, ok := c.cache[remote]
	c.mux.RUnlock()
	if ok {
		return callerKey, result
	}
	remoteHosts, err := net.LookupAddr(remote)
	if err != nil {
		c.errorLogger.Warn("could not perform reverse DNS lookup", c.structAttrProvider.ErrorKey(), err)
		return callerKey, remote
	}
	if len(remoteHosts) > 0 {
		sort.Strings(remoteHosts)
		result = remoteHosts[0]
	} else {
		result = remote
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.cache[remote] = result
	return callerKey, result
}
