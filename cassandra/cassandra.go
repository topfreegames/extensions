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

package cassandra

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/cassandra/interfaces"
)

// Client is the struct that connects to Cassandra
type Client struct {
	Config  *viper.Viper
	DB      interfaces.DB
	Session interfaces.Session
}

// NewClient returns a new Cassandra client
func NewClient(prefix string, config *viper.Viper, cqlOrNil interfaces.DB, sessOrNil interfaces.Session) (*Client, error) {
	client := &Client{
		Config: config,
	}

	var db interfaces.DB
	if cqlOrNil != nil {
		db = cqlOrNil
	}
	err := client.Connect(prefix, db)
	if err != nil {
		return nil, err
	}
	if sessOrNil != nil {
		client.Session = sessOrNil
	} else {
		session := c.DB.CreateSession()
		client.Session = session
	}
	return client, nil
}

// Connect connects to Cassandra cluster
func (c *Client) Connect(prefix string, db interfaces.DB) error {
	if db != nil {
		c.DB = db
		return nil
	}
	hosts := strings.Split(c.Config.GetString(fmt.Sprintf("%s.hosts", prefix)), ",")
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = c.Config.GetString(fmt.Sprintf("%s.keyspace", prefix))
	cluster.Consistency = gocql.Quorum
	c.DB = cluster

	return nil
}
