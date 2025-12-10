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
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"

	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/pg/interfaces"
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
func (t *TxWrapper) DbBegin(db interfaces.DB) (interfaces.DB, error) {
	return Begin(db)
}

// CtxWrapper is the struct for the CTxWrapper
type CtxWrapper struct{}

// WithContext is a wrapper for returning db with a given context
func (t *CtxWrapper) WithContext(ctx context.Context, db interfaces.DB) interfaces.DB {
	return WithContext(ctx, db)
}

// NewClient creates a new client
func NewClient(prefix string, config *viper.Viper, dbIfaces ...interface{}) (*Client, error) {
	client := &Client{Config: config}

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
	sslMode := c.Config.GetString(fmt.Sprintf("%s.sslMode", prefix))

	var tlsConfig *tls.Config
	switch sslMode {
	case "allow", "prefer", "require":
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	default:
		tlsConfig = nil
	}

	c.Options = &pg.Options{
		Addr:       fmt.Sprintf("%s:%d", host, port),
		User:       user,
		Password:   pass,
		Database:   database,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
		TLSConfig:  tlsConfig,
	}

	// Default is 5 seconds
	dialTimeout := c.Config.GetDuration(fmt.Sprintf("%s.dialTimeout", prefix))
	if dialTimeout > 0 {
		c.Options.DialTimeout = dialTimeout
	}
	readTimeout := c.Config.GetDuration(fmt.Sprintf("%s.readTimeout", prefix))
	if readTimeout > 0 {
		c.Options.ReadTimeout = readTimeout
	}
	writeTimeout := c.Config.GetDuration(fmt.Sprintf("%s.writeTimeout", prefix))
	if writeTimeout > 0 {
		c.Options.WriteTimeout = writeTimeout
	}
	// Default is 5 minutes
	idleTimeout := c.Config.GetDuration(fmt.Sprintf("%s.idleTimeout", prefix))
	if idleTimeout > 0 {
		c.Options.IdleTimeout = idleTimeout
	}
	// Default is 1 minute, -1 disables idle connections reaper
	idleCheckFrequency := c.Config.GetDuration(fmt.Sprintf("%s.idleCheckFrequency", prefix))
	if idleCheckFrequency > 0 || idleCheckFrequency == -1 {
		c.Options.IdleCheckFrequency = idleCheckFrequency
	}
	// Default is 30 seconds if ReadTimeOut is not defined, otherwise,
	// ReadTimeout + 1 second.
	poolTimeout := c.Config.GetDuration(fmt.Sprintf("%s.poolTimeout", prefix))
	if poolTimeout > 0 {
		c.Options.PoolTimeout = poolTimeout
	}
	// Default is to not close aged connections.
	maxConnAge := c.Config.GetDuration(fmt.Sprintf("%s.maxConnAge", prefix))
	if maxConnAge > 0 {
		c.Options.MaxConnAge = maxConnAge
	}

	minIdleConns := c.Config.GetInt(fmt.Sprintf("%s.minIdleConns", prefix))
	if minIdleConns > 0 {
		c.Options.MinIdleConns = minIdleConns
	}

	retryStatementTimeout := c.Config.GetBool(fmt.Sprintf("%s.retryStatementTimeout", prefix))
	if retryStatementTimeout {
		c.Options.RetryStatementTimeout = retryStatementTimeout
	}

	// Default is 250 milliseconds; -1 disables backoff.
	minRetryBackoff := c.Config.GetDuration(fmt.Sprintf("%s.minRetryBackoff", prefix))
	if minRetryBackoff > 0 || minRetryBackoff == -1 {
		c.Options.MinRetryBackoff = minRetryBackoff
	}

	// Default is 4 seconds; -1 disables backoff.
	maxRetryBackoff := c.Config.GetDuration(fmt.Sprintf("%s.maxRetryBackoff", prefix))
	if maxRetryBackoff > 0 || maxRetryBackoff == -1 {
		c.Options.MaxRetryBackoff = maxRetryBackoff
	}

	if db == nil {
		pgDB := pg.Connect(c.Options)
		c.DB = &DB{inner: &pgDBWrapper{db: pgDB}}
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

// Cleanup closes PG connection
func (c *Client) Cleanup() error {
	err := c.Close()
	return err
}

// WithContext calls CtxWrapper WithContext if available or the DB's WithContext method otherwise
func (c *Client) WithContext(ctx context.Context) interfaces.DB {
	if c.CtxWrapper != nil {
		return c.CtxWrapper.WithContext(ctx, c.DB)
	}
	return WithContext(ctx, c.DB)
}

// Begin calls TxWrapper DbBegin if available or the DB's Begin method otherwise
func (c *Client) Begin(dbOrNil ...interfaces.DB) (interfaces.DB, error) {
	db := c.DB
	if len(dbOrNil) == 1 && dbOrNil[0] != nil {
		db = dbOrNil[0]
	}
	if c.TxWrapper != nil {
		return c.TxWrapper.DbBegin(db)
	}
	return Begin(db)
}
