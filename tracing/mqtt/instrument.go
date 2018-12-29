/*
 * Copyright (c) 2018 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package mqtt

import (
	"context"
	"fmt"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/opentracing/opentracing-go"
	"github.com/topfreegames/extensions/tracing"
)

// Trace wraps an MQTT request and reports it to tracing
func Trace(ctx context.Context, method string, topic string, qos byte, timeout time.Duration, next func() mqtt.Token) {
	var parent opentracing.SpanContext

	if span := opentracing.SpanFromContext(ctx); span != nil {
		parent = span.Context()
	}

	operationName := "MQTT " + method
	reference := opentracing.ChildOf(parent)
	tags := opentracing.Tags{
		"mqtt.qos":   qos,
		"mqtt.topic": topic,

		"span.kind": "client",
	}

	span := opentracing.StartSpan(operationName, reference, tags)
	defer tracing.LogPanic(span)

	token := next()
	go wait(span, token, timeout)
}

func wait(span opentracing.Span, token mqtt.Token, timeout time.Duration) {
	defer span.Finish()

	ok := token.WaitTimeout(timeout)
	err := token.Error()

	if !ok {
		message := fmt.Sprintf("Exceded maximum expected duration: %v", timeout)
		tracing.LogError(span, message)
	}

	if err != nil {
		message := err.Error()
		tracing.LogError(span, message)
	}
}
