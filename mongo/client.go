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
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/mongo/interfaces"
)

// Client is the struct that connects to PostgreSQL
type Client struct {
	Config  *viper.Viper
	MongoDB interfaces.MongoDB
}

//NewClient connects to Mongo server and return its client
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

func makeKey(prefix, sufix string) string {
	return fmt.Sprintf("%s.%s", prefix, sufix)
}

//Connect connects to mongo database and saves on Client
func (c *Client) Connect(prefix string, db interfaces.MongoDB) error {
	url := c.Config.GetString(makeKey(prefix, "url"))
	user := c.Config.GetString(makeKey(prefix, "user"))
	pass := c.Config.GetString(makeKey(prefix, "pass"))
	database := c.Config.GetString(makeKey(prefix, "database"))
	timeout := c.Config.GetDuration(makeKey(prefix, "connectionTimeout"))

	if db != nil {
		c.MongoDB = db
	} else {
		var session *mgo.Session
		var err error

		if timeout > 0 {
			session, err = mgo.DialWithTimeout(url, timeout)
		} else {
			session, err = mgo.Dial(url)
		}

		if err != nil {
			return err
		}
		mongoDB := session.DB(database)
		if user != "" && pass != "" {
			err = mongoDB.Login(user, pass)
			if err != nil {
				return err
			}
		}
		c.MongoDB = NewMongo(session, mongoDB)
	}

	return nil
}

//Close closes the session and the connection with database
func (c *Client) Close() {
	c.MongoDB.Close()
}
