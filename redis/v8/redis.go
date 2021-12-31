package redis

/*
 * Copyright (c) 2021 TFG Co <backend@tfgco.com>
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

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	redistracing "github.com/topfreegames/extensions/v9/tracing/redis/v8"
)

type ClientConfig struct {
	URL string
	// Timeout in seconds
	ConnectionTimeout int
}

// NewClient creates and returns a new redis client based on the given settings
func NewClient(ctx context.Context, prefix string, config *viper.Viper) (*redis.Client, error) {
	clientConfig := &ClientConfig{
		URL:               config.GetString(fmt.Sprintf("%s.url", prefix)),
		ConnectionTimeout: config.GetInt(fmt.Sprintf("%s.connectionTimeout", prefix)),
	}
	return NewClientFromConfig(ctx, clientConfig)
}

// NewClientFromConfig creates a Client with a ClientConfig.
func NewClientFromConfig(ctx context.Context, config *ClientConfig) (*redis.Client, error) {
	options, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}
	client := redis.NewClient(options)

	timeout := time.Duration(config.ConnectionTimeout) * time.Second
	err = waitForConnection(ctx, client, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	redistracing.Instrument(client)

	return client, nil
}

func waitForConnection(ctx context.Context, client *redis.Client, timeout time.Duration) error {
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			return errors.New("timed out waiting for Redis to connect")
		case <-ticker.C:
			if isConnected(ctx, client) {
				return nil
			}
		}
	}
}

func isConnected(ctx context.Context, client *redis.Client) bool {
	result, err := client.Ping(ctx).Result()
	if err != nil {
		return false
	}
	return result == "PONG"
}
