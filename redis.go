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

package extensions

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/interfaces"
)

// RedisClient identifies uniquely one redis client with a pool of connections
type RedisClient struct {
	Client  interfaces.RedisClient
	Config  *viper.Viper
	Options *redis.Options
}

// NewRedisClient creates and returns a new redis client based on the given settings
func NewRedisClient(prefix string, config *viper.Viper, clientOrNil ...interfaces.RedisClient) (*RedisClient, error) {
	client := &RedisClient{
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

	timeout := config.GetInt("extensions.pg.connectionTimeout")
	err = client.WaitForConnection(timeout)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Connect to Redis
func (c *RedisClient) Connect(prefix string, client interfaces.RedisClient) error {
	host := c.Config.GetString(fmt.Sprintf("%s.host", prefix))
	port := c.Config.GetInt(fmt.Sprintf("%s.port", prefix))
	pass := c.Config.GetString(fmt.Sprintf("%s.pass", prefix))
	database := c.Config.GetInt(fmt.Sprintf("%s.database", prefix))
	poolSize := c.Config.GetInt(fmt.Sprintf("%s.poolSize", prefix))

	c.Options = &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: pass,
		DB:       database,
		PoolSize: poolSize,
	}

	if client == nil {
		c.Client = redis.NewClient(c.Options)
	} else {
		c.Client = client
	}

	return nil
}

// IsConnected determines if the client is connected to redis
func (c *RedisClient) IsConnected() bool {
	res, err := c.Client.Ping().Result()
	if err != nil {
		return false
	}
	return res == "PONG"
}

// WaitForConnection loops until redis is connected
func (c *RedisClient) WaitForConnection(timeout int) error {
	start := time.Now().UnixNano() / 1000000
	t := int64(timeout)

	ellapsed := func() int64 {
		return (time.Now().UnixNano() / 1000000) - start
	}

	for !c.IsConnected() && ellapsed() <= t {
		time.Sleep(10 * time.Millisecond)
	}

	if ellapsed() > t {
		return fmt.Errorf("Timed out waiting for Redis to connect.")
	}

	return nil
}

// Close the connections to redis
func (c *RedisClient) Close() error {
	err := c.Client.Close()
	if err != nil {
		return err
	}
	return nil
}

// Cleanup closes redis connection
func (c *RedisClient) Cleanup() error {
	err := c.Close()
	if err != nil {
		return err
	}
	return nil
}
