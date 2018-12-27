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

package http

import (
	"net/http"

	tracing "github.com/topfreegames/extensions/tracing/http"
)

// New creates and instruments an HTTP client
func New() *http.Client {
	client := &http.Client{}
	Instrument(client)
	return client
}

// Instrument instruments the internal transport object
func Instrument(client *http.Client) {
	inner := getTransport(client)
	client.Transport = &Transport{inner}
}

func getTransport(client *http.Client) http.RoundTripper {
	if client.Transport == nil {
		return http.DefaultTransport
	}
	return client.Transport
}

// Transport instruments another RoundTripper
type Transport struct {
	inner http.RoundTripper
}

// RoundTrip executes a single HTTP transaction
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	tracing.Trace(req, func() error {
		resp, err = t.inner.RoundTrip(req)
		return err
	})

	return resp, err
}
