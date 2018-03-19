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

	"github.com/eclipse/paho.mqtt.golang"
	jaeger "github.com/topfreegames/extensions/jaeger/mqtt"
	"github.com/topfreegames/extensions/mqtt/interfaces"
)

// Client wraps an MQTT client
type Client struct {
	ctx   context.Context
	inner mqtt.Client
	opts  *mqtt.ClientOptions
}

// NewClient will create an MQTT client with all of the options specified
func NewClient(opts *mqtt.ClientOptions) interfaces.Client {
	ctx := context.Background()
	inner := mqtt.NewClient(opts)
	return &Client{ctx, inner, opts}
}

func (c *Client) WithContext(ctx context.Context) interfaces.Client {
	if ctx == nil {
		panic("Context must be non-nil")
	}
	return &Client{ctx, c.inner, c.opts}
}

// Publish will publish a message with the specified QoS and content
func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	var token mqtt.Token

	jaeger.Trace(c.ctx, "PUBLISH", topic, qos, c.opts.PingTimeout, func() mqtt.Token {
		token = c.inner.Publish(topic, qos, retained, payload)
		return token
	})

	return token
}

// Subscribe starts a new subscription
func (c *Client) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	var token mqtt.Token

	jaeger.Trace(c.ctx, "SUBSCRIBE", topic, qos, c.opts.PingTimeout, func() mqtt.Token {
		token = c.inner.Subscribe(topic, qos, callback)
		return token
	})

	return token
}

// Connect will create a connection to the message broker
func (c *Client) Connect() mqtt.Token {
	return c.inner.Connect()
}

// IsConnected returns a bool signifying whether the client is connected or not
func (c *Client) IsConnected() bool {
	return c.inner.IsConnected()
}
