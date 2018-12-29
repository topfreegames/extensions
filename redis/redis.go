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

	"github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/redis/interfaces"
	tredis "github.com/topfreegames/extensions/tracing/redis"
)

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
	return c.WithContext(ctx)
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
	copy := c.Client.WithContext(ctx)
	tredis.Instrument(copy)
	return copy
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
	result := c.Client.Ping()
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
func (c *Client) EnterCriticalSection(client interfaces.RedisClient, key string, lockTimeout, retryTimeout, retryInterval time.Duration) (*lock.Lock, error) {
	lockOptions := &lock.LockOptions{
		LockTimeout: lockTimeout,
		WaitTimeout: retryTimeout,
		WaitRetry:   retryInterval,
	}
	return lock.ObtainLock(client, key, lockOptions)
}

// LeaveCriticalSection unlocks key
func (c *Client) LeaveCriticalSection(lock *lock.Lock) error {
	return lock.Unlock()
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
