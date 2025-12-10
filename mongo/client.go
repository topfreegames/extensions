/*
 * Copyright (c) 2017 TFG Co <backend@tfgco.com>
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

package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/topfreegames/extensions/v9/mongo/interfaces"
)

// Client is the struct that connects to MongoDB
type Client struct {
	Config  *viper.Viper
	MongoDB interfaces.MongoDB
}

// NewClient connects to Mongo server and returns its client
func NewClient(prefix string, config *viper.Viper, mongoOrNil ...interfaces.MongoDB) (*Client, error) {
	client := &Client{
		Config: config,
	}
	var mongoDB interfaces.MongoDB
	if len(mongoOrNil) > 0 {
		mongoDB = mongoOrNil[0]
	}
	err := client.Connect(prefix, mongoDB)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func makeKey(prefix, suffix string) string {
	return fmt.Sprintf("%s.%s", prefix, suffix)
}

// Connect connects to mongo database and saves on Client
func (c *Client) Connect(prefix string, db interfaces.MongoDB) error {
	url := c.Config.GetString(makeKey(prefix, "url"))
	user := c.Config.GetString(makeKey(prefix, "user"))
	pass := c.Config.GetString(makeKey(prefix, "pass"))
	database := c.Config.GetString(makeKey(prefix, "database"))
	timeout := c.Config.GetDuration(makeKey(prefix, "connectionTimeout"))

	if db != nil {
		c.MongoDB = db
	} else {
		ctx := context.Background()
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		// Build connection options
		clientOpts := options.Client().ApplyURI(url)

		// Set authentication if provided
		if user != "" && pass != "" {
			clientOpts.SetAuth(options.Credential{
				Username: user,
				Password: pass,
			})
		}

		// Set timeout
		if timeout > 0 {
			clientOpts.SetServerSelectionTimeout(timeout)
			clientOpts.SetConnectTimeout(timeout)
		}

		// Connect to MongoDB
		client, err := mongo.Connect(clientOpts)
		if err != nil {
			return err
		}

		// Ping the database to verify connection
		pingCtx := context.Background()
		if timeout > 0 {
			var cancel context.CancelFunc
			pingCtx, cancel = context.WithTimeout(pingCtx, timeout)
			defer cancel()
		}

		err = client.Ping(pingCtx, nil)
		if err != nil {
			return err
		}

		c.MongoDB = NewMongo(client, client.Database(database))
	}

	return nil
}

// Close closes the connection with database
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.MongoDB.Close(ctx)
}
