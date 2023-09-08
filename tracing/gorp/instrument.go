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

package gorp

import (
	"context"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/topfreegames/extensions/v9/tracing"
)

// Trace wraps a gorp database call and reports it to tracing
func Trace(ctx context.Context, name string, query string, next func() error) {
	var parent opentracing.SpanContext

	if ctx == nil {
		ctx = context.Background()
	}

	if span := opentracing.SpanFromContext(ctx); span != nil {
		parent = span.Context()
	}

	operationName := "SQL " + parse(query)
	reference := opentracing.ChildOf(parent)
	tags := opentracing.Tags{
		"db.instance":  name,
		"db.statement": query,
		"db.type":      "postgres",

		"span.kind": "client",
	}

	span := opentracing.StartSpan(operationName, reference, tags)
	tracing.RunCustomTracingHooks(span)
	defer span.Finish()
	defer tracing.LogPanic(span)

	err := next()
	if err != nil {
		message := err.Error()
		tracing.LogError(span, message)
	}
}

func parse(query string) string {
	re := regexp.MustCompile("\\s+")
	array := re.Split(" "+query, 3)
	command := array[1]
	return strings.ToUpper(command)
}
