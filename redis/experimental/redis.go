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
	tredis "github.com/topfreegames/extensions/v9/tracing/redis/experimental"
	"time"

	lock "github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/redis/experimental/instances"
	interfaces "github.com/topfreegames/extensions/v9/redis/interfaces/experimental"
)

// Client identifies uniquely one redis client with a pool of connections
type Client struct {
	instance     interfaces.RedisInstance
	traceWrapper interfaces.TraceWrapper
	config       *viper.Viper
	clusterMode  bool
}

// TraceWrapper is the struct for the TraceWrapper
type TraceWrapper struct{}

// WithContext is a wrapper for returning a client with context
func (t *TraceWrapper) WithContext(ctx context.Context, c interfaces.RedisInstance) interfaces.RedisInstance {
	return c.WithContext(ctx)
}

// NewClient creates and returns a new redis client based on the given settings
func NewClient(prefix string, config *viper.Viper, clusterMode bool) (*Client, error) {
	client := &Client{
		config:      config,
		clusterMode: clusterMode,
	}

	err := client.Connect(prefix)
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

func (c *Client) Instance() interfaces.RedisInstance {
	return c.instance
}

func (c *Client) SetupTraceWrapper(traceWrapper interfaces.TraceWrapper) {
	c.traceWrapper = traceWrapper
}

// Trace creates a Redis client that sends traces to tracing
func (c *Client) Trace(ctx context.Context) interfaces.RedisInstance {
	if c.traceWrapper != nil {
		return c.traceWrapper.WithContext(ctx, c.instance)
	}
	copiedInstance := c.instance.WithContext(ctx)
	tredis.Instrument(copiedInstance)
	return copiedInstance
}

// Connect to Redis
func (c *Client) Connect(prefix string) error {
	if c.clusterMode {
		urls := c.config.GetStringSlice(fmt.Sprintf("%s.urls", prefix))
		pass := c.config.GetString(fmt.Sprintf("%s.pass"))

		options := &redis.ClusterOptions{
			Addrs:    urls,
			Password: pass,
		}

		c.instance = instances.NewRedisClusterClientInstance(redis.NewClusterClient(options))

		return nil
	}

	options, err := redis.ParseURL(c.config.GetString(fmt.Sprintf("%s.url", prefix)))
	if err != nil {
		return err
	}
	c.instance = instances.NewRedisClientInstance(redis.NewClient(options))

	return nil
}

// IsConnected determines if the client is connected to redis
func (c *Client) IsConnected() bool {
	result := c.instance.Ping()
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
func (c *Client) EnterCriticalSection(instance interfaces.RedisInstance, key string, lockTimeout, retryTimeout, retryInterval time.Duration) (*lock.Lock, error) {
	lockOptions := &lock.LockOptions{
		LockTimeout: lockTimeout,
		WaitTimeout: retryTimeout,
		WaitRetry:   retryInterval,
	}
	return lock.ObtainLock(instance, key, lockOptions)
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
	err := c.instance.Close()
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
