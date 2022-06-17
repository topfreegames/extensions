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

package jaeger

import (
	"io"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	opentracing "github.com/opentracing/opentracing-go"
)

// Options holds configuration options for Jaeger
type Options struct {
	Disabled    bool
	Probability float64
	ServiceName string
}

// Configure configures a global Jaeger tracer
func Configure(options Options) (io.Closer, error) {
	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		cfg = config.Configuration{
			Disabled: options.Disabled,
			Sampler: &config.SamplerConfig{
				Type:  jaeger.SamplerTypeProbabilistic,
		 		Param: options.Probability,
			},
		}
	} else {
		if cfg.ServiceName == "" {
			cfg.ServiceName = options.ServiceName
		}
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, err
	}
        opentracing.SetGlobalTracer(tracer)
	return closer, nil
}
