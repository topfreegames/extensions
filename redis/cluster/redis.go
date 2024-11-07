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

package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// Client identifies uniquely one redis client with a pool of connections
type Client struct {
	Instance    redis.UniversalClient
	ClusterMode bool
}

type ClientArgs struct {
	Url           string
	ClusterMode   bool
	EnableMetrics bool
	EnableTracing bool
}

// NewClient creates and returns a new redis client based on the given settings. It only supports redis 7 engine and uses go-redis v9.
func NewClient(args *ClientArgs) (*Client, error) {
	client := &Client{
		ClusterMode: args.ClusterMode,
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

	return client, nil
}

// NewClientFromConfig creates and returns a new redis client based on the given settings from predefined env vars. It only supports redis 7 engine and uses go-redis v9.
func NewClientFromConfig(config *viper.Viper, prefix string) (*Client, error) {
	args := CreateClientArgs(config, prefix)

	return NewClient(args)
}

func CreateClientArgs(config *viper.Viper, prefix string) *ClientArgs {
	url := config.GetString(fmt.Sprintf("%s.url", prefix))

	if url == "" {
		tls := config.GetBool(fmt.Sprintf("%s.tls", prefix))

		endpoint := config.GetString(fmt.Sprintf("%s.endpoint", prefix))
		port := config.GetString(fmt.Sprintf("%s.port", prefix))
		pass := config.GetString(fmt.Sprintf("%s.password", prefix))

		var baseUrl string
		if tls {
			baseUrl = "rediss://"
		} else {
			baseUrl = "redis://"
		}

		url = fmt.Sprintf("%s:%s@%s:%s", baseUrl, pass, endpoint, port)
	}

	clusterMode := config.GetBool(fmt.Sprintf("%s.clusterMode", prefix))
	enableMetrics := config.GetBool(fmt.Sprintf("%s.enableMetrics", prefix))
	enableTracing := config.GetBool(fmt.Sprintf("%s.enableTracing", prefix))

	// Extra parameters
	var queryParameters []string
	maxRedirects := config.GetString(fmt.Sprintf("%s.maxRedirects", prefix))
	if maxRedirects != "" {
		queryParameters = append(queryParameters, fmt.Sprintf("max_redirects=%s", maxRedirects))
	}

	poolsize := config.GetString("poolSize")
	if poolsize != "" {
		queryParameters = append(queryParameters, fmt.Sprintf("pool_size=%s", poolsize))
	}

	readTimeout := config.GetString(fmt.Sprintf("%s.timeout", prefix))
	if readTimeout != "" {
		queryParameters = append(queryParameters, fmt.Sprintf("read_timeout=%s", readTimeout))
	}

	if len(queryParameters) > 0 {
		url = fmt.Sprintf("%s?%s", url, strings.Join(queryParameters, "&"))
	}

	return &ClientArgs{
		ClusterMode:   clusterMode,
		Url:           url,
		EnableMetrics: enableMetrics,
		EnableTracing: enableTracing,
	}
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

// IsConnected determines if the client is connected to redis
func (c *Client) IsConnected(ctx context.Context) (bool, error) {
	result := c.Instance.Ping(ctx)
	if result != nil {
		res, err := result.Result()
		if err != nil {
			return false, err
		}
		return res == "PONG", nil
	}
	return true, nil
}
