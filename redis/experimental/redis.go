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

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// Client identifies uniquely one redis client with a pool of connections
type Client struct {
	Instance redis.UniversalClient
	Args     *ClientArgs
}

type ClientArgs struct {
	Url           string
	ClusterMode   bool
	EnableMetrics bool
	EnableTracing bool
}

// NewClient creates and returns a new redis client based on the given settings. It only supports redis 7 clients.
func NewClient(args *ClientArgs) (*Client, error) {
	client := &Client{
		Args: args,
	}

	if args.ClusterMode {
		if err := client.ConnectCluster(args.Url); err != nil {
			return nil, err
		}
	} else {
		if err := client.Connect(args.Url); err != nil {
			return nil, err
		}
	}
	if args.EnableTracing {
		if err := redisotel.InstrumentTracing(client.Instance); err != nil {
			return nil, err
		}
	}

	if args.EnableMetrics {
		if err := redisotel.InstrumentMetrics(client.Instance); err != nil {
			return nil, err
		}
	}

	err := client.Instance.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %s", err)
	}

	return client, nil
}

// Connect to Redis
func (c *Client) Connect(url string) error {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return err
	}

	c.Instance = redis.NewClient(opts)

	return nil
}

// ConnectCluster to Redis cluster
func (c *Client) ConnectCluster(url string) error {
	opts, err := redis.ParseClusterURL(url)
	if err != nil {
		return err
	}

	c.Instance = redis.NewClusterClient(opts)

	return nil
}
