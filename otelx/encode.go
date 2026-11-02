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
	"fmt"
	"strings"
)

// encodeBaggageValue percent-encodes every byte outside the W3C baggage-octets, and the percent sign,
// because baggage.NewMember decodes its input: the member's Value() is the original string again,
// and the header carries the encoded form.
func encodeBaggageValue(value string) string {
	result := strings.Builder{}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if isBaggageOctet(c) && c != '%' {
			result.WriteByte(c)
		} else {
			result.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return result.String()
}

// isBaggageOctet is printable ascii without space, double quote, comma, semicolon and backslash
func isBaggageOctet(c byte) bool {
	switch {
	case c == '!',
		c >= '#' && c <= '+',
		c >= '-' && c <= ':',
		c >= '<' && c <= '[',
		c >= ']' && c <= '~':
		return true
	}
	return false
}
