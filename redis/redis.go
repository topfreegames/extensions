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
	"github.com/topfreegames/extensions/jaeger"
	"github.com/topfreegames/extensions/redis/interfaces"
)

// Client identifies uniquely one redis client with a pool of connections
type Client struct {
	Client  interfaces.RedisClient
	Config  *viper.Viper
	Options *redis.Options
}

// NewClient creates and returns a new redis client based on the given settings
func NewClient(prefix string, config *viper.Viper, clientOrNil ...interfaces.RedisClient) (*Client, error) {
	client := &Client{
		Config: config,
	}

	var cl interfaces.RedisClient
	if len(clientOrNil) == 1 {
		cl = clientOrNil[0]
	}
	err := client.Connect(prefix, cl)
	if err != nil {
		return nil, err
	}

	client.Client = client.Trace(nil)

	timeout := config.GetInt(fmt.Sprintf("%s.connectionTimeout", prefix))
	err = client.WaitForConnection(timeout)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Trace(ctx context.Context) interfaces.RedisClient {
	if ctx == nil {
		ctx = context.Background()
	}
	copy := c.Client.WithContext(ctx)
	jaeger.InstrumentRedis(copy)
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
