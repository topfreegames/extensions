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

package pg

import (
	"context"
	"fmt"
	"time"

	pg "github.com/go-pg/pg"

	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/pg/interfaces"
)

// Client is the struct that connects to PostgreSQL
type Client struct {
	Config     *viper.Viper
	DB         interfaces.DB
	Options    *pg.Options
	TxWrapper  interfaces.TxWrapper
	CtxWrapper interfaces.CtxWrapper
	context    context.Context
}

// TxWrapper is the struct for the TxWrapper
type TxWrapper struct{}

// DbBegin is a wrapper for returning db transactions
func (t *TxWrapper) DbBegin(db interfaces.DB) (interfaces.Tx, error) {
	return db.Begin()
}

// CtxWrapper is the struct for the CTxWrapper
type CtxWrapper struct{}

// WithContext is a wrapper for returning db with a given context
func (t *CtxWrapper) WithContext(ctx context.Context, db interfaces.DB) interfaces.DB {
	return WithContext(ctx, db)
}

// NewClient creates a new client
func NewClient(prefix string, config *viper.Viper, dbIfaces ...interface{}) (*Client, error) {
	client := &Client{
		Config:     config,
		TxWrapper:  &TxWrapper{},
		CtxWrapper: &CtxWrapper{},
	}

	var db interfaces.DB
	if len(dbIfaces) > 0 && dbIfaces[0] != nil {
		if v, ok := dbIfaces[0].(interfaces.DB); ok {
			db = v
		}
	}
	if len(dbIfaces) > 1 && dbIfaces[1] != nil {
		if v, ok := dbIfaces[1].(interfaces.TxWrapper); ok {
			client.TxWrapper = v
		}

	}
	if len(dbIfaces) > 2 && dbIfaces[2] != nil {
		if v, ok := dbIfaces[2].(interfaces.CtxWrapper); ok {
			client.CtxWrapper = v
		}
	}

	err := client.Connect(prefix, db)
	if err != nil {
		return nil, err
	}

	if db == nil {
		timeout := config.GetInt(fmt.Sprintf("%s.connectionTimeout", prefix))
		err = client.WaitForConnection(timeout)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

// Connect to PG
func (c *Client) Connect(prefix string, db interfaces.DB) error {
	user := c.Config.GetString(fmt.Sprintf("%s.user", prefix))
	pass := c.Config.GetString(fmt.Sprintf("%s.pass", prefix))
	host := c.Config.GetString(fmt.Sprintf("%s.host", prefix))
	database := c.Config.GetString(fmt.Sprintf("%s.database", prefix))
	port := c.Config.GetInt(fmt.Sprintf("%s.port", prefix))
	poolSize := c.Config.GetInt(fmt.Sprintf("%s.poolSize", prefix))
	maxRetries := c.Config.GetInt(fmt.Sprintf("%s.maxRetries", prefix))

	c.Options = &pg.Options{
		Addr:       fmt.Sprintf("%s:%d", host, port),
		User:       user,
		Password:   pass,
		Database:   database,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
	}

	if db == nil {
		c.DB = &DB{inner: pg.Connect(c.Options)}
	} else {
		c.DB = &DB{inner: db}
	}

	return nil
}

// IsConnected determines if the client is connected to PG
func (c *Client) IsConnected() bool {
	res, err := c.DB.Exec("select 1")
	if err != nil {
		return false
	}
	return res.RowsReturned() == 1
}

// Close the connections to PG
func (c *Client) Close() error {
	err := c.DB.Close()
	if err != nil {
		return err
	}
	return nil
}

// WaitForConnection loops until PG is connected
func (c *Client) WaitForConnection(timeout int) error {
	t := time.Duration(timeout) * time.Second
	timeoutTimer := time.NewTimer(t)
	defer timeoutTimer.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			return fmt.Errorf("timed out waiting for PostgreSQL to connect")
		case <-ticker.C:
			if c.IsConnected() {
				return nil
			}
		}
	}
}

//Cleanup closes PG connection
func (c *Client) Cleanup() error {
	err := c.Close()
	return err
}
