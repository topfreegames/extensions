package redis

/*
 * Copyright (c) 2021 TFG Co <backend@tfgco.com>
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

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/opentracing/opentracing-go"
	"github.com/topfreegames/extensions/v9/tracing"
)

type redisTracingHook struct {
	client *redis.Client
}

// Instrument adds tracing instrumentation on a Redis client
func Instrument(client *redis.Client) {
	client.AddHook(redisTracingHook{client: client})
}

func (hook redisTracingHook) createSpan(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	tags := opentracing.Tags{
		"db.instance": hook.client.Options().DB,
		"db.type":     "redis",
		"span.kind":   "client",
	}
	tags = tracing.RunCustomTracingTagsHooks(ctx, tags)
	span := opentracing.SpanFromContext(ctx)
	if span != nil {

		childSpan := opentracing.StartSpan(operationName, opentracing.ChildOf(span.Context()), tags)
		tracing.RunCustomTracingHooks(ctx, operationName, childSpan)
		return childSpan, opentracing.ContextWithSpan(ctx, childSpan)
	}

	return opentracing.StartSpanFromContext(ctx, operationName, tags)
}

func (hook redisTracingHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (hook redisTracingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		operationName := "redis " + cmd.Name()
		span, ctxWithSpan := hook.createSpan(ctx, operationName)
		span.SetTag("db.statement", cmd.String())
		defer span.Finish()

		err := next(ctxWithSpan, cmd)
		if err != nil {
			tracing.LogError(span, err.Error())
		}
		return err
	}
}

func (hook redisTracingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		operationName := "redis pipe"
		span, ctxWithSpan := hook.createSpan(ctx, operationName)
		span.SetTag("db.statement", cmdersString(cmds))
		defer span.Finish()

		err := next(ctxWithSpan, cmds)
		errorIndex, cmdErr := cmdersError(cmds)
		if cmdErr != nil {
			tracing.LogError(span, fmt.Errorf("pipeline error %v: %w", errorIndex, cmdErr).Error())
		}
		return err
	}
}

func cmdersString(cmds []redis.Cmder) string {
	strs := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		strs = append(strs, cmd.String())
	}
	return strings.Join(strs, "\n")
}

func cmdersError(cmds []redis.Cmder) (int, error) {
	for index, cmd := range cmds {
		if cmd.Err() != nil {
			return index, cmd.Err()
		}
	}
	return 0, nil
}
