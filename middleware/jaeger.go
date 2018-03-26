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

package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/topfreegames/extensions/jaeger"
)

// Jaeger is a middleware for jaeger tracing
func Jaeger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := opentracing.GlobalTracer()
			route, _ := mux.CurrentRoute(r).GetPathTemplate()
			header := opentracing.HTTPHeadersCarrier(r.Header)

			parent, _ := tracer.Extract(opentracing.HTTPHeaders, header)
			operationName := fmt.Sprintf("HTTP %s %s", r.Method, route)
			reference := opentracing.ChildOf(parent)
			tags := opentracing.Tags{
				"http.method":   r.Method,
				"http.host":     r.Host,
				"http.pathname": r.URL.Path,
				"http.query":    r.URL.RawQuery,
				"span.kind":     "server",
			}

			span := opentracing.StartSpan(operationName, reference, tags)
			defer span.Finish()
			defer jaeger.LogPanic(span)
			defer func() {
				status := GetStatusCode(w)
				span.SetTag("http.status_code", status)
			}()

			ctx := opentracing.ContextWithSpan(r.Context(), span)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
