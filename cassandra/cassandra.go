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
	ConfigPrefix string
	Config       *viper.Viper
	DB           interfaces.DB
	Session      interfaces.Session
}

// ClientParams is a wrapper for all creation parameters for a new client
type ClientParams struct {
	ConfigPrefix  string
	Config        *viper.Viper
	CqlOrNil      interfaces.DB
	SessOrNil     interfaces.Session
	ClusterConfig *gocql.ClusterConfig
}

// NewClient returns a new Cassandra client
func NewClient(params *ClientParams) (*Client, error) {
	client := &Client{
		ConfigPrefix: params.ConfigPrefix,
		Config:       params.Config,
	}
	if params.CqlOrNil != nil {
		client.DB = params.CqlOrNil
	} else if params.ClusterConfig != nil {
		params.ClusterConfig.Hosts = client.getHosts()
		params.ClusterConfig.Keyspace = client.getKeyspace()
		client.DB = params.ClusterConfig
	} else {
		client.setDefaultCluster()
	}
	if params.SessOrNil != nil {
		client.Session = params.SessOrNil
	} else {
		session, err := client.DB.CreateSession()
		if err != nil {
			return nil, err
		}
		client.Session = session
	}
	return client, nil
}

func (c *Client) getHosts() []string {
	return strings.Split(
		c.Config.GetString(fmt.Sprintf("%s.hosts", c.ConfigPrefix)), ",",
	)
}

func (c *Client) getKeyspace() string {
	return c.Config.GetString(
		fmt.Sprintf("%s.keyspace", c.ConfigPrefix),
	)
}

func (c *Client) setDefaultCluster() {
	cluster := gocql.NewCluster(c.getHosts()...)
	cluster.Keyspace = c.getKeyspace()
	c.DB = cluster
}
