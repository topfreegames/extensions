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
	"fmt"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
	"github.com/topfreegames/extensions/jaeger"
)

// Instrument adds Jaeger instrumentation on an Echo app
func Instrument(app *echo.Echo) {
	middleware := makeMiddleware()
	app.Use(middleware)
}

func makeMiddleware() func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tracer := opentracing.GlobalTracer()

			request := c.Request()
			route := c.Path()

			method := request.Method()
			url := request.URL()

			header := getCarrier(request)
			parent, _ := tracer.Extract(opentracing.HTTPHeaders, header)

			operationName := fmt.Sprintf("HTTP %s %s", method, route)
			reference := opentracing.ChildOf(parent)
			tags := opentracing.Tags{
				"http.method":   method,
				"http.host":     request.Host(),
				"http.pathname": url.Path(),
				"http.query":    url.QueryString(),

				"span.kind": "server",
			}

			span := opentracing.StartSpan(operationName, reference, tags)
			defer span.Finish()
			defer jaeger.LogPanic(span)

			ctx := c.StdContext()
			ctx = opentracing.ContextWithSpan(ctx, span)
			c.SetStdContext(ctx)

			err := next(c)
			if err != nil {
				message := err.Error()
				jaeger.LogError(span, message)
			}

			response := c.Response()
			statusCode := response.Status()

			span.SetTag("http.status_code", statusCode)

			return err
		}
	}
}

func getCarrier(request engine.Request) opentracing.HTTPHeadersCarrier {
	if header, ok := request.Header().(*standard.Header); ok {
		return opentracing.HTTPHeadersCarrier(header.Header)
	}
	return nil
}
