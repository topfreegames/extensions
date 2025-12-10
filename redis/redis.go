/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
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
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/redis/interfaces"
	tredis "github.com/topfreegames/extensions/v9/tracing/redis"
)

// ClientConfig holds the input configs for a client.
type ClientConfig struct {
	URL               string
	ConnectionTimeout int
}

// Client identifies uniquely one redis client with a pool of connections
type Client struct {
	Client       interfaces.RedisClient
	TraceWrapper interfaces.TraceWrapper
	Config       *viper.Viper
	Options      *redis.Options
}

// TraceWrapper is the struct for the TraceWrapper
type TraceWrapper struct{}

// WithContext is a wrapper for returning a client with context
func (t *TraceWrapper) WithContext(ctx context.Context, c interfaces.RedisClient) interfaces.RedisClient {
	// In v9, context is passed per-command, so just return the same client
	return c
}

// NewClientFromConfig creates a Client with a ClientConfig.
func NewClientFromConfig(config *ClientConfig, ifaces ...interface{}) (*Client, error) {
	viperConfig := viper.New()
	viperConfig.Set("prefix.url", config.URL)
	viperConfig.Set("prefix.connectionTimeout", config.ConnectionTimeout)
	return NewClient("prefix", viperConfig, ifaces...)
}

// NewClient creates and returns a new redis client based on the given settings
func NewClient(prefix string, config *viper.Viper, ifaces ...interface{}) (*Client, error) {
	client := &Client{
		Config: config,
	}

	var cl interfaces.RedisClient
	if len(ifaces) > 0 && ifaces[0] != nil {
		if v, ok := ifaces[0].(interfaces.RedisClient); ok {
			cl = v
		}
	}
	if len(ifaces) > 1 && ifaces[1] != nil {
		if v, ok := ifaces[1].(interfaces.TraceWrapper); ok {
			client.TraceWrapper = v
		}
	}
	err := client.Connect(prefix, cl)
	if err != nil {
		return nil, err
	}

	timeout := config.GetInt(fmt.Sprintf("%s.connectionTimeout", prefix))
	err = client.WaitForConnection(timeout)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Trace creates a Redis client that sends traces to tracing
func (c *Client) Trace(ctx context.Context) interfaces.RedisClient {
	if c.TraceWrapper != nil {
		return c.TraceWrapper.WithContext(ctx, c.Client)
	}
	// In v9, instrumentation is added once to the client, context is passed per-command
	if redisClient, ok := c.Client.(*redis.Client); ok {
		tredis.Instrument(redisClient)
	}
	return c.Client
}

// Connect to Redis
func (c *Client) Connect(prefix string, client interfaces.RedisClient) error {
	var err error
	c.Options, err = redis.ParseURL(c.Config.GetString(fmt.Sprintf("%s.url", prefix)))
	if err != nil {
		return err
	}

	if client == nil {
		c.Client = redis.NewClient(c.Options)
	} else {
		c.Client = client
	}

	return nil
}

// IsConnected determines if the client is connected to redis
func (c *Client) IsConnected() bool {
	ctx := context.Background()
	result := c.Client.Ping(ctx)
	if result != nil {
		res, err := result.Result()
		if err != nil {
			return false
		}
		return res == "PONG"
	}
	return true
}

// EnterCriticalSection locks key using redlock algorithm
func (c *Client) EnterCriticalSection(client interfaces.RedisClient, key string, lockTimeout, retryTimeout, retryInterval time.Duration) (*redislock.Lock, error) {
	ctx := context.Background()

	// Cast to *redis.Client for redislock compatibility
	redisClient, ok := client.(*redis.Client)
	if !ok {
		return nil, fmt.Errorf("client must be *redis.Client for lock operations")
	}

	locker := redislock.New(redisClient)

	// Create options
	opts := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(retryInterval), int(retryTimeout/retryInterval)),
	}

	return locker.Obtain(ctx, key, lockTimeout, opts)
}

// LeaveCriticalSection unlocks key
func (c *Client) LeaveCriticalSection(lock *redislock.Lock) error {
	ctx := context.Background()
	return lock.Release(ctx)
}

// WaitForConnection loops until redis is connected
func (c *Client) WaitForConnection(timeout int) error {
	t := time.Duration(timeout) * time.Second
	timeoutTimer := time.NewTimer(t)
	defer timeoutTimer.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			return fmt.Errorf("timed out waiting for Redis to connect")
		case <-ticker.C:
			if c.IsConnected() {
				return nil
			}
		}
	}
}

// Close the connections to redis
func (c *Client) Close() error {
	err := c.Client.Close()
	if err != nil {
		return err
	}
	return nil
}

// Cleanup closes redis connection
func (c *Client) Cleanup() error {
	err := c.Close()
	if err != nil {
		return err
	}
	return nil
}
