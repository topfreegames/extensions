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

package redis

import (
	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/topfreegames/extensions/v9/tracing"
)

// Instrument adds tracing instrumentation on a Redis client
func Instrument(client *redis.Client) {
	middleware := makeMiddleware(client)
	client.WrapProcess(middleware)
	middlewarePipe := makeMiddlewarePipe(client)
	client.WrapProcessPipeline(middlewarePipe)
}

func makeMiddleware(client *redis.Client) func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			var parent opentracing.SpanContext

			ctx := client.Context()
			if span := opentracing.SpanFromContext(ctx); span != nil {
				parent = span.Context()
			}

			operationName := "redis " + cmd.Name()
			reference := opentracing.ChildOf(parent)
			tags := opentracing.Tags{
				"db.instance":  client.Options().DB,
				"db.statement": parseLong(cmd),
				"db.type":      "redis",
				"span.kind":    "client",
			}

			span := opentracing.StartSpan(operationName, reference, tags)
			tracing.RunCustomTracingHooks(span)
			defer span.Finish()
			defer tracing.LogPanic(span)

			err := old(cmd)
			if err != nil {
				message := err.Error()
				tracing.LogError(span, message)
			}

			return err
		}
	}
}

func makeMiddlewarePipe(client *redis.Client) func(old func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(old func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			var parent opentracing.SpanContext

			ctx := client.Context()
			if span := opentracing.SpanFromContext(ctx); span != nil {
				parent = span.Context()
			}

			operationName := "redis pipe"
			statement := ""
			for idx, cmd := range cmds {
				if idx > 0 {
					statement = statement + "\n" + parseLong(cmd)
				} else {
					statement = parseLong(cmd)
				}
			}
			reference := opentracing.ChildOf(parent)
			tags := opentracing.Tags{
				"db.instance":  client.Options().DB,
				"db.statement": statement,
				"db.type":      "redis",
				"span.kind":    "client",
			}

			span := opentracing.StartSpan(operationName, reference, tags)
			tracing.RunCustomTracingHooks(span)
			defer span.Finish()
			defer tracing.LogPanic(span)

			err := old(cmds)
			if err != nil {
				message := err.Error()
				tracing.LogError(span, message)
			}

			return err
		}
	}
}

func parseLong(cmd redis.Cmder) string {
	str := cmd.String()
	return str
}
